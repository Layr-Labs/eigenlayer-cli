package allocations

import "github.com/urfave/cli/v2"

var (
	BipsToAllocateFlag = cli.Uint64Flag{
		Name:    "bips-to-allocate",
		Aliases: []string{"bta", "bips", "bps"},
		Usage:   "Bips to allocate to the strategy",
		EnvVars: []string{"BIPS_TO_ALLOCATE"},
	}

	EnvironmentFlag = cli.StringFlag{
		Name:    "environment",
		Aliases: []string{"env"},
		Usage:   "environment to use. Currently supports 'preprod' ,'testnet' and 'prod'. If not provided, it will be inferred based on network",
		EnvVars: []string{"ENVIRONMENT"},
	}
)
