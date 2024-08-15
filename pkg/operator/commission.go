package operator

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/operator/commission"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/urfave/cli/v2"
)

func CommissionCmd(p utils.Prompter) *cli.Command {
	var commissionCmd = &cli.Command{
		Name:  "commission",
		Usage: "Execute onchain operations for the operator's commission",
		Subcommands: []*cli.Command{
			commission.UpdateCmd(p),
		},
	}

	return commissionCmd
}
