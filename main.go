package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"
)

var (
	version   = "0.3.0"
	github    = "https://github.com/ekalinin/awsping"
	useragent = fmt.Sprintf("AwsPing/%s (+%s)", version, github)
)

var (
	repeats = flag.Int("repeats", 1, "Number of repeats")
	useHTTP = flag.Bool("http", false, "Use http transport (default is tcp)")
	showVer = flag.Bool("v", false, "Show version")
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

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
	Latencies []time.Duration
	Error     error
}

// CheckLatencyHTTP Test Latency via HTTP
func (r *AWSRegion) CheckLatencyHTTP(wg *sync.WaitGroup) {
	defer wg.Done()
	url := fmt.Sprintf("http://dynamodb.%s.amazonaws.com/ping?x=%s",
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
	tcpAddr, err := net.ResolveTCPAddr("tcp4",
		fmt.Sprintf("dynamodb.%s.amazonaws.com:80", r.Code))

	start := time.Now()
	conn, _ := net.DialTCP("tcp", nil, tcpAddr)
	r.Latencies = append(r.Latencies, time.Since(start))
	defer conn.Close()
	_, _ = conn.Write([]byte("HEAD / HTTP/1.0\r\n\r\n"))
	_, _ = ioutil.ReadAll(conn)

	r.Error = err
}

// GetLatency returns Latency in ms
func (r *AWSRegion) GetLatency() float64 {
	sum := float64(0)
	for _, l := range r.Latencies {
		sum += float64(l.Nanoseconds()) / 1000 / 1000
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
func CalcLatency(repeats int, useHTTP bool) *AWSRegions {
	regions := AWSRegions{
		{Name: "US-East (Virginia)", Code: "us-east-1"},
		{Name: "US-West (California)", Code: "us-west-1"},
		{Name: "US-West (Oregon)", Code: "us-west-2"},
		{Name: "Asia Pacific (Mumbai)", Code: "ap-south-1"},
		{Name: "Asia Pacific (Seoul)", Code: "ap-northeast-2"},
		{Name: "Asia Pacific (Singapore)", Code: "ap-southeast-1"},
		{Name: "Asia Pacific (Sydney)", Code: "ap-southeast-2"},
		{Name: "Asia Pacific (Tokyo)", Code: "ap-northeast-1"},
		{Name: "Europe (Ireland)", Code: "eu-west-1"},
		{Name: "Europe (Frankfurt)", Code: "eu-central-1"},
		{Name: "South America (São Paulo)", Code: "sa-east-1"},
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

func main() {

	flag.Parse()

	if *showVer {
		fmt.Println(version)
		os.Exit(0)
	}

	regions := *CalcLatency(*repeats, *useHTTP)

	outFmt := "%5v %-30s %20s\n"
	fmt.Printf(outFmt, "", "Region", "Latency")
	for i, r := range regions {
		ms := fmt.Sprintf("%.2f ms", r.GetLatency())
		fmt.Printf(outFmt, i, r.Name, ms)
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
