package operator

import (
	"sort"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/operator/split"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/rewards"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/urfave/cli/v2"
)

func GetOperatorSetSplitCmd(p utils.Prompter) *cli.Command {
	var operatorSplitCmd = &cli.Command{
		Name:  "get-pi-split",
		Usage: "Get programmatic incentives rewards split",
		Action: func(cCtx *cli.Context) error {
			return GetOperatorSplit(cCtx, true, true)
		},
		After: telemetry.AfterRunAction(),
		Flags: getGetOperatorSetSplitFlags(),
	}

	return operatorSplitCmd
}

func getGetOperatorSetSplitFlags() []cli.Flag {
	baseFlags := []cli.Flag{
		&flags.NetworkFlag,
		&flags.ETHRpcUrlFlag,
		&flags.OperatorAddressFlag,
		&split.OperatorSplitFlag,
		&rewards.RewardsCoordinatorAddressFlag,
	}

	sort.Sort(cli.FlagsByName(baseFlags))
	return baseFlags
}
