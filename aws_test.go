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
		fmt.Fprintln(w, "X")
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

	if got < want || got > want+2 {
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
		t.Errorf("failed:\ngot=%f\nwant=%f\nregion=%q", got, want, regions[0])
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

// ---------------------------------------------

func TestAWSRegionsLen(t *testing.T) {
	regions := GetRegions()

	got := regions.Len()
	want := len(regions)

	if got != want {
		t.Errorf("failed:\ngot=%q\nwant=%q", got, want)
	}
}

func TestAWSRegionsLess(t *testing.T) {
	regions := GetRegions()

	regions[0].Latencies = []time.Duration{15 * time.Millisecond}
	regions[1].Latencies = []time.Duration{25 * time.Millisecond}

	if !regions.Less(0, 1) {
		t.Errorf("failed: not less, regions=%q", regions)
	}
}

func TestAWSRegionsSwap(t *testing.T) {
	regions := GetRegions()

	regions[0].Latencies = []time.Duration{15 * time.Millisecond}
	regions[1].Latencies = []time.Duration{25 * time.Millisecond}

	regions.Swap(0, 3)

	if len(regions[0].Latencies) != 0 {
		t.Errorf("failed: not swapped, regions=%q", regions)
	}
}

func TestAWSRegionsSetService(t *testing.T) {
	regions := GetRegions()
	service := "ec2"

	regions.SetService(service)

	if regions[0].Service != service || regions[len(regions)-1].Service != service {
		t.Errorf("failed: not setted, regions=%q, service=%s", regions, service)
	}
}

func TestAWSRegionsSetCheckType(t *testing.T) {
	regions := GetRegions()
	checkType := CheckTypeHTTP

	regions.SetCheckType(checkType)

	if regions[0].CheckType != checkType || regions[len(regions)-1].CheckType != checkType {
		t.Errorf("failed: not setted, regions=%q, checkType=%d", regions, checkType)
	}
}

func TestAWSRegionsSetDefaulTarget(t *testing.T) {
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
