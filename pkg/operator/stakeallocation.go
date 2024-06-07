package operator

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/operator/stakeallocation"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	"github.com/urfave/cli/v2"
)

func StakeAllocationCmd(p utils.Prompter) *cli.Command {
	return &cli.Command{
		Name:   "stake-allocation",
		Usage:  "Stake allocation commands",
		Hidden: true,
		Subcommands: []*cli.Command{
			stakeallocation.ShowCmd(p),
			stakeallocation.UpdateCmd(p),
		},
	}
}
