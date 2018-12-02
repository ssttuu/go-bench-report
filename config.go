package main

import (
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

type Config struct {
	ProjectID string
	Branch    string
	Githash   string
	Version   string
}

func readInConfig() (*Config, error) {
	projectID := pflag.StringP("projectID", "p", "", "Google Project ID")
	branch := pflag.StringP("branch", "b", "", "Git Branch")
	githash := pflag.StringP("githash", "g", "", "Git Hash")
	version := pflag.StringP("version", "v", "", "Release Version")
	pflag.Parse()

	if *projectID == "" {
		return nil, errors.New("projectID must be set")
	}

	return &Config{
		ProjectID: *projectID,
		Branch:    *branch,
		Githash:   *githash,
		Version:   *version,
	}, nil
}
