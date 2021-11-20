package awsping

import (
	"bytes"
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

func TestOutputShowOnlyRegions(t *testing.T) {
	var b bytes.Buffer

	lo := NewOutput(ShowOnlyRegions, 0)
	lo.w = &b

	regions := GetRegions()[:2]
	lo.Show(&regions)

	got := b.String()
	want := "us-east-1       US-East (N. Virginia)\n" +
		"us-east-2       US-East (Ohio)      \n"
	if got != want {
		t.Errorf("Show: got=%q\nwant=%q", got, want)
	}
}

func TestOutputShow0(t *testing.T) {
	var b bytes.Buffer

	lo := NewOutput(0, 0)
	lo.w = &b

	regions := GetRegions()[:2]
	regions[0].Latencies = []time.Duration{15 * time.Millisecond}
	regions[1].Latencies = []time.Duration{25 * time.Millisecond}

	lo.Show(&regions)

	want := "US-East (N. Virginia)                 15.00 ms\n" +
		"US-East (Ohio)                        25.00 ms\n"
	got := b.String()
	if got != want {
		t.Errorf("Show0 failed:\ngot=%q\nwant=%q", got, want)
	}
}

func TestOutputShow1(t *testing.T) {
	var b bytes.Buffer

	lo := NewOutput(1, 0)
	lo.w = &b

	regions := GetRegions()[:2]
	regions[0].Latencies = []time.Duration{15 * time.Millisecond}
	regions[1].Latencies = []time.Duration{25 * time.Millisecond}

	lo.Show(&regions)

	got := b.String()
	want := "      Code            Region                                      Latency\n" +
		"    0 us-east-1       US-East (N. Virginia)                      15.00 ms\n" +
		"    1 us-east-2       US-East (Ohio)                             25.00 ms\n"
	if got != want {
		t.Errorf("Show1 failed:\ngot=%q\nwant=%q", got, want)
	}
}

func TestOutputShow2(t *testing.T) {
	var b bytes.Buffer

	lo := NewOutput(2, 2)
	lo.w = &b

	regions := GetRegions()[:2]
	regions[0].Latencies = []time.Duration{15 * time.Millisecond, 17 * time.Millisecond}
	regions[1].Latencies = []time.Duration{25 * time.Millisecond, 26 * time.Millisecond}

	lo.Show(&regions)

	got := b.String()
	want := "      Code            Region                             Try #1          Try #2     Avg Latency\n" +
		"    0 us-east-1       US-East (N. Virginia)            15.00 ms        17.00 ms        16.00 ms\n" +
		"    1 us-east-2       US-East (Ohio)                   25.00 ms        26.00 ms        25.50 ms\n"
	if got != want {
		t.Errorf("Show2 failed:\ngot=%q\nwant=%q", got, want)
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
