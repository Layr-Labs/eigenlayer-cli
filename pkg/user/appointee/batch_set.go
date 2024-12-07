package appointee

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/urfave/cli/v2"
)

func BatchSetCmd() *cli.Command {
	batchSetCmd := &cli.Command{
		Name:      "batch set",
		Usage:     "user appointee batch-set --batch-set-file <BatchSetFile>",
		UsageText: "Appoint multiple users permissions at a time.",
		Description: `
		Appoint multiple users permissions at a time.
		`,
		After: telemetry.AfterRunAction(),
		Flags: batchSetFlags(),
	}

	return batchSetCmd
}

func batchSetFlags() []cli.Flag {
	cmdFlags := []cli.Flag{
		&flags.VerboseFlag,
		&BatchSetFileFlag,
	}
	return append(cmdFlags, flags.GetSignerFlags()...)
}
