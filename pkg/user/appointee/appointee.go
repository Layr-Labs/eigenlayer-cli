package appointee

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/urfave/cli/v2"
)

func AppointeeCmd() *cli.Command {
	appointeeCmd := &cli.Command{
		Name:      "appointee",
		Usage:     "user appointee <command>",
		UsageText: "Manage user permissions.",
		Description: `
		The appointee command allows you to manage user permissions.
		`,
		After: telemetry.AfterRunAction(),
		Flags: []cli.Flag{
			&flags.VerboseFlag,
		},
		Subcommands: []*cli.Command{
			BatchSetCmd(),
			CanCallCmd(),
			ListCmd(),
			ListPermissionsCmd(),
			RemoveCmd(),
			SetCmd(),
		},
	}

	return appointeeCmd
}
