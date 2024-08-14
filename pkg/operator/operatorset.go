package operator

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/operator/operatorset"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/urfave/cli/v2"
)

func OperatorSetCmd(p utils.Prompter) *cli.Command {
	var operatorSetCmd = &cli.Command{
		Name:        "operatorset",
		Usage:       "Manage operator's operator set configurations",
		Description: "Manage the operator's operator sets",
		Subcommands: []*cli.Command{
			operatorset.DeregisterCmd(p),
		},
	}

	return operatorSetCmd
}
