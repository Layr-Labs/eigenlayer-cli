package pkg

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/operator"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	"github.com/urfave/cli/v2"
)

func OperatorCmd(p utils.Prompter) *cli.Command {
	var operatorCmd = &cli.Command{
		Name:  "operator",
		Usage: "Execute onchain operations for the operator",
		Subcommands: []*cli.Command{
			operator.KeysCmd(p),
			operator.ConfigCmd(p),
			operator.RegisterCmd(p),
			operator.StatusCmd(p),
			operator.UpdateCmd(p),
			operator.StakeAllocationCmd(p),
			operator.RewardsCmd(p),
		},
	}

	return operatorCmd

}
