package awsping

import (
	"errors"
	"net"
	"net/http"
	"testing"
)

type testTCPClient struct {
	err error
}

func (c *testTCPClient) Dial(n, a string) (net.Conn, error) {
	if c.err != nil {
		return nil, c.err
	}
	var con net.Conn
	return con, nil
}

type testHTTPClient struct {
	err error
}

func (c *testHTTPClient) Do(r *http.Request) (*http.Response, error) {
	if c.err != nil {
		return nil, c.err
	}
	return &http.Response{}, nil
}

func TestRequestDoTCPError(t *testing.T) {

	r := &AWSRequest{
		tcpClient: &testTCPClient{
			err: net.ErrWriteToConnected,
		},
	}

	l, err := r.DoTCP("net", "some-addr")
	if err == nil {
		t.Errorf("Error should not be empty")
	}
	if !errors.Is(err, net.ErrWriteToConnected) {
		t.Errorf("Want=%v, got=%v", net.ErrWriteToConnected, err)
	}
	if l != 0 {
		t.Errorf("Latency for error should be 0, but got=%d", l)
	}
}

func TestDoErr(t *testing.T) {
	errTCP := errors.New("error from tcp")
	errHTTP := errors.New("error from http")

	r := &AWSRequest{
		tcpClient: &testTCPClient{
			err: errTCP,
		},
		httpClient: &testHTTPClient{
			err: errHTTP,
		},
	}

	l, err := r.Do("ua", "addr", RequestTypeTCP)
	if err == nil {
		t.Errorf("Error should not be empty")
	}
	if !errors.Is(err, errTCP) {
		t.Errorf("Want=%v, got=%v", errTCP, err)
	}
	if l != 0 {
		t.Errorf("Latency for error should be 0, but got=%d", l)
	}

	l, err = r.Do("ua", "addr", RequestTypeHTTP)
	if err == nil {
		t.Errorf("Error should not be empty")
	}
	if !errors.Is(err, errHTTP) {
		t.Errorf("Want=%v, got=%v", errHTTP, err)
	}
	if l != 0 {
		t.Errorf("Latency for error should be 0, but got=%d", l)
	}
}
