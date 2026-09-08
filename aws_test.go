package awsping

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestAWSRegionError(t *testing.T) {
	AWSErr := errors.New("something bad")
	r := AWSRegion{Error: AWSErr}

	got := r.GetLatencyStr()
	want := AWSErr.Error()

	if got != want {
		t.Errorf("failed:\ngot=%q\nwant=%q", got, want)
	}
}

type testTarget struct {
	URL string
	IP  *net.TCPAddr
}

func (r *testTarget) GetURL() string {
	return r.URL
}

// GetIP return IP for AWS target
func (r *testTarget) GetIP() (*net.TCPAddr, error) {
	return r.IP, nil
}

func TestAWSRegionCheckLatencyHTTP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(15 * time.Millisecond)
		_, _ = fmt.Fprintln(w, "X")
	}))
	defer ts.Close()

	tt := testTarget{URL: ts.URL}

	regions := GetRegions()
	service := "ec2"
	checkType := CheckTypeHTTP

	regions.SetService(service)
	regions.SetCheckType(checkType)
	regions.SetTarget(func(r *AWSRegion) {
		r.Target = &tt
	})

	var wg sync.WaitGroup
	wg.Add(1)
	regions[0].CheckLatency(&wg)

	got := regions[0].GetLatency()
	want := 15.0

	if got < want || got > want*2 {
		t.Errorf("failed:\ngot=%f\nwant=%f", got, want)
	}

	// check "error"
	errTxt := "something bad"
	regions[0].Request = &testRequest{err: errors.New(errTxt)}

	wg.Add(1)
	regions[0].CheckLatency(&wg)

	if regions[0].Error == nil {
		t.Errorf("failed: error should not be empty")
	}

	if regions[0].Error.Error() != errTxt {
		t.Errorf("failed: error should be empty=%s", errTxt)
	}
}

type testRequest struct {
	duration time.Duration
	err      error
}

func (d *testRequest) Do(_, _ string, _ RequestType) (time.Duration, error) {
	if d.err != nil {
		return 0, d.err
	}
	return d.duration, nil
}

type testRequestResult struct {
	duration time.Duration
	err      error
}

type sequenceRequest struct {
	results []testRequestResult
	next    int
}

func (r *sequenceRequest) Do(_, _ string, _ RequestType) (time.Duration, error) {
	result := r.results[r.next]
	r.next++
	return result.duration, result.err
}

func TestAWSRegionCheckLatencyTCP(t *testing.T) {
	// just random local IP
	tt := testTarget{IP: &net.TCPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 67890,
	}}

	regions := GetRegions()
	service := "ec2"
	checkType := CheckTypeTCP

	regions.SetService(service)
	regions.SetCheckType(checkType)
	regions.SetTarget(func(r *AWSRegion) {
		r.Target = &tt
	})
	regions[0].Request = &testRequest{duration: 15 * time.Millisecond}

	var wg sync.WaitGroup
	wg.Add(1)
	regions[0].CheckLatency(&wg)

	got := regions[0].GetLatency()
	want := 15.0

	if got < want || got > want+1 {
		t.Errorf("failed:\ngot=%f\nwant=%f\nregion=%v", got, want, regions[0])
	}

	if regions[0].Error != nil {
		t.Errorf("failed: error should be empty")
	}

	// check "error"
	errTxt := "something bad"
	regions[0].Request = &testRequest{err: errors.New(errTxt)}

	wg.Add(1)
	regions[0].CheckLatency(&wg)

	if regions[0].Error == nil {
		t.Errorf("failed: error should not be empty")
	}

	if regions[0].Error.Error() != errTxt {
		t.Errorf("failed: error should be empty=%s", errTxt)
	}
}

func TestAWSRegionAttemptsPreserveFailures(t *testing.T) {
	errFirst := errors.New("first attempt failed")
	region := NewRegion("Test", "test-1")
	region.Target = &testTarget{IP: &net.TCPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 67890,
	}}
	region.Request = &sequenceRequest{results: []testRequestResult{
		{err: errFirst},
		{duration: 15 * time.Millisecond},
		{duration: 25 * time.Millisecond},
	}}

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		region.CheckLatency(&wg)
	}

	if got, want := len(region.Attempts), 3; got != want {
		t.Fatalf("attempt count: got %d, want %d", got, want)
	}
	if !errors.Is(region.Attempts[0].Err, errFirst) {
		t.Errorf("first attempt error: got %v, want %v", region.Attempts[0].Err, errFirst)
	}
	if region.Attempts[1].Err != nil {
		t.Errorf("second attempt error: got %v, want nil", region.Attempts[1].Err)
	}
	if got, want := region.Attempts[1].Latency, 15*time.Millisecond; got != want {
		t.Errorf("second attempt latency: got %v, want %v", got, want)
	}
	if got, want := region.SuccessfulAttempts(), 2; got != want {
		t.Errorf("successful attempts: got %d, want %d", got, want)
	}
	if got, want := region.FailedAttempts(), 1; got != want {
		t.Errorf("failed attempts: got %d, want %d", got, want)
	}
	if got, want := region.GetLatencyStr(), "20.00 ms"; got != want {
		t.Errorf("latency string: got %q, want %q", got, want)
	}
}

