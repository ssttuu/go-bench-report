# Benchmark

`benchmark` takes the standard benchmark format of
`go test` and uploads the metrics to StackDriver.

# Quickstart

```
go get github.com/stupschwartz/benchmark 
```

```
go test -bench=. -benchmem ./... | \
    benchmark \
        --projectID=[GCP Project ID] \
        --branch=`git rev-parse --abbrev-ref HEAD` \
        --githash=`git rev-parse --short=10 HEAD` \
        --version=`cat VERSION`
```
