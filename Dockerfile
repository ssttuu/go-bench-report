FROM golang:1.12 AS go-build

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go install

################
FROM scratch
COPY --from=go-build /go/bin/go-bench-report /go/bin/go-bench-report
ENTRYPOINT [ "/go/bin/go-bench-report" ]