func TestAWSRegionAllAttemptsFailed(t *testing.T) {
	errFirst := errors.New("first attempt failed")
	errLast := errors.New("last attempt failed")
	region := AWSRegion{Attempts: []AttemptResult{
		{Err: errFirst},
		{Err: errLast},
	}}

	if latency, ok := region.AverageLatency(); ok || latency != 0 {
		t.Errorf("average latency: got (%v, %t), want (0, false)", latency, ok)
	}
	if got := region.GetLatency(); got != 0 {
		t.Errorf("latency: got %f, want 0", got)
	}
	if got, want := region.GetLatencyStr(), errLast.Error(); got != want {
		t.Errorf("latency string: got %q, want %q", got, want)
	}
	if !errors.Is(region.LastError(), errLast) {
		t.Errorf("last error: got %v, want %v", region.LastError(), errLast)
	}
}

func TestAWSRegionWithoutAttempts(t *testing.T) {
	region := AWSRegion{}

	if latency, ok := region.AverageLatency(); ok || latency != 0 {
		t.Errorf("average latency: got (%v, %t), want (0, false)", latency, ok)
	}
	if got := region.GetLatency(); got != 0 {
		t.Errorf("latency: got %f, want 0", got)
	}
	if got, want := region.GetLatencyStr(), "-"; got != want {
		t.Errorf("latency string: got %q, want %q", got, want)
	}
}

// ---------------------------------------------

func TestAWSRegionsLen(t *testing.T) {
	regions := GetRegions()

	got := regions.Len()
	want := len(regions)

	if got != want {
		t.Errorf("failed:\ngot=%d\nwant=%d", got, want)
	}
}

func TestAWSRegionsLess(t *testing.T) {
	regions := GetRegions()

	regions[0].Latencies = []time.Duration{15 * time.Millisecond}
	regions[1].Latencies = []time.Duration{25 * time.Millisecond}

	if !regions.Less(0, 1) {
		t.Errorf("failed: not less, regions=%v", regions)
	}
}

func TestAWSRegionsLessPutsFailedRegionsLast(t *testing.T) {
	regions := AWSRegions{
		{Name: "Failed", Attempts: []AttemptResult{{Err: errors.New("failed")}}},
		{Name: "Successful", Attempts: []AttemptResult{{Latency: 25 * time.Millisecond}}},
	}

	if regions.Less(0, 1) {
		t.Error("failed region should not sort before a successful region")
	}
	if !regions.Less(1, 0) {
		t.Error("successful region should sort before a failed region")
	}
}

func TestAWSRegionsSwap(t *testing.T) {
	regions := GetRegions()

	regions[0].Latencies = []time.Duration{15 * time.Millisecond}
	regions[1].Latencies = []time.Duration{25 * time.Millisecond}

	regions.Swap(0, 3)

	if len(regions[0].Latencies) != 0 {
		t.Errorf("failed: not swapped, regions=%v", regions)
	}
}

func TestAWSRegionsSetService(t *testing.T) {
	regions := GetRegions()
	service := "ec2"

	regions.SetService(service)

	if regions[0].Service != service || regions[len(regions)-1].Service != service {
		t.Errorf("failed: not set, regions=%v, service=%s", regions, service)
	}
}

func TestAWSRegionsSetCheckType(t *testing.T) {
	regions := GetRegions()
	checkType := CheckTypeHTTP

	regions.SetCheckType(checkType)

	if regions[0].CheckType != checkType || regions[len(regions)-1].CheckType != checkType {
		t.Errorf("failed: not set, regions=%v, checkType=%d", regions, checkType)
	}
}

func TestAWSRegionsSetDefaultTarget(t *testing.T) {
	regions := GetRegions()
	service := "ec2"
	checkType := CheckTypeHTTPS

	regions.SetService(service)
	regions.SetCheckType(checkType)
	regions.SetDefaultTarget()

	got := regions[0].Target.GetURL()
	want := fmt.Sprintf("https://ec2.%s.amazonaws.com/ping?x=", regions[0].Code)

	if !strings.HasPrefix(got, want) {
		t.Errorf("failed: wrong url\ngot=%s\nneed=%s", got, want)
	}
}
