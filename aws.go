package awsping

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

type CheckType int

const (
	TCPCheck CheckType = iota
	HTTPCheck
	HTTPSCheck
)

// --------------------------------------------

// AWSRegion description of the AWS EC2 region
type AWSRegion struct {
	Name      string
	Code      string
	Service   string
	Latencies []time.Duration
	Error     error
	Type      CheckType

	Target  Targetter
	Dialler TargetDialler
}

func NewRegion(name, code string) AWSRegion {
	return AWSRegion{
		Name:    name,
		Code:    code,
		Type:    TCPCheck,
		Dialler: &net.Dialer{},
	}
}

// CheckLatency
func (r *AWSRegion) CheckLatency(wg *sync.WaitGroup) {
	defer wg.Done()

	if r.Type == HTTPCheck || r.Type == HTTPSCheck {
		r.checkLatencyHTTP(r.Type == HTTPSCheck)
	} else {
		r.checkLatencyTCP()
	}
}

// checkLatencyHTTP Test Latency via HTTP
func (r *AWSRegion) checkLatencyHTTP(https bool) {
	url := r.Target.GetURL()
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		r.Error = err
		return
	}
	req.Header.Set("User-Agent", useragent)

	start := time.Now()
	resp, err := client.Do(req)
	r.Latencies = append(r.Latencies, time.Since(start))
	if err != nil {
		r.Error = err
		return
	}
	defer resp.Body.Close()
}

// checkLatencyTCP Test Latency via TCP
func (r *AWSRegion) checkLatencyTCP() {
	tcpAddr, err := r.Target.GetIP()
	if err != nil {
		r.Error = err
		return
	}

	start := time.Now()
	conn, err := r.Dialler.Dial("tcp", tcpAddr.String())
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

// --------------------------------------------

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

// SetService sets service for all regions
func (rs AWSRegions) SetService(service string) {
	for i := range rs {
		rs[i].Service = service
	}
}

// SetCheckType sets Check Type for all regions
func (rs AWSRegions) SetCheckType(checkType CheckType) {
	for i := range rs {
		rs[i].Type = checkType
	}
}

// SetDefaultTarget sets default target instance
func (rs AWSRegions) SetDefaultTarget() {
	rs.SetTarget(func(r *AWSRegion) {
		r.Target = &AWSTarget{
			HTTPS:   r.Type == HTTPSCheck,
			Code:    r.Code,
			Service: r.Service,
			Rnd:     mkRandomString(13),
		}
	})
}

// SetDefaultTarget sets default target instance for all regions
func (rs AWSRegions) SetTarget(fn func(r *AWSRegion)) {
	for i := range rs {
		fn(&rs[i])
	}
}
