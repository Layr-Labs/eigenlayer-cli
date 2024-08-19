package flags

import "github.com/urfave/cli/v2"

var (
	AvsAddressFlag = cli.StringFlag{
		Name:    "avs-address",
		Usage:   "The address of the AVS",
		EnvVars: []string{"AVS_ADDRESS"},
		Aliases: []string{"avsa"},
	}
)
