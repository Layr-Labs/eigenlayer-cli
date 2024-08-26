package operator

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/operator/allocations"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/urfave/cli/v2"
)

func AllocationsCmd(p utils.Prompter) *cli.Command {
	var allocationsCmd = &cli.Command{
		Name:  "allocations",
		Usage: "Stake allocation commands for operators",
		Subcommands: []*cli.Command{
			allocations.ShowCmd(p),
			allocations.UpdateCmd(p),
			allocations.InitializeDelayCmd(p),
		},
	}

	return allocationsCmd
}
