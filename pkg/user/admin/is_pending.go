package admin

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/urfave/cli/v2"
)

func IsPendingCmd() *cli.Command {
	isPendingCmd := &cli.Command{
		Name:      "is-pending-admin",
		Usage:     "user admin is-pending-admin --account-address <AccountAddress> --pending-admin-address <PendingAdminAddress>",
		UsageText: "Checks if a user is pending acceptance to admin.",
		Description: `
		Checks if a user is pending acceptance to admin.
		`,
		After: telemetry.AfterRunAction(),
		Flags: []cli.Flag{
			&flags.VerboseFlag,
			&AccountAddressFlag,
			&PendingAdminAddressFlag,
		},
	}

	return isPendingCmd
}
