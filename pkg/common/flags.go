package common

import "github.com/urfave/cli/v2"

var (
	NetworkFlag = cli.StringFlag{
		Name:        "network",
		Aliases:     []string{"n"},
		Usage:       "Network to use. Currently supports 'mainnet' and 'holesky'",
		DefaultText: "mainnet",
		EnvVars:     []string{"NETWORK"},
	}

	OutputFileFlag = cli.StringFlag{
		Name:    "output-file-path",
		Aliases: []string{"o"},
		Usage:   "Path to the output file",
		EnvVars: []string{"OUTPUT_FILE_PATH"},
	}
)
