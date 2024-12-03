package admin

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/urfave/cli/v2"
)

func IsAdminCmd() *cli.Command {
	isAdmin := &cli.Command{
		Name:      "is-admin",
		Usage:     "user admin is-admin <AccountAddress> <CallerAddress>",
		UsageText: "Checks if a user is an admin.",
		Description: `
		Checks if a user is an admin.
		`,
		After: telemetry.AfterRunAction(),
		Flags: []cli.Flag{
			&flags.VerboseFlag,
			&AccountAddressFlag,
			&CallerAddressFlag,
		},
	}

	return isAdmin
}
