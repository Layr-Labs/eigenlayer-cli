package split

import "github.com/urfave/cli/v2"

var (
	OperatorSplitFlag = cli.IntFlag{
		Name:     "operator-split",
		Aliases:  []string{"os"},
		Usage:    "Split for the operator in bips (e.g. 1000 = 10%)",
		Required: false,
		EnvVars:  []string{"OPERATOR_SPLIT"},
	}

	AVSAddressFlag = cli.StringFlag{
		Name:    "avs-address",
		Aliases: []string{"aa"},
		Usage:   "AVS address to set operator split",
		EnvVars: []string{"AVS_ADDRESS"},
	}

	OperatorSetIdFlag = cli.IntFlag{
		Name:    "operator-set-id",
		Aliases: []string{"osi"},
		Usage:   "Operator set ID to set operator split",
		EnvVars: []string{"OPERATOR_SET_ID"},
	}
)
