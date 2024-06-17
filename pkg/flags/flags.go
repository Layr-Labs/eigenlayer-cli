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
		Aliases:  []string{"as"},
	}

	AvsAddressFlag = cli.StringSliceFlag{
		Name:     "avs-address",
		Usage:    "an AVS address",
		Required: true,
		Aliases:  []string{"a"},
	}

	OperatorSetsFlag = cli.StringSliceFlag{
		Name:     "operator-sets",
		Usage:    "Comma separated list of operator sets AVSAddress#OperatorSetId",
		Required: false,
		Aliases:  []string{"ops"},
	}

	OperatorSetFlag = cli.StringFlag{
		Name:     "operator-set",
		Usage:    "Operator set identifier AVSAddress#OperatorSetId",
		Required: true,
		Aliases:  []string{"op"},
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

	AllocationBipsFlag = cli.StringFlag{
		Name:     "allocation-bips",
		Usage:    "Allocation to update",
		Required: true,
		Aliases:  []string{"ap"},
	}

	StakeSourceFlag = cli.StringFlag{
		Name:     "stake-source",
		Usage:    "The source of stake in case of allocation. The destination of stake if deallocation. Options are 'slashable', 'nonslashable' or 'both'. ",
		Required: true,
		Aliases:  []string{"s"},
	}

	ShowMagnitudesFlag = cli.StringFlag{
		Name:     "show-magnitudes",
		Usage:    "Show magnitudes of stake share",
		Required: true,
		Aliases:  []string{"m"},
	}

	RebalanceFilePathFlag = cli.PathFlag{
		Name: "rebalance-file-path",
		Usage: `Path to the CSV file. 
	The CSV file should have the following columns: operator set,allocation percentage.
	This file must have all the operator sets and their allocation percentages for a strategy.`,
		Required: true,
		Aliases:  []string{"r"},
	}

	StrategyAddressFlag = cli.StringFlag{
		Name:     "strategy-address",
		Usage:    "Address of the strategy contract",
		Required: true,
		Aliases:  []string{"sa"},
	}

	OutputFilePathFlag = cli.StringFlag{
		Name:     "output-file-path",
		Usage:    "Path to the output file. It will be in a CSV format",
		Required: false,
		Aliases:  []string{"o"},
	}
)
