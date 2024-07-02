package rewards

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/urfave/cli/v2"
)

func ShowCmd(p utils.Prompter) *cli.Command {
	showCmd := &cli.Command{
		Name:      "show",
		Usage:     "Show rewards for an address",
		UsageText: "show",
		Description: `
		Command to show rewards for earners
		`,
		Flags: []cli.Flag{
			&EarnerAddressFlag,
			&NumberOfDaysFlag,
			&AVSAddressesFlag,
			&common.NetworkFlag,
			&common.OutputFileFlag,
		},
		Action: func(cCtx *cli.Context) error {

			return nil
		},
	}

	return showCmd
}
