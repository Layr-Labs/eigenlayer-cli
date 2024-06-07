package rewards

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
		Usage:   "Show rewards",
		After:   telemetry.AfterRunAction(),
		Action:  showRewards,
		Flags: []cli.Flag{
			&flags.ConfigurationFileFlag,
			&flags.NumberOfDaysFlag,
			&flags.OperatorSetsFlag,
			&flags.AvsAddressesFlag,
		},
	}
}

func showRewards(cCtx *cli.Context) error {
	fmt.Println("unimplemented")
	return nil
}
