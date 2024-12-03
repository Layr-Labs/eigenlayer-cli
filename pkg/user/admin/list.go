package admin

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/urfave/cli/v2"
)

func ListCmd() *cli.Command {
	listCmd := &cli.Command{
		Name:      "list-admins",
		Usage:     "user admin list-admins <AccountAddress>",
		UsageText: "List all users who are admins.",
		Description: `
		List all users who are admins.
		`,
		After: telemetry.AfterRunAction(),
		Flags: []cli.Flag{
			&flags.VerboseFlag,
			&AccountAddressFlag,
		},
	}

	return listCmd
}
