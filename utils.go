package awsping

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	// Version describes application version
	Version   = "2.0.0"
	github    = "https://github.com/ekalinin/awsping"
	useragent = fmt.Sprintf("AwsPing/%s (+%s)", Version, github)
)

const (
	// ShowOnlyRegions describes a type of output when only region's name and code printed out
	ShowOnlyRegions = -1
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// Duration2ms converts time.Duration to ms (float64)
func Duration2ms(d time.Duration) float64 {
	return float64(d.Nanoseconds()) / 1000 / 1000
}

// mkRandomString returns random string
func mkRandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// LatencyOutput prints data into console
type LatencyOutput struct {
	Level   int
	Repeats int
	w       io.Writer
}

// NewOutput creates a new LatencyOutput instance
func NewOutput(level, repeats int) *LatencyOutput {
	return &LatencyOutput{
		Level:   level,
		Repeats: repeats,
		w:       os.Stdout,
	}
}

func (lo *LatencyOutput) show(regions *AWSRegions) error {
	for _, r := range *regions {
		if _, err := fmt.Fprintf(lo.w, "%-15s %-s\n", r.Code, r.Name); err != nil {
			return err
		}
	}
	return nil
}

func (lo *LatencyOutput) show0(regions *AWSRegions) error {
	for _, r := range *regions {
		if _, err := fmt.Fprintf(lo.w, "%-25s %20s\n", r.Name, r.GetLatencyStr()); err != nil {
			return err
		}
	}
	return nil
}

func (lo *LatencyOutput) show1(regions *AWSRegions) error {
	outFmt := "%5v %-15s %-30s %20s\n"
	if _, err := fmt.Fprintf(lo.w, outFmt, "", "Code", "Region", "Latency"); err != nil {
		return err
	}
	for i, r := range *regions {
		if _, err := fmt.Fprintf(lo.w, outFmt, i, r.Code, r.Name, r.GetLatencyStr()); err != nil {
			return err
		}
	}
	return nil
}

func (lo *LatencyOutput) show2(regions *AWSRegions) error {
	// format
	outFmt := "%5v %-15s %-25s"
	outFmt += strings.Repeat(" %15s", lo.Repeats) + " %15s\n"
	// header
	outStr := []interface{}{"", "Code", "Region"}
	for i := 0; i < lo.Repeats; i++ {
		outStr = append(outStr, "Try #"+strconv.Itoa(i+1))
	}
	outStr = append(outStr, "Avg Latency")

	// show header
	if _, err := fmt.Fprintf(lo.w, outFmt, outStr...); err != nil {
		return err
	}

	// each region stats
	for i, r := range *regions {
		outData := []interface{}{strconv.Itoa(i), r.Code, r.Name}
		for n := 0; n < lo.Repeats; n++ {
			switch {
			case n < len(r.Attempts) && r.Attempts[n].Err == nil:
				outData = append(outData, fmt.Sprintf("%.2f ms",
					Duration2ms(r.Attempts[n].Latency)))
			case len(r.Attempts) == 0 && n < len(r.Latencies):
				outData = append(outData, fmt.Sprintf("%.2f ms",
					Duration2ms(r.Latencies[n])))
			default:
				outData = append(outData, "-")
			}
		}
		if latency, ok := r.AverageLatency(); ok {
			outData = append(outData, fmt.Sprintf("%.2f ms", Duration2ms(latency)))
		} else {
			outData = append(outData, "-")
		}
		if _, err := fmt.Fprintf(lo.w, outFmt, outData...); err != nil {
			return err
		}
	}
	return nil
}

// Show prints data and returns any output error.
func (lo *LatencyOutput) Show(regions *AWSRegions) error {
	switch lo.Level {
	case ShowOnlyRegions:
		return lo.show(regions)
	case 0:
		return lo.show0(regions)
	case 1:
		return lo.show1(regions)
	case 2:
		return lo.show2(regions)
	default:
		return nil
	}
}

// GetRegions returns a list of regions
func GetRegions() AWSRegions {
	return AWSRegions{
		NewRegion("Africa (Cape Town)", "af-south-1"),
		NewRegion("Asia Pacific (Hong Kong)", "ap-east-1"),
		NewRegion("Asia Pacific (Taipei)", "ap-east-2"),
		NewRegion("Asia Pacific (Tokyo)", "ap-northeast-1"),
		NewRegion("Asia Pacific (Seoul)", "ap-northeast-2"),
		NewRegion("Asia Pacific (Osaka)", "ap-northeast-3"),
		NewRegion("Asia Pacific (Mumbai)", "ap-south-1"),
		NewRegion("Asia Pacific (Hyderabad)", "ap-south-2"),
		NewRegion("Asia Pacific (Singapore)", "ap-southeast-1"),
		NewRegion("Asia Pacific (Sydney)", "ap-southeast-2"),
		NewRegion("Asia Pacific (Jakarta)", "ap-southeast-3"),
		NewRegion("Asia Pacific (Melbourne)", "ap-southeast-4"),
		NewRegion("Asia Pacific (Malaysia)", "ap-southeast-5"),
		NewRegion("Asia Pacific (New Zealand)", "ap-southeast-6"),
		NewRegion("Asia Pacific (Thailand)", "ap-southeast-7"),
		NewRegion("Canada (Central)", "ca-central-1"),
		NewRegion("Canada West (Calgary)", "ca-west-1"),
		NewRegion("Europe (Frankfurt)", "eu-central-1"),
		NewRegion("Europe (Zurich)", "eu-central-2"),
		NewRegion("Europe (Stockholm)", "eu-north-1"),
		NewRegion("Europe (Milan)", "eu-south-1"),
		NewRegion("Europe (Spain)", "eu-south-2"),
		NewRegion("Europe (Ireland)", "eu-west-1"),
		NewRegion("Europe (London)", "eu-west-2"),
		NewRegion("Europe (Paris)", "eu-west-3"),
		NewRegion("Middle East (UAE)", "me-central-1"),
		NewRegion("Middle East (Bahrain)", "me-south-1"),
		NewRegion("Mexico (Central)", "mx-central-1"),
		NewRegion("South America (São Paulo)", "sa-east-1"),
		NewRegion("US East (N. Virginia)", "us-east-1"),
		NewRegion("US East (Ohio)", "us-east-2"),
		NewRegion("US West (N. California)", "us-west-1"),
		NewRegion("US West (Oregon)", "us-west-2"),
		NewRegion("Israel (Tel Aviv)", "il-central-1"),
	}
}

// CalcLatency returns list of aws regions sorted by Latency
func CalcLatency(regions AWSRegions, repeats int, useHTTP bool, useHTTPS bool, service string) {
	regions.SetService(service)
	switch {
	case useHTTP:
		regions.SetCheckType(CheckTypeHTTP)
	case useHTTPS:
		regions.SetCheckType(CheckTypeHTTPS)
	default:
		regions.SetCheckType(CheckTypeTCP)
	}
	regions.SetDefaultTarget()

	var wg sync.WaitGroup
	for n := 1; n <= repeats; n++ {
		wg.Add(len(regions))
		for i := range regions {
			go regions[i].CheckLatency(&wg)
		}
		wg.Wait()
	}

	sort.Stable(regions)
}
