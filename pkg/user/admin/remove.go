package admin

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/urfave/cli/v2"
)

func RemoveCmd() *cli.Command {
	removeCmd := &cli.Command{
		Name:      "remove-admin",
		Usage:     "user admin remove-admin <AccountAddress> <AdminAddress>",
		UsageText: "Remove a user's admin distinction.",
		Description: `
		The remove command allows you to remove an admin user.
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
