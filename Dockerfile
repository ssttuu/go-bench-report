FROM golang:1.11 AS go-build

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go install

ENTRYPOINT [ "/go/bin/go-bench-report" ]
