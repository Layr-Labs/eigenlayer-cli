package pkg

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/verifiable/container"
	"github.com/urfave/cli/v2"
)

func NewVerifiableCmd(prompter utils.Prompter) *cli.Command {
	var userCmd = &cli.Command{
		Name:  "verifiable",
		Usage: "Manage operations related to verifiable systems.",
		Subcommands: []*cli.Command{
			container.NewSignContainerCmd(prompter),
			container.NewVerifyContainerCmd(),
		},
	}

	return userCmd
}
