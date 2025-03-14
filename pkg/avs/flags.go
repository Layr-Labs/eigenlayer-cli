package avs

import "github.com/urfave/cli/v2"

var (
	ArgFlag = cli.StringSliceFlag{
		Name:    "arg",
		Aliases: []string{"a"},
		Usage:   "Specify or override configuration parameter (key=value)",
	}
)
