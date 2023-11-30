package operator

import (
	"github.com/Layr-Labs/eigenlayer-cli/operator/keys"
	"github.com/Layr-Labs/eigenlayer-cli/utils"
	"github.com/urfave/cli/v2"
)

func KeysCmd(p utils.Prompter) *cli.Command {
	var keysCmd = &cli.Command{
		Name:  "keys",
		Usage: "Manage the operator's keys",
		Subcommands: []*cli.Command{
			keys.CreateCmd(p),
		},
	}

	return keysCmd

}
