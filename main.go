package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/spf13/pflag"
	"log"
	"os"

	"golang.org/x/tools/benchmark/parse"
)

func main() {
	ctx := context.Background()

	cfg, err := readInConfig()
	if err != nil {
		pflag.Usage()
		os.Exit(1)
	}

	set, err := parse.ParseSet(bufio.NewReader(os.Stdin))
	if err != nil {
		log.Fatalf("Failed to parse benchmark set")
	}

	s, err := NewStackDriverClient(ctx, cfg.ProjectID)
	if err != nil {
		log.Fatalf("Failed to create stackdriver client %v", err)
	}

	if err = s.Upload(ctx, set, cfg); err != nil {
		log.Fatalf("Failed to upload benchmarks %v", err)
	}

	if err = s.Close(); err != nil {
		log.Fatalf("Failed to close client %v", err)
	}

	fmt.Printf("Done writing time series data.\n")
}
