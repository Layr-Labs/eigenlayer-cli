package verifiable

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/registry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/verifiable/container"
	"github.com/urfave/cli/v2"
)

func NewContainerCmd(prompter utils.Prompter) *cli.Command {
	var containerCmd = &cli.Command{
		Name:  "container",
		Usage: "Manage operations related to container verification.",
		Subcommands: []*cli.Command{
			container.NewSignContainerCmd(prompter, registry.NewOciRegistryController()),
			container.NewVerifyContainerCmd(registry.NewOciRegistryController()),
		},
	}

	return containerCmd
}
