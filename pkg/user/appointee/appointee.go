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
		UsageText: "User permission management operations.",
		Description: `
		User permission management operations.
		`,
		After: telemetry.AfterRunAction(),
		Flags: []cli.Flag{
			&flags.VerboseFlag,
		},
		Subcommands: []*cli.Command{
			BatchSetCmd(),
			canCallCmd(generateUserCanCallReader),
			ListCmd(),
			ListPermissionsCmd(),
			RemoveCmd(),
			SetCmd(),
		},
	}

	return appointeeCmd
}
