package admin

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/urfave/cli/v2"
)

func AdminCmd(prompter utils.Prompter) *cli.Command {
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
			AcceptCmd(generateAcceptAdminWriter(prompter)),
			AddPendingCmd(generateAddPendingAdminWriter(prompter)),
			IsAdminCmd(generateIsAdminReader),
			IsPendingCmd(generateIsPendingAdminReader),
			ListCmd(generateListAdminsReader),
			ListPendingCmd(generateListPendingAdminsReader),
			RemoveCmd(generateRemoveAdminWriter(prompter)),
			RemovePendingCmd(generateRemovePendingAdminWriter(prompter)),
		},
	}

	return adminCmd
}
