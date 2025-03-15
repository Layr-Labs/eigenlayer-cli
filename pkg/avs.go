package pkg

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/avs"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/urfave/cli/v2"
)

func AvsCmd(prompter utils.Prompter) *cli.Command {
	var avsCmd = &cli.Command{
		Name:  "avs",
		Usage: "Manage registrations with actively validated services",
		Subcommands: []*cli.Command{
			avs.SpecsCmd(),
			avs.ConfigCmd(prompter),
			avs.RegisterCmd(prompter),
			avs.OptInCmd(prompter),
			avs.OptOutCmd(prompter),
			avs.DeregisterCmd(prompter),
			avs.StatusCmd(prompter),
		},
	}

	return avsCmd
}
