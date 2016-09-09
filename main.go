package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

var (
	Version   string = "0.1.0"
	GitHub    string = "https://github.com/ekalinin/awsping"
	UserAgent string = fmt.Sprintf("AwsPing/%s (+%s)", Version, GitHub)
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func mkRandoString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

type AWSRegion struct {
	Name    string
	Code    string
	Latency time.Duration
	Error   error
}

func (r *AWSRegion) CheckLatency(wg *sync.WaitGroup) {
	url := fmt.Sprintf("http://dynamodb.%s.amazonaws.com/ping?x=%s",
		r.Code, mkRandoString(13))
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", UserAgent)

	start := time.Now()
	resp, err := client.Do(req)
	r.Latency = time.Since(start)

	r.Error = err
	resp.Body.Close()

	wg.Done()
}

func main() {
	regions := []AWSRegion{
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
	wg.Add(len(regions))

	for i := range regions {
		go regions[i].CheckLatency(&wg)
	}

	wg.Wait()

	outFmt := "|%5v|%30s|%20s|%20v|\n"
	fmt.Printf(outFmt, "", "Region", "Latency", "Error")
	for i, r := range regions {
		fmt.Printf(outFmt, i, r.Name, r.Latency, r.Error)
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
