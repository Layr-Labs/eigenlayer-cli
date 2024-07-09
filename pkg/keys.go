package pkg

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/keys"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	"github.com/urfave/cli/v2"
)

func KeysCmd(p utils.Prompter) *cli.Command {
	var keysCmd = &cli.Command{
		Name:  "keys",
		Usage: "Manage the keys used in EigenLayer ecosystem",
		Subcommands: []*cli.Command{
			keys.CreateCmd(p),
			keys.ListCmd(),
			keys.ImportCmd(p),
			keys.ExportCmd(p),
		},
	}

	return keysCmd

}
