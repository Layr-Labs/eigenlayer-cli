package admin

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/urfave/cli/v2"
)

func AcceptCmd() *cli.Command {
	acceptCmd := &cli.Command{
		Name:      "accept-admin",
		Usage:     "user admin accept-admin <AccountAddress>",
		UsageText: "Accepts a user to become admin who is currently pending admin acceptance.",
		Description: `
		Accepts a user to become admin who is currently pending admin acceptance.
		`,
		After: telemetry.AfterRunAction(),
		Flags: []cli.Flag{
			&flags.VerboseFlag,
			&AccountAddressFlag,
		},
	}

	return acceptCmd
}
