package admin

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/urfave/cli/v2"
)

func AdminCmd() *cli.Command {
	adminCmd := &cli.Command{
		Name:      "admin",
		Usage:     "user admin <command>",
		UsageText: "Manage admin users.",
		Description: `
		The admin command allows you to manage admin users.
		`,
		After: telemetry.AfterRunAction(),
		Flags: []cli.Flag{
			&flags.VerboseFlag,
		},
		Subcommands: []*cli.Command{
			AcceptCmd(),
			AddPendingCmd(),
			IsAdminCmd(),
			IsPendingCmd(),
			ListCmd(),
			ListPendingCmd(),
			RemoveCmd(),
			RemovePendingCmd(),
		},
	}

	return adminCmd
}
