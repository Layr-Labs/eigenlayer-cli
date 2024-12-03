package appointee

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/urfave/cli/v2"
)

func ListPermissionsCmd() *cli.Command {
	listPermissions := &cli.Command{
		Name:      "list-permissions",
		Usage:     "user appointee list-permissions <AccountAddress> <AppointeeAddress>",
		UsageText: "List all permissions for a user.",
		Description: `
		List all permissions of a user.
		`,
		After: telemetry.AfterRunAction(),
		Flags: []cli.Flag{
			&flags.VerboseFlag,
			&AccountAddressFlag,
			&AppointeeAddressFlag,
		},
	}

	return listPermissions
}
