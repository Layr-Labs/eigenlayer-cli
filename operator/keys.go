package operator

import (
	"github.com/Layr-Labs/eigenlayer-cli/operator/keys"
	"github.com/urfave/cli/v2"
)

func KeysCmd() *cli.Command {
	var keysCmd = &cli.Command{
		Name:  "keys",
		Usage: "Manage the operator's keys",
		Subcommands: []*cli.Command{
			keys.CreateCmd,
		},
	}

	return keysCmd

}
