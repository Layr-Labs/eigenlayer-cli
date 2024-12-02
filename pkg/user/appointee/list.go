package appointee

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/urfave/cli/v2"
)

func ListCmd() *cli.Command {
	listCmd := &cli.Command{
		Name:      "list",
		Usage:     "user appointee list <AccountAddress> <TargetAddress> <Selector>",
		UsageText: "Lists all appointee users for an account with the provided permissions.",
		Description: `
		Lists all appointee users for an account with the provided permissions.
		`,
		After: telemetry.AfterRunAction(),
		Flags: []cli.Flag{
			&flags.VerboseFlag,
			&AccountAddressFlag,
		},
	}

	return listCmd
}
