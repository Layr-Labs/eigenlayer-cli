package stakeallocation

import (
	"fmt"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	
	"github.com/urfave/cli/v2"
)

func ShowCmd(p utils.Prompter) *cli.Command {
	return &cli.Command{
		Name:    "show",
		Aliases: []string{"s"},
		Usage:   "Show stake allocation",
		Description: `
		Show the stake allocation for the operator
		`,
		Action: showStakeAllocation,
		After:  telemetry.AfterRunAction(),
		Flags: []cli.Flag{
			&flags.ConfigurationFileFlag,
			&flags.AvsAddressesFlag,
			&flags.OperatorSetsFlag,
		},
	}
}

func showStakeAllocation(ctx *cli.Context) error {
	fmt.Println("unimplemented")
	return nil
}
