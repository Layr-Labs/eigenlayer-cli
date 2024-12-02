package admin

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/urfave/cli/v2"
)

func AddPendingCmd() *cli.Command {
	addPendingCmd := &cli.Command{
		Name:      "add-pending-admin",
		Usage:     "user admin add-pending-admin <AccountAddress> <AdminAddress>",
		UsageText: "Add an admin to be pending until accepted.",
		Description: `
		Moves a user to pending admin status -- the user must be accepted to be an admin.
		`,
		After: telemetry.AfterRunAction(),
		Flags: []cli.Flag{
			&flags.VerboseFlag,
			&AccountAddressFlag,
			&AdminAddressFlag,
		},
	}

	return addPendingCmd
}
