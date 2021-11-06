package awsping

import (
	"fmt"
	"net"
	"net/http"
	"sort"
	"sync"
	"time"
)

// AWSRegion description of the AWS EC2 region
type AWSRegion struct {
	Name      string
	Code      string
	Service   string
	Latencies []time.Duration
	Error     error
}

// AWSTarget describes aws region network details (host, ip)
type AWSTarget struct {
	Hostname string
	IPAddr   *net.IPAddr
}

// CheckLatencyHTTP Test Latency via HTTP
func (r *AWSRegion) CheckLatencyHTTP(wg *sync.WaitGroup, https bool) {
	defer wg.Done()
	url := fmt.Sprintf("http://%s/ping?x=%s", r.GetTarget().Hostname, mkRandoString(13))

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		r.Error = err
	}
	req.Header.Set("User-Agent", useragent)

	start := time.Now()
	resp, err := client.Do(req)
	r.Latencies = append(r.Latencies, time.Since(start))
	defer resp.Body.Close()

	r.Error = err
}

func (r *AWSRegion) GetTarget() AWSTarget {

	hostname := fmt.Sprintf("%s.%s.amazonaws.com", r.Service, r.Code)
	ipAddr, _ := net.ResolveIPAddr("ip4", hostname)

	return AWSTarget{
		Hostname: hostname,
		IPAddr:   ipAddr,
	}

}

// CheckLatencyTCP Test Latency via TCP
func (r *AWSRegion) CheckLatencyTCP(wg *sync.WaitGroup) {
	defer wg.Done()

	tcpAddr :=
		net.TCPAddr{
			IP:   r.GetTarget().IPAddr.IP,
			Port: 80}

	start := time.Now()
	conn, err := net.DialTCP("tcp", nil, &tcpAddr)
	if err != nil {
		r.Error = err
		return
	}
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

// GetLatencyStr returns Latency in string
func (r *AWSRegion) GetLatencyStr() string {
	if r.Error != nil {
		return r.Error.Error()
	}
	return fmt.Sprintf("%.2f ms", r.GetLatency())
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

// GetRegions returns a list of regions
func GetRegions(service string) AWSRegions {
	return AWSRegions{
		{Service: service, Name: "US-East (N. Virginia)", Code: "us-east-1"},
		{Service: service, Name: "US-East (Ohio)", Code: "us-east-2"},
		{Service: service, Name: "US-West (N. California)", Code: "us-west-1"},
		{Service: service, Name: "US-West (Oregon)", Code: "us-west-2"},
		{Service: service, Name: "Canada (Central)", Code: "ca-central-1"},
		{Service: service, Name: "Europe (Ireland)", Code: "eu-west-1"},
		{Service: service, Name: "Europe (Frankfurt)", Code: "eu-central-1"},
		{Service: service, Name: "Europe (London)", Code: "eu-west-2"},
		{Service: service, Name: "Europe (Milan)", Code: "eu-south-1"},
		{Service: service, Name: "Europe (Paris)", Code: "eu-west-3"},
		{Service: service, Name: "Europe (Stockholm)", Code: "eu-north-1"},
		{Service: service, Name: "Africa (Cape Town)", Code: "af-south-1"},
		{Service: service, Name: "Asia Pacific (Osaka)", Code: "ap-northeast-3"},
		{Service: service, Name: "Asia Pacific (Hong Kong)", Code: "ap-east-1"},
		{Service: service, Name: "Asia Pacific (Tokyo)", Code: "ap-northeast-1"},
		{Service: service, Name: "Asia Pacific (Seoul)", Code: "ap-northeast-2"},
		{Service: service, Name: "Asia Pacific (Singapore)", Code: "ap-southeast-1"},
		{Service: service, Name: "Asia Pacific (Mumbai)", Code: "ap-south-1"},
		{Service: service, Name: "Asia Pacific (Sydney)", Code: "ap-southeast-2"},
		{Service: service, Name: "South America (São Paulo)", Code: "sa-east-1"},
		{Service: service, Name: "Middle East (Bahrain)", Code: "me-south-1"},
	}
}

// CalcLatency returns list of aws regions sorted by Latency
func CalcLatency(repeats int, useHTTP bool, useHTTPS bool, service string) *AWSRegions {
	regions := GetRegions(service)
	var wg sync.WaitGroup

	for n := 1; n <= repeats; n++ {
		wg.Add(len(regions))
		for i := range regions {
			if useHTTP || useHTTPS {
				go regions[i].CheckLatencyHTTP(&wg, useHTTPS)
			} else {
				go regions[i].CheckLatencyTCP(&wg)
			}
		}
		wg.Wait()
	}

	sort.Sort(regions)
	return &regions
}
