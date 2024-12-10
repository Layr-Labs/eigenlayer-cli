package pkg

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/rewards"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/urfave/cli/v2"
)

func RewardsCmd(p utils.Prompter) *cli.Command {
	var rewardsCmd = &cli.Command{
		Name:  "rewards",
		Usage: "Execute onchain operations for the rewards",
		Subcommands: []*cli.Command{
			rewards.ClaimCmd(p),
			rewards.SetClaimerCmd(p),
			rewards.ShowCmd(p),
			rewards.SetOperatorSplitCmd(p),
			rewards.GetOperatorSplitCmd(p),
		},
	}

	return rewardsCmd
}
