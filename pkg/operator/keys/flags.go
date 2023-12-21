package keys

import "github.com/urfave/cli/v2"

var (
	KeyTypeFlag = cli.StringFlag{
		Name:     "key-type",
		Aliases:  []string{"k"},
		Required: true,
		Usage:    "Type of key you want to create. Currently supports 'ecdsa' and 'bls'",
		EnvVars:  []string{"KEY_TYPE"},
	}

	InsecureFlag = cli.BoolFlag{
		Name:    "insecure",
		Aliases: []string{"i"},
		Usage:   "Use this flag to skip password validation",
		EnvVars: []string{"INSECURE"},
	}

	KeyPathFlag = cli.StringFlag{
		Name:    "key-path",
		Aliases: []string{"p"},
		Usage:   "Use this flag to specify the path of the key",
		EnvVars: []string{"KEY_PATH"},
	}
)
