package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	version   = "0.5.2"
	github    = "https://github.com/ekalinin/awsping"
	useragent = fmt.Sprintf("AwsPing/%s (+%s)", version, github)
)

var (
	repeats = flag.Int("repeats", 1, "Number of repeats")
	useHTTP = flag.Bool("http", false, "Use http transport (default is tcp)")
	showVer = flag.Bool("v", false, "Show version")
	verbose = flag.Int("verbose", 0, "Verbosity level")
	service = flag.String("service", "dynamodb", "AWS Service: ec2, sdb, sns, sqs, ...")
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// Duration2ms converts time.Duration to ms (float64)
func Duration2ms(d time.Duration) float64 {
	return float64(d.Nanoseconds()) / 1000 / 1000
}

func mkRandoString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// AWSRegion description of the AWS EC2 region
type AWSRegion struct {
	Name      string
	Code      string
	Service   string
	Latencies []time.Duration
	Error     error
}

// CheckLatencyHTTP Test Latency via HTTP
func (r *AWSRegion) CheckLatencyHTTP(wg *sync.WaitGroup) {
	defer wg.Done()
	url := fmt.Sprintf("http://%s.%s.amazonaws.com/ping?x=%s", r.Service,
		r.Code, mkRandoString(13))
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", useragent)

	start := time.Now()
	resp, err := client.Do(req)
	r.Latencies = append(r.Latencies, time.Since(start))
	defer resp.Body.Close()

	r.Error = err
}

// CheckLatencyTCP Test Latency via TCP
func (r *AWSRegion) CheckLatencyTCP(wg *sync.WaitGroup) {
	defer wg.Done()
	tcpURI := fmt.Sprintf("%s.%s.amazonaws.com:80", r.Service, r.Code)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", tcpURI)
	if err != nil {
		r.Error = err
		return
	}
	start := time.Now()
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	r.Latencies = append(r.Latencies, time.Since(start))
	defer conn.Close()

	r.Error = err
}

// GetLatency returns Latency in ms
func (r *AWSRegion) GetLatency() float64 {
	sum := float64(0)
	for _, l := range r.Latencies {
		sum += Duration2ms(l)
	}
	return sum / float64(len(r.Latencies))
}

// AWSRegions slice of the AWSRegion
type AWSRegions []AWSRegion

func (rs AWSRegions) Len() int {
	return len(rs)
}

func (rs AWSRegions) Less(i, j int) bool {
	return rs[i].GetLatency() < rs[j].GetLatency()
}

func (rs AWSRegions) Swap(i, j int) {
	rs[i], rs[j] = rs[j], rs[i]
}

// CalcLatency returns list of aws regions sorted by Latency
func CalcLatency(repeats int, useHTTP bool, service string) *AWSRegions {
	regions := AWSRegions{
		{Service: service, Name: "US-East (Virginia)", Code: "us-east-1"},
		{Service: service, Name: "US-East (Ohio)", Code: "us-east-2"},
		{Service: service, Name: "US-West (California)", Code: "us-west-1"},
		{Service: service, Name: "US-West (Oregon)", Code: "us-west-2"},
		{Service: service, Name: "Canada (Ceentral)", Code: "ca-central-1"},
		{Service: service, Name: "Europe (Ireland)", Code: "eu-west-1"},
		{Service: service, Name: "Europe (Frankfurt)", Code: "eu-central-1"},
		{Service: service, Name: "Europe (London)", Code: "eu-west-2"},
		{Service: service, Name: "Asia Pacific (Tokyo)", Code: "ap-northeast-1"},
		{Service: service, Name: "Asia Pacific (Seoul)", Code: "ap-northeast-2"},
		{Service: service, Name: "Asia Pacific (Singapore)", Code: "ap-southeast-1"},
		{Service: service, Name: "Asia Pacific (Mumbai)", Code: "ap-south-1"},
		{Service: service, Name: "Asia Pacific (Sydney)", Code: "ap-southeast-2"},
		{Service: service, Name: "South America (São Paulo)", Code: "sa-east-1"},
		//{Name: "China (Beijing)", Code: "cn-north-1"},
	}
	var wg sync.WaitGroup

	for n := 1; n <= repeats; n++ {
		wg.Add(len(regions))
		for i := range regions {
			if useHTTP {
				go regions[i].CheckLatencyHTTP(&wg)
			} else {
				go regions[i].CheckLatencyTCP(&wg)
			}
		}
		wg.Wait()
	}

	sort.Sort(regions)
	return &regions
}

// LatencyOutput prints data into console
type LatencyOutput struct {
	Level int
}

func (lo *LatencyOutput) show0(regions *AWSRegions) {
	for _, r := range *regions {
		fmt.Printf("%-25s %20s\n", r.Name,
			fmt.Sprintf("%.2f ms", r.GetLatency()))
	}
}

func (lo *LatencyOutput) show1(regions *AWSRegions) {
	outFmt := "%5v %-15s %-30s %20s\n"
	fmt.Printf(outFmt, "", "Code", "Region", "Latency")
	for i, r := range *regions {
		ms := fmt.Sprintf("%.2f ms", r.GetLatency())
		fmt.Printf(outFmt, i, r.Code, r.Name, ms)
	}
}

func (lo *LatencyOutput) show2(regions *AWSRegions) {
	// format
	outFmt := "%5v %-15s %-25s"
	outFmt += strings.Repeat(" %15s", *repeats) + " %15s\n"
	// header
	outStr := []interface{}{"", "Code", "Region"}
	for i := 0; i < *repeats; i++ {
		outStr = append(outStr, "Try #"+strconv.Itoa(i+1))
	}
	outStr = append(outStr, "Avg Latency")

	// show header
	fmt.Printf(outFmt, outStr...)

	// each region stats
	for i, r := range *regions {
		outData := []interface{}{strconv.Itoa(i), r.Code, r.Name}
		for n := 0; n < *repeats; n++ {
			outData = append(outData, fmt.Sprintf("%.2f ms",
				Duration2ms(r.Latencies[n])))
		}
		outData = append(outData, fmt.Sprintf("%.2f ms", r.GetLatency()))
		fmt.Printf(outFmt, outData...)
	}
}

// Show print data
func (lo *LatencyOutput) Show(regions *AWSRegions) {
	switch lo.Level {
	case 0:
		lo.show0(regions)
	case 1:
		lo.show1(regions)
	case 2:
		lo.show2(regions)
	}
}

func main() {

	flag.Parse()

	if *showVer {
		fmt.Println(version)
		os.Exit(0)
	}

	regions := CalcLatency(*repeats, *useHTTP, *service)
	lo := LatencyOutput{*verbose}
	lo.Show(regions)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
