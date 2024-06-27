package operator

import "github.com/urfave/cli/v2"

var (
	ConfigurationFilePathFlag = cli.StringFlag{
		Name:     "configuration-file",
		Aliases:  []string{"c"},
		Usage:    "Path to the configuration file",
		Required: true,
		EnvVars:  []string{"NODE_OPERATOR_CONFIG_FILE"},
	}

	ClaimerAddressFlag = cli.StringFlag{
		Name:     "claimer-address",
		Aliases:  []string{"a"},
		Usage:    "Address of the claimer",
		Required: true,
		EnvVars:  []string{"NODE_OPERATOR_CLAIMER_ADDRESS"},
	}
)
