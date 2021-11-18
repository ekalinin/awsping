package awsping

import (
	"net"
	"net/http"
	"time"
)

type RequestType int

const (
	RequestTypeHTTP RequestType = iota
	RequestTypeTCP
)

type Requester interface {
	Do(ua, url string, reqType RequestType) (time.Duration, error)
}

type AWSRequest struct{}

func (r *AWSRequest) DoHTTP(ua, url string) (time.Duration, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", ua)

	start := time.Now()
	resp, err := client.Do(req)
	latency := time.Since(start)

	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return latency, nil
}

func (r *AWSRequest) DoTCP(ua, addr string) (time.Duration, error) {
	d := net.Dialer{}

	start := time.Now()
	conn, err := d.Dial("tcp", addr)
	if err != nil {
		return 0, err
	}
	l := time.Since(start)
	defer conn.Close()

	return l, nil
}

func (r *AWSRequest) Do(ua, url string, reqType RequestType) (time.Duration, error) {
	if reqType == RequestTypeHTTP {
		return r.DoHTTP(ua, url)
	}
	return r.DoTCP(ua, url)
}
