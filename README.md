# Go Bench Report
[![Build Status](https://travis-ci.org/ssttuu/go-bench-report.svg?branch=master)](https://travis-ci.org/ssttuu/go-bench-report)
[![Go Report Card](https://goreportcard.com/badge/github.com/ssttuu/go-bench-report)](https://goreportcard.com/report/github.com/ssttuu/go-bench-report)

`go-bench-report` takes the standard benchmark output of
`go test` and uploads the metrics to StackDriver.

## Quickstart

```console
$ go get github.com/ssttuu/go-bench-report 
```

```console
$ export GOOGLE_APPLICATION_CREDENTIALS="~/path/to/credentials.json"
$ go test -bench=. -benchmem ./... | \
    go-bench-report \
        --projectID=[GCP Project ID] \
        --branch=`git rev-parse --abbrev-ref HEAD` \
        --githash=`git rev-parse --short=10 HEAD` \
        --version=`cat VERSION`
```

## Future Work

* [ ] Support multiple Backends (ie. Datadog)
* [ ] Support file input
* [ ] Support multiple input Benchmark formats
