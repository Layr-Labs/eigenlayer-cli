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
		Manage admin users.
		`,
		After: telemetry.AfterRunAction(),
		Flags: []cli.Flag{
			&flags.VerboseFlag,
		},
		Subcommands: []*cli.Command{
			AcceptCmd(),
			AddPendingCmd(),
			IsAdminCmd(generateIsAdminReader),
			IsPendingCmd(generateIsPendingAdminReader),
			ListCmd(generateListAdminsReader),
			ListPendingCmd(generateListPendingAdminsReader),
			RemoveCmd(),
			RemovePendingCmd(),
		},
	}

	return adminCmd
}
