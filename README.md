# awsping
Console tool to check the latency to each AWS region

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

# Get binary file

```bash
$ wget https://github.com/ekalinin/awsping/releases/download/0.3.0/awsping.linux.amd64.tgz
$ tar xzvf awsping.linux.amd64.tgz
$ chmod +x awsping
$ ./awsping -v
0.3.0
```

# Build from sources

```bash
➥ make build
```
