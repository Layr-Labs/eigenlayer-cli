package stakeallocation

import (
	"fmt"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/urfave/cli/v2"
)

func RebalanceCmd(p utils.Prompter) *cli.Command {
	return &cli.Command{
		Name:    "rebalance",
		Aliases: []string{"r"},
		Usage:   "Rebalance stake allocation",
		Description: `
Rebalance the stake allocation for the operator for a particular strategy.
This CSV file requires the following columns for only one stragegy:
avs address,operator set,allocation bips
Example:
0xabcd,1,10
0x1234,2,15
		`,
		Action: rebalanceStakeAllocation,
		After:  telemetry.AfterRunAction(),
		Flags: []cli.Flag{
			&flags.ConfigurationFileFlag,
			&flags.StrategyAddressFlag,
			&flags.ShowMagnitudesFlag,
			&flags.RebalanceFilePathFlag,
			&flags.DryRunFlag,
			&flags.BroadcastFlag,
			&flags.OutputFilePathFlag,
		},
	}
}

func rebalanceStakeAllocation(ctx *cli.Context) error {
	fmt.Println("unimplemented")
	return nil
}
