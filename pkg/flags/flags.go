package flags

import "github.com/urfave/cli/v2"

var (
	ConfigurationFileFlag = cli.StringFlag{
		Name:     "configuration-file",
		Usage:    "Path to the configuration file",
		Required: true,
		Aliases:  []string{"c"},
	}

	AvsAddressesFlag = cli.StringSliceFlag{
		Name:     "avs-addresses",
		Usage:    "Comma separated list of AVS addresses",
		Required: false,
		Aliases:  []string{"a"},
	}

	OperatorSetsFlag = cli.StringSliceFlag{
		Name:     "operator-sets",
		Usage:    "Comma separated list of operator sets",
		Required: false,
		Aliases:  []string{"os"},
	}

	OperatorSetFlag = cli.StringFlag{
		Name:     "operator-set",
		Usage:    "Operator set identifier",
		Required: true,
		Aliases:  []string{"o"},
	}

	NumberOfDaysFlag = cli.IntFlag{
		Name:        "number-of-days",
		Usage:       "Number of days to show rewards for. Negative values to view retroactive rewards.",
		Required:    false,
		DefaultText: "21",
		Aliases:     []string{"n"},
	}

	DryRunFlag = cli.BoolFlag{
		Name:     "dry-run",
		Usage:    "Dry run the command",
		Required: false,
		Aliases:  []string{"d"},
	}

	BroadcastFlag = cli.BoolFlag{
		Name:     "broadcast",
		Usage:    "Broadcast the transaction",
		Required: false,
		Aliases:  []string{"b"},
	}

	AllocationPercentageFlag = cli.StringFlag{
		Name:     "allocation-percentage",
		Usage:    "Allocation to update",
		Required: true,
		Aliases:  []string{"a"},
	}

	StakeSourceFlag = cli.StringFlag{
		Name:     "stake-source",
		Usage:    "The source of stake in case of allocation. The destination of stake if deallocation. Options are 'slashable', 'nonslashable' or 'both'. ",
		Required: true,
		Aliases:  []string{"s"},
	}
)
