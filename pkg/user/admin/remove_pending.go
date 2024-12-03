package admin

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/urfave/cli/v2"
)

func RemovePendingCmd() *cli.Command {
	removeCmd := &cli.Command{
		Name:      "remove-pending-admin",
		Usage:     "user admin remove-pending-admin <AccountAddress> <AdminAddress>",
		UsageText: "Remove a user who is pending admin acceptance.",
		Description: `
		Remove a user who is pending admin acceptance.
		`,
		After: telemetry.AfterRunAction(),
		Flags: []cli.Flag{
			&flags.VerboseFlag,
			&AccountAddressFlag,
			&AdminAddressFlag,
		},
	}

	return removeCmd
}
