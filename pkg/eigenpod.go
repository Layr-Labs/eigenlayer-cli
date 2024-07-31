package pkg

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/eigenpod"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/urfave/cli/v2"
)

func EigenPodCmd(p utils.Prompter) *cli.Command {
	var eigenPodCmd = &cli.Command{
		Name:  "eigenpod",
		Usage: "Manage the EigenPods in EigenLayer ecosystem",
		Subcommands: []*cli.Command{
			eigenpod.StatusCmd(p),
		},
	}

	return eigenPodCmd

}
