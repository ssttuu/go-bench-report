FROM golang:1.11 AS go-build

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go install

FROM alpine:3.8
COPY --from=go-build /go/bin/go-bench-report /bin/go-bench-report
ENTRYPOINT [ "/bin/go-bench-report" ]
