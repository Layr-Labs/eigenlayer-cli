package appointee

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/urfave/cli/v2"
)

func RemoveCmd() *cli.Command {
	removeCmd := &cli.Command{
		Name:      "remove",
		Usage:     "user appointee remove <AccountAddress> <AppointeeAddress> <TargetAddress> <Selector>",
		UsageText: "The remove command allows you to check remove for a user's permission",
		Description: `
		The remove command allows you to check remove for a user's permission'.
		`,
		After: telemetry.AfterRunAction(),
		Flags: []cli.Flag{
			&flags.VerboseFlag,
			&AccountAddressFlag,
			&AppointeeAddressFlag,
			&TargetAddressFlag,
			&SelectorFlag,
		},
	}

	return removeCmd
}
