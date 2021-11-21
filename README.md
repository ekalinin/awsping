# awsping
Console tool to check the latency to each AWS region

[![Go Report Card](https://goreportcard.com/badge/github.com/ekalinin/awsping)](https://goreportcard.com/report/github.com/ekalinin/awsping)
[![codecov](https://codecov.io/gh/ekalinin/awsping/branch/master/graph/badge.svg)](https://codecov.io/gh/ekalinin/awsping)
[![Go Reference](https://pkg.go.dev/badge/github.com/ekalinin/awsping.svg)](https://pkg.go.dev/github.com/ekalinin/awsping)
[![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://opensource.org/licenses/MIT)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/ekalinin/awsping)

# ToC

* [Usage](#usage)
  * [Test via TCP](#test-via-tcp)
  * [Test via HTTP](#test-via-http)
  * [Test via HTTPS](#test-via-https)
  * [Test several times](#test-several-times)
  * [Verbose mode](#verbose-mode)
  * [Get Help](#get-help)
* [Get binary file](#get-binary-file)
* [Build from sources](#build-from-sources)
* [Use with Docker](#use-with-docker)
  * [Build a Docker image](#build-a-docker-image)
  * [Run the Docker image](#run-the-docker-image)

# Usage

## Test via TCP

```bash
➥ ./awsping
Europe (Frankfurt)                    51.86 ms
Europe (Ireland)                      62.86 ms
US-East (Virginia)                   126.39 ms
US-East (Ohio)                       154.81 ms
Asia Pacific (Mumbai)                181.09 ms
US-West (California)                 194.27 ms
US-West (Oregon)                     211.87 ms
South America (São Paulo)            246.20 ms
Asia Pacific (Tokyo)                 309.27 ms
Asia Pacific (Seoul)                 322.76 ms
Asia Pacific (Sydney)                346.37 ms
Asia Pacific (Singapore)             407.91 ms
```

## Test via HTTP

```bash
➥ ./awsping -http
Europe (Frankfurt)                   222.56 ms
Europe (Ireland)                     226.76 ms
US-East (Virginia)                   349.17 ms
US-West (California)                 488.12 ms
US-East (Ohio)                       513.69 ms
Asia Pacific (Mumbai)                528.51 ms
US-West (Oregon)                     532.05 ms
South America (São Paulo)            599.36 ms
Asia Pacific (Seoul)                 715.92 ms
Asia Pacific (Sydney)                721.47 ms
Asia Pacific (Tokyo)                 745.24 ms
Asia Pacific (Singapore)             847.36 ms
```

## Test via HTTPS

```bash
➥ ./awsping -https
Europe (Stockholm)                   216.67 ms
Europe (Frankfurt)                   263.20 ms
Europe (Paris)                       284.32 ms
Europe (Milan)                       305.63 ms
Europe (Ireland)                     327.34 ms
Europe (London)                      332.17 ms
Middle East (Bahrain)                590.74 ms
US-East (N. Virginia)                595.13 ms
Canada (Central)                     628.44 ms
US-East (Ohio)                       635.32 ms
Asia Pacific (Mumbai)                755.56 ms
Asia Pacific (Hong Kong)             843.90 ms
US-West (N. California)              870.65 ms
Asia Pacific (Singapore)             899.50 ms
Africa (Cape Town)                   912.06 ms
US-West (Oregon)                     919.34 ms
South America (São Paulo)            985.93 ms
Asia Pacific (Tokyo)                1122.67 ms
Asia Pacific (Seoul)                1138.76 ms
Asia Pacific (Osaka)                1167.40 ms
Asia Pacific (Sydney)               1328.90 ms
```

## Test several times

```bash
➥ ./awsping -repeats 3
Europe (Frankfurt)                    50.13 ms
Europe (Ireland)                      62.67 ms
US-East (Virginia)                   126.88 ms
US-East (Ohio)                       155.37 ms
US-West (California)                 195.75 ms
US-West (Oregon)                     206.19 ms
Asia Pacific (Mumbai)                222.34 ms
South America (São Paulo)            254.28 ms
Asia Pacific (Tokyo)                 308.52 ms
Asia Pacific (Seoul)                 325.93 ms
Asia Pacific (Sydney)                349.62 ms
Asia Pacific (Singapore)             378.53 ms
```

## Verbose mode

```bash
➥ ./awsping -repeats 3 -verbose 1
      Code            Region                                      Latency
    0 eu-central-1    Europe (Frankfurt)                         47.39 ms
    1 eu-west-1       Europe (Ireland)                           62.28 ms
    2 us-east-1       US-East (Virginia)                        128.45 ms
    3 us-east-2       US-East (Ohio)                            155.53 ms
    4 us-west-1       US-West (California)                      194.37 ms
    5 us-west-2       US-West (Oregon)                          208.91 ms
    6 ap-south-1      Asia Pacific (Mumbai)                     226.59 ms
    7 sa-east-1       South America (São Paulo)                 254.67 ms
    8 ap-northeast-1  Asia Pacific (Tokyo)                      301.97 ms
    9 ap-northeast-2  Asia Pacific (Seoul)                      323.10 ms
   10 ap-southeast-2  Asia Pacific (Sydney)                     341.26 ms
   11 ap-southeast-1  Asia Pacific (Singapore)                  397.47 ms
```

```bash
➥ ./awsping -repeats 3 -verbose 2
      Code            Region                             Try #1          Try #2          Try #3     Avg Latency
    0 eu-central-1    Europe (Frankfurt)               45.18 ms        45.46 ms        45.68 ms        45.44 ms
    1 eu-west-1       Europe (Ireland)                 61.89 ms        62.99 ms        62.98 ms        62.62 ms
    2 us-east-1       US-East (Virginia)              125.15 ms       126.75 ms       126.49 ms       126.13 ms
    3 us-east-2       US-East (Ohio)                  154.05 ms       154.28 ms       153.53 ms       153.96 ms
    4 us-west-1       US-West (California)            196.20 ms       195.05 ms       193.76 ms       195.00 ms
    5 us-west-2       US-West (Oregon)                204.04 ms       203.97 ms       203.84 ms       203.95 ms
    6 ap-south-1      Asia Pacific (Mumbai)           175.27 ms       300.68 ms       172.18 ms       216.05 ms
    7 sa-east-1       South America (São Paulo)       243.48 ms       247.12 ms       248.32 ms       246.31 ms
    8 ap-northeast-1  Asia Pacific (Tokyo)            324.78 ms       312.70 ms       319.02 ms       318.83 ms
    9 ap-northeast-2  Asia Pacific (Seoul)            328.96 ms       327.65 ms       326.17 ms       327.59 ms
   10 ap-southeast-2  Asia Pacific (Sydney)           388.17 ms       347.74 ms       393.58 ms       376.50 ms
   11 ap-southeast-1  Asia Pacific (Singapore)        409.53 ms       403.61 ms       405.84 ms       406.33 ms
```

## Get Help

```bash
➜ ./awsping -h
Usage of ./awsping:
  -http
    	Use http transport (default is tcp)
  -https
    	Use https transport (default is tcp)
  -list-regions
    	Show list of regions
  -repeats int
    	Number of repeats (default 1)
  -service string
    	AWS Service: ec2, sdb, sns, sqs, ... (default "dynamodb")
  -v	Show version
  -verbose int
    	Verbosity level
```

# Get binary file

```bash
$ wget https://github.com/ekalinin/awsping/releases/download/0.5.2/awsping.linux.amd64.tgz
$ tar xzvf awsping.linux.amd64.tgz
$ chmod +x awsping
$ ./awsping -v
0.5.2
```

# Build from sources

```bash
➥ make build
```

# Use with Docker
## Build a Docker image

```
$ docker build -t awsping .
```

## Run the Docker image
```
$ docker run --rm awsping
```

Arguments can be used as mentioned in the _Usage_ section.

i.e.:
```
$ docker run --rm awsping -repeats 3 -verbose 2
```
