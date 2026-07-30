package awsping

import (
	"bytes"
	"errors"
	"testing"
	"time"
)

func TestDuration(t *testing.T) {
	input := time.Duration(1 * time.Second)
	want := 1000.0
	got := Duration2ms(input)
	if got != want {
		t.Errorf("Duration was incorrect, got: %f, want: %f.", got, want)
	}
}

func TestRandomString(t *testing.T) {
	tests := []struct {
		n   int
		res string
	}{
		{1, ""},
		{5, ""},
		{10, ""},
	}

	for idx, test := range tests {
		test.res = mkRandomString(test.n)
		if len(test.res) != test.n {
			t.Errorf("Try %d: n=%d, got: %s (len=%d), want: %d.",
				idx, test.n, test.res, len(test.res), test.n)
		}
	}
}

func TestGetRegionsIncludesLatestRegions(t *testing.T) {
	want := map[string]string{
		"ap-east-2":    "Asia Pacific (Taipei)",
		"ca-west-1":    "Canada West (Calgary)",
		"mx-central-1": "Mexico (Central)",
		"sa-east-1":    "South America (São Paulo)",
	}

	for _, region := range GetRegions() {
		name, ok := want[region.Code]
		if !ok {
			continue
		}
		if region.Name != name {
			t.Errorf("region %s: got name %q, want %q", region.Code, region.Name, name)
		}
		delete(want, region.Code)
	}

	for code := range want {
		t.Errorf("missing region %s", code)
	}
}

func TestOutputShowOnlyRegions(t *testing.T) {
	var b bytes.Buffer

	lo := NewOutput(ShowOnlyRegions, 0)
	lo.w = &b

	regions := GetRegions()[:2]
	if err := lo.Show(&regions); err != nil {
		t.Fatalf("Show error: %v", err)
	}

	got := b.String()
	want := "af-south-1      Africa (Cape Town)\n" +
		"ap-east-1       Asia Pacific (Hong Kong)\n"

	if got != want {
		t.Errorf("Show:\ngot =%q\nwant=%q", got, want)
	}
}

func TestOutputShow0(t *testing.T) {
	var b bytes.Buffer

	lo := NewOutput(0, 0)
	lo.w = &b

	regions := GetRegions()[:2]
	regions[0].Latencies = []time.Duration{15 * time.Millisecond}
	regions[1].Latencies = []time.Duration{25 * time.Millisecond}

	if err := lo.Show(&regions); err != nil {
		t.Fatalf("Show error: %v", err)
	}

	want := "Africa (Cape Town)                    15.00 ms\n" +
		"Asia Pacific (Hong Kong)              25.00 ms\n"
	got := b.String()
	if got != want {
		t.Errorf("Show0 failed:\ngot =%q\nwant=%q", got, want)
	}
}

func TestOutputShow1(t *testing.T) {
	var b bytes.Buffer

	lo := NewOutput(1, 0)
	lo.w = &b

	regions := GetRegions()[:2]
	regions[0].Latencies = []time.Duration{15 * time.Millisecond}
	regions[1].Latencies = []time.Duration{25 * time.Millisecond}

	if err := lo.Show(&regions); err != nil {
		t.Fatalf("Show error: %v", err)
	}

	got := b.String()
	want := "      Code            Region                                      Latency\n" +
		"    0 af-south-1      Africa (Cape Town)                         15.00 ms\n" +
		"    1 ap-east-1       Asia Pacific (Hong Kong)                   25.00 ms\n"
	if got != want {
		t.Errorf("Show1 failed:\ngot =%q\nwant=%q", got, want)
	}
}

func TestOutputShow2(t *testing.T) {
	var b bytes.Buffer

	lo := NewOutput(2, 2)
	lo.w = &b

	regions := GetRegions()[:2]
	regions[0].Latencies = []time.Duration{15 * time.Millisecond, 17 * time.Millisecond}
	regions[1].Latencies = []time.Duration{25 * time.Millisecond, 26 * time.Millisecond}

	if err := lo.Show(&regions); err != nil {
		t.Fatalf("Show error: %v", err)
	}

	got := b.String()
	want := "      Code            Region                             Try #1          Try #2     Avg Latency\n" +
		"    0 af-south-1      Africa (Cape Town)               15.00 ms        17.00 ms        16.00 ms\n" +
		"    1 ap-east-1       Asia Pacific (Hong Kong)         25.00 ms        26.00 ms        25.50 ms\n"
	if got != want {
		t.Errorf("Show2 failed:\ngot =%q\nwant=%q", got, want)
	}
}

