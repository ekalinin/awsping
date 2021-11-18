package awsping

import (
	"fmt"
	"net"
)

// Targetter is an interface to get target's IP or URL
type Targetter interface {
	GetURL() string
	GetIP() (*net.TCPAddr, error)
}

// AWSTarget implements Targetter for AWS
type AWSTarget struct {
	HTTPS   bool
	Code    string
	Service string
	Rnd     string
}

// GetURL return URL for AWS target
func (r *AWSTarget) GetURL() string {
	proto := "http"
	if r.HTTPS {
		proto = "https"
	}
	hostname := fmt.Sprintf("%s.%s.amazonaws.com", r.Service, r.Code)
	url := fmt.Sprintf("%s://%s/ping?x=%s", proto, hostname, r.Rnd)
	return url
}

// GetIP return IP for AWS target
func (r *AWSTarget) GetIP() (*net.TCPAddr, error) {
	tcpURI := fmt.Sprintf("%s.%s.amazonaws.com:80", r.Service, r.Code)
	return net.ResolveTCPAddr("tcp4", tcpURI)
}
