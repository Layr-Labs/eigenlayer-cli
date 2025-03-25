package pkg

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/verifiable"
	"github.com/urfave/cli/v2"
)

func NewVerifiableCmd(prompter utils.Prompter) *cli.Command {
	var verifiableCmd = &cli.Command{
		Name:  "verifiable",
		Usage: "Manage operations related to verifiable systems.",
		Subcommands: []*cli.Command{
			verifiable.NewContainerCmd(prompter),
		},
	}

	return verifiableCmd
}
