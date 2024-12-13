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

func SetOperatorPISplitCmd(p utils.Prompter) *cli.Command {
	var operatorSplitCmd = &cli.Command{
		Name:  "set-pi-split",
		Usage: "Set operator programmatic incentives split",
		Action: func(cCtx *cli.Context) error {
			return SetOperatorSplit(cCtx, p, true)
		},
		After: telemetry.AfterRunAction(),
		Flags: getSetOperatorPISplitFlags(),
	}

	return operatorSplitCmd
}

func getSetOperatorPISplitFlags() []cli.Flag {
	baseFlags := []cli.Flag{
		&flags.NetworkFlag,
		&flags.ETHRpcUrlFlag,
		&flags.OperatorAddressFlag,
		&split.OperatorSplitFlag,
		&rewards.RewardsCoordinatorAddressFlag,
		&flags.BroadcastFlag,
		&flags.OutputTypeFlag,
		&flags.OutputFileFlag,
		&flags.SilentFlag,
	}

	allFlags := append(baseFlags, flags.GetSignerFlags()...)
	sort.Sort(cli.FlagsByName(allFlags))
	return allFlags
}
