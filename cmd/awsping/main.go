package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/ekalinin/awsping"
)

var (
	repeats     = flag.Int("repeats", 1, "Number of repeats")
	useHTTP     = flag.Bool("http", false, "Use http transport (default is tcp)")
	useHTTPS    = flag.Bool("https", false, "Use https transport (default is tcp)")
	showVer     = flag.Bool("v", false, "Show version")
	verbose     = flag.Int("verbose", 0, "Verbosity level")
	service     = flag.String("service", "dynamodb", "AWS Service: ec2, sdb, sns, sqs, ...")
	listRegions = flag.Bool("list-regions", false, "Show list of regions")
)

func main() {
	flag.Parse()

	if *showVer {
		fmt.Println(awsping.Version)
		os.Exit(0)
	}

	if *listRegions {
		regions := awsping.GetRegions(*service)
		lo := awsping.LatencyOutput{Level: awsping.ShowOnlyRegions}
		lo.Show(&regions)
		os.Exit(0)
	}

	rand.Seed(time.Now().UnixNano())
	regions := awsping.CalcLatency(*repeats, *useHTTP, *useHTTPS, *service)
	lo := awsping.LatencyOutput{Level: *verbose, Repeats: *repeats}
	lo.Show(regions)
}
