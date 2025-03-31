package release

import "github.com/urfave/cli/v2"

var (
	AvsIdFlag = cli.StringFlag{
		Name:    "avs-id",
		Aliases: []string{"aid"},
		Usage:   "AVS ID to list release keys for",
		EnvVars: []string{"AVS_ID"},
	}

	OperatorIdFlag = cli.StringFlag{
		Name:    "operator-id",
		Aliases: []string{"oid"},
		Usage:   "Operator ID to list releases for",
		EnvVars: []string{"OPERATOR_ID"},
	}
)
