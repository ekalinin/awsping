package awsping

import (
	"fmt"
	"sync"
	"time"
)

// CheckType describes a type for a check
type CheckType int

const (
	// CheckTypeTCP is TCP type of check
	CheckTypeTCP CheckType = iota
	// CheckTypeHTTP is HTTP type of check
	CheckTypeHTTP
	// CheckTypeHTTPS is HTTPS type of check
	CheckTypeHTTPS
)

// AttemptResult describes the outcome of a single latency check.
type AttemptResult struct {
	Latency time.Duration
	Err     error
}

// --------------------------------------------

// AWSRegion description of the AWS EC2 region
type AWSRegion struct {
	Name     string
	Code     string
	Service  string
	Attempts []AttemptResult
	// Latencies is kept for backward compatibility. New code should use Attempts.
	Latencies []time.Duration
	// Error is kept for backward compatibility. New code should use LastError.
	Error     error
	CheckType CheckType

	Target  Targetter
	Request Requester
}

// NewRegion creates a new region with a name and code
func NewRegion(name, code string) AWSRegion {
	return AWSRegion{
		Name:      name,
		Code:      code,
		CheckType: CheckTypeTCP,
		Request:   NewAWSRequest(),
	}
}

// CheckLatency does a latency check for a region
func (r *AWSRegion) CheckLatency(wg *sync.WaitGroup) {
	defer wg.Done()

	if r.CheckType == CheckTypeHTTP || r.CheckType == CheckTypeHTTPS {
		r.checkLatencyHTTP(r.CheckType == CheckTypeHTTPS)
	} else {
		r.checkLatencyTCP()
	}
}

// checkLatencyHTTP Test Latency via HTTP
func (r *AWSRegion) checkLatencyHTTP(https bool) {
	url := r.Target.GetURL()
	l, err := r.Request.Do(useragent, url, RequestTypeHTTP)
	r.recordAttempt(l, err)
}

// checkLatencyTCP Test Latency via TCP
func (r *AWSRegion) checkLatencyTCP() {
	tcpAddr, err := r.Target.GetIP()
	if err != nil {
		r.recordAttempt(0, err)
		return
	}

	l, err := r.Request.Do(useragent, tcpAddr.String(), RequestTypeTCP)
	r.recordAttempt(l, err)
}

func (r *AWSRegion) recordAttempt(latency time.Duration, err error) {
	r.Attempts = append(r.Attempts, AttemptResult{
		Latency: latency,
		Err:     err,
	})

	if err != nil {
		r.Error = err
		return
	}

	r.Latencies = append(r.Latencies, latency)
}

// SuccessfulAttempts returns the number of checks completed without an error.
func (r *AWSRegion) SuccessfulAttempts() int {
	if len(r.Attempts) == 0 {
		return len(r.Latencies)
	}

	successful := 0
	for _, attempt := range r.Attempts {
		if attempt.Err == nil {
			successful++
		}
	}
	return successful
}

// FailedAttempts returns the number of checks completed with an error.
func (r *AWSRegion) FailedAttempts() int {
	if len(r.Attempts) == 0 {
		if r.Error != nil {
			return 1
		}
		return 0
	}

	failed := 0
	for _, attempt := range r.Attempts {
		if attempt.Err != nil {
			failed++
		}
	}
	return failed
}

// AverageLatency returns the average latency of successful attempts.
func (r *AWSRegion) AverageLatency() (time.Duration, bool) {
	if len(r.Attempts) == 0 {
		if len(r.Latencies) == 0 {
			return 0, false
		}

		var total time.Duration
		for _, latency := range r.Latencies {
			total += latency
		}
		return total / time.Duration(len(r.Latencies)), true
	}

	var total time.Duration
	successful := 0
	for _, attempt := range r.Attempts {
		if attempt.Err != nil {
			continue
		}
		total += attempt.Latency
		successful++
	}
	if successful == 0 {
		return 0, false
	}
	return total / time.Duration(successful), true
}

// LastError returns the error from the most recent failed attempt.
func (r *AWSRegion) LastError() error {
	for i := len(r.Attempts) - 1; i >= 0; i-- {
		if r.Attempts[i].Err != nil {
			return r.Attempts[i].Err
		}
	}
	return r.Error
}

// GetLatency returns Latency in ms
func (r *AWSRegion) GetLatency() float64 {
	latency, ok := r.AverageLatency()
	if !ok {
		return 0
	}
	return Duration2ms(latency)
}

// GetLatencyStr returns Latency in string
func (r *AWSRegion) GetLatencyStr() string {
	if latency, ok := r.AverageLatency(); ok {
		return fmt.Sprintf("%.2f ms", Duration2ms(latency))
	}
	if err := r.LastError(); err != nil {
		return err.Error()
	}
	return "-"
}

// --------------------------------------------

// AWSRegions slice of the AWSRegion
type AWSRegions []AWSRegion

// Len returns a count of regions
func (rs AWSRegions) Len() int {
	return len(rs)
}

// Less return a result of latency compare between two regions
func (rs AWSRegions) Less(i, j int) bool {
	left, leftOK := rs[i].AverageLatency()
	right, rightOK := rs[j].AverageLatency()

	if leftOK != rightOK {
		return leftOK
	}
	if !leftOK {
		return false
	}
	return left < right
}

// Swap two regions by index
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
		rs[i].CheckType = checkType
	}
}

// SetDefaultTarget sets default target instance
func (rs AWSRegions) SetDefaultTarget() {
	rs.SetTarget(func(r *AWSRegion) {
		r.Target = &AWSTarget{
			HTTPS:   r.CheckType == CheckTypeHTTPS,
			Code:    r.Code,
			Service: r.Service,
			Rnd:     mkRandomString(13),
		}
	})
}

// SetTarget sets default target instance for all regions
func (rs AWSRegions) SetTarget(fn func(r *AWSRegion)) {
	for i := range rs {
		fn(&rs[i])
	}
}
