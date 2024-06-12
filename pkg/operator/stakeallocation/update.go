package stakeallocation

import (
	"fmt"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/urfave/cli/v2"
)

func UpdateCmd(p utils.Prompter) *cli.Command {
	return &cli.Command{
		Name:    "update",
		Aliases: []string{"u"},
		Usage:   "Update stake allocation",
		Description: `
		Update the stake allocation for the operator
		`,
		Action: func(context *cli.Context) error {
			return updateStakeAllocation(context, p)
		},
		After: telemetry.AfterRunAction(),
		Flags: []cli.Flag{
			&flags.ConfigurationFileFlag,
			&flags.AvsAddressFlag,
			&flags.OperatorSetFlag,
			&flags.StrategyAddressFlag,
			&flags.AllocationBipsFlag,
			&flags.StakeSourceFlag,
			&flags.ShowMagnitudesFlag,
			&flags.DryRunFlag,
			&flags.BroadcastFlag,
			&flags.OutputFilePathFlag,
		},
	}
}

func updateStakeAllocation(ctx *cli.Context, p utils.Prompter) error {
	fmt.Println("unimplemented")
	return nil
}
