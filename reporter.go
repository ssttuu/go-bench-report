package main

import (
	"context"

	"golang.org/x/tools/benchmark/parse"
)

type Reporter interface {
	Upload(ctx context.Context, set parse.Set, cfg *Config) error
	Close() error
}
