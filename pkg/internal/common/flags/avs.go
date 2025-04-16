package flags

import "github.com/urfave/cli/v2"

var (
	AVSAddressesFlag = cli.StringSliceFlag{
		Name:    "avs-addresses",
		Usage:   "AVS addresses",
		Aliases: []string{"aa"},
		EnvVars: []string{"AVS_ADDRESSES"},
	}

	AVSAddressFlag = cli.StringFlag{
		Name:    "avs-address",
		Usage:   "AVS addresses",
		Aliases: []string{"aa"},
		EnvVars: []string{"AVS_ADDRESS"},
	}

	StrategyAddressesFlag = cli.StringSliceFlag{
		Name:    "strategy-addresses",
		Usage:   "Strategy addresses",
		Aliases: []string{"sa"},
		EnvVars: []string{"STRATEGY_ADDRESSES"},
	}

	StrategyAddressFlag = cli.StringFlag{
		Name:    "strategy-address",
		Usage:   "Strategy addresses",
		Aliases: []string{"sa"},
		EnvVars: []string{"STRATEGY_ADDRESS"},
	}

	OperatorSetIdFlag = cli.Uint64Flag{
		Name:    "operator-set-id",
		Usage:   "Operator set ID",
		Aliases: []string{"osid"},
		EnvVars: []string{"OPERATOR_SET_ID"},
	}

	OperatorSetIdsFlag = cli.Uint64SliceFlag{
		Name:    "operator-set-ids",
		Usage:   "Operator set IDs. Comma separated list of operator set IDs",
		Aliases: []string{"osids"},
		EnvVars: []string{"OPERATOR_SET_IDS"},
	}
)
