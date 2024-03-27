package config

import "github.com/urfave/cli/v2"

var (
	YesFlag = cli.BoolFlag{
		Name:    "yes",
		Aliases: []string{"y"},
		Usage:   "Use this flag to skip confirmation prompts. When used the operator config file and metadata file will be created with default values.",
		EnvVars: []string{"YES"},
	}
)
