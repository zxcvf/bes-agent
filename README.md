# BES Agent

[![Build Status](https://travis-ci.org/BES/bes-agent.svg?branch=master)](https://travis-ci.org/BES/bes-agent)
[![Go Report Card](https://goreportcard.com/badge/bes-agent)](https://goreportcard.com/report/bes-agent)
[![codecov](https://codecov.io/gh/BES/bes-agent/branch/master/graph/badge.svg)](https://codecov.io/gh/BES/bes-agent)

[中文版 README](README_zh-CN.md)

BES Agent is written in Go for collecting metrics from the system it's
running on, or from other services, and sending them to [BES](https://cloud.oneapm.com).

## Building from source

To build BES Agent from the source code yourself you need to have a working Go environment with [version 1.7+](https://golang.org/doc/install).

```
$ mkdir -p $GOPATH/src/github.com/BES
$ cd $GOPATH/src/github.com/BES
$ git clone https://bes-agent
$ cd bes-agent
$ make build
```

## Usage

First you need to set a license key, which can be found at [https://cloud.oneapm.com/#/settings](https://cloud.oneapm.com/#/settings).

```
$ cp bes-agent.conf.example bes-agent.conf
$ vi bes-agent.conf
...
license_key = "*********************"
```

Run the agent in foreground:

```
$ ./bin/bes-agent
```

For more options, see:

```
$ ./bin/bes-agent --help
```

## Related works

I have been influenced by the following great works:

- [ddagent](https://github.com/datadog/dd-agent)
- [telegraf](https://github.com/influxdata/telegraf)
- [prometheus](https://github.com/prometheus/prometheus)
- [mackerel](https://github.com/mackerelio/mackerel-agent)
