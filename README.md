# awsping
Console tool to check the latency to each AWS region

# Usage

## Test via TCP

```bash
➥ ./awsping 
      Region                                      Latency
    0 Europe (Frankfurt)                         44.67 ms
    1 Europe (Ireland)                           46.96 ms
    2 US-East (Virginia)                         61.49 ms
    3 US-West (Oregon)                           76.90 ms
    4 US-West (California)                       77.75 ms
    5 Asia Pacific (Mumbai)                      78.99 ms
    6 Asia Pacific (Singapore)                  104.05 ms
    7 Asia Pacific (Tokyo)                      104.88 ms
    8 South America (São Paulo)                 104.69 ms
    9 Asia Pacific (Sydney)                     104.72 ms
   10 Asia Pacific (Seoul)                      106.98 ms
```

## Test via HTTP

```bash
➥ ./awsping -http
      Region                                      Latency
    0 Europe (Frankfurt)                         77.57 ms
    1 Europe (Ireland)                          128.57 ms
    2 US-East (Virginia)                        254.92 ms
    3 Asia Pacific (Mumbai)                     326.27 ms
    4 US-West (California)                      389.81 ms
    5 US-West (Oregon)                          408.48 ms
    6 South America (São Paulo)                 497.38 ms
    7 Asia Pacific (Seoul)                      665.45 ms
    8 Asia Pacific (Tokyo)                      665.49 ms
    9 Asia Pacific (Sydney)                     669.75 ms
   10 Asia Pacific (Singapore)                  763.94 ms
```

## Test several times

```bash
➥ ./awsping -repeats 3
      Region                                      Latency
    0 Europe (Frankfurt)                         37.19 ms
    1 Europe (Ireland)                           62.48 ms
    2 US-East (Virginia)                        126.30 ms
    3 Asia Pacific (Mumbai)                     161.25 ms
    4 US-West (California)                      193.53 ms
    5 US-West (Oregon)                          204.38 ms
    6 South America (São Paulo)                 246.15 ms
    7 Asia Pacific (Tokyo)                      318.53 ms
    8 Asia Pacific (Seoul)                      327.51 ms
    9 Asia Pacific (Sydney)                     333.77 ms
   10 Asia Pacific (Singapore)                  387.18 ms
``` 

# Get binary file

```bash
$ wget https://github.com/ekalinin/awsping/releases/download/0.3.0/awsping.linux.amd64.tgz
$ tar xzvf awsping.linux.amd64.tgz
$ chmod +x awsping
$ $ ./awsping -v
0.3.0
```

# Build from sources

```bash
➥ make build
```