func TestOutputShow2PreservesFailedAttemptPositions(t *testing.T) {
	var b bytes.Buffer

	lo := NewOutput(2, 3)
	lo.w = &b

	regions := GetRegions()[:2]
	regions[0].Attempts = []AttemptResult{
		{Err: errors.New("timeout")},
		{Latency: 17 * time.Millisecond},
		{Err: errors.New("connection refused")},
	}
	regions[1].Attempts = []AttemptResult{
		{Err: errors.New("timeout")},
		{Err: errors.New("timeout")},
		{Err: errors.New("timeout")},
	}

	if err := lo.Show(&regions); err != nil {
		t.Fatalf("Show error: %v", err)
	}

	got := b.String()
	want := "      Code            Region                             Try #1          Try #2          Try #3     Avg Latency\n" +
		"    0 af-south-1      Africa (Cape Town)                      -        17.00 ms               -        17.00 ms\n" +
		"    1 ap-east-1       Asia Pacific (Hong Kong)                -               -               -               -\n"
	if got != want {
		t.Errorf("Show2 failed:\ngot =%q\nwant=%q", got, want)
	}
}

type errorWriter struct {
	err error
}

func (w errorWriter) Write([]byte) (int, error) {
	return 0, w.err
}

func TestOutputShowError(t *testing.T) {
	want := errors.New("write failed")
	lo := NewOutput(0, 0)
	lo.w = errorWriter{err: want}
	regions := GetRegions()[:1]

	err := lo.Show(&regions)
	if !errors.Is(err, want) {
		t.Errorf("Show error: got %v, want %v", err, want)
	}
}

func TestCalcLatency(t *testing.T) {

	regions := GetRegions()[:3]
	regions[0].Request = &testRequest{duration: 30 * time.Millisecond}
	regions[1].Request = &testRequest{duration: 7 * time.Millisecond}
	regions[2].Request = &testRequest{duration: 15 * time.Millisecond}

	regionsStats := make(AWSRegions, regions.Len())

	checkSort := func(origIndex, sortedIdx int) {
		got := regionsStats[sortedIdx].Name
		want := regions[origIndex].Name

		if got != want {
			t.Errorf("CalcLatency failed:\ngot=%q\nwant=%q\norig=%d\nsorted=%d",
				got, want, origIndex, sortedIdx)
		}
	}

	for i := 1; i < 4; i++ {
		copy(regionsStats, regions)

		switch i {
		case 1:
			CalcLatency(regionsStats, 1, false, false, "ec2")
		case 2:
			CalcLatency(regionsStats, 1, true, false, "ec2")
		default:
			CalcLatency(regionsStats, 1, true, true, "ec2")
		}

		checkSort(0, 2)
		checkSort(1, 0)
		checkSort(2, 1)
	}
}

func TestCalcLatencySortsFailedRegionsLast(t *testing.T) {
	regions := GetRegions()[:3]
	regions[0].Request = &testRequest{err: errors.New("first failed")}
	regions[1].Request = &testRequest{duration: 7 * time.Millisecond}
	regions[2].Request = &testRequest{err: errors.New("third failed")}

	CalcLatency(regions, 1, false, false, "ec2")

	if got, want := regions[0].Name, "Asia Pacific (Hong Kong)"; got != want {
		t.Errorf("first region: got %q, want %q", got, want)
	}
	if got, want := regions[1].Name, "Africa (Cape Town)"; got != want {
		t.Errorf("second region: got %q, want %q", got, want)
	}
	if got, want := regions[2].Name, "Asia Pacific (Taipei)"; got != want {
		t.Errorf("third region: got %q, want %q", got, want)
	}
}
