package operator

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/operator/rewards"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	"github.com/urfave/cli/v2"
)

func RewardsCmd(p utils.Prompter) *cli.Command {
	return &cli.Command{
		Name:   "rewards",
		Usage:  "Rewards commands",
		Hidden: true,
		Subcommands: []*cli.Command{
			rewards.ShowCmd(p),
		},
	}
}
