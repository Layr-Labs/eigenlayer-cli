package operator

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/operator/config"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/urfave/cli/v2"
)

func ConfigCmd(p utils.Prompter) *cli.Command {
	var configCmd = &cli.Command{
		Name:  "config",
		Usage: "Manage the operator's config",
		Subcommands: []*cli.Command{
			config.CreateCmd(p),
		},
	}

	return configCmd

}
