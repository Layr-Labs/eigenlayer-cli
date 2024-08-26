package admin

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/urfave/cli/v2"
)

func ListPendingCmd() *cli.Command {
	listPendingCmd := &cli.Command{
		Name:      "list-pending-admins",
		Usage:     "user admin list-pending-admins --account-address <AccountAddress>",
		UsageText: "List all users who are pending admin acceptance.",
		Description: `
		List all users who are pending admin acceptance.
		`,
		After: telemetry.AfterRunAction(),
		Flags: []cli.Flag{
			&flags.VerboseFlag,
			&AccountAddressFlag,
		},
	}

	return listPendingCmd
}
