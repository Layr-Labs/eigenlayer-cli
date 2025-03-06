package avs

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/urfave/cli/v2"
)

func OptOutCmd(prompter utils.Prompter) *cli.Command {
	optOutCmd := &cli.Command{
		Name:      "opt-out",
		Usage:     "Opt-out an operator with an AVS",
		UsageText: "opt-out <operator-config-file> <avs-config-file>",
		Description: `
		This command will opt-out an operator with an AVS.
		`,
		After: telemetry.AfterRunAction(),
		Flags: []cli.Flag{
			&ArgFlag,
			&flags.DryRunFlag,
			&flags.VerboseFlag,
		},
		Action: func(context *cli.Context) error {
			logger := common.GetLogger(context)
			logger.Info("Starting AVS opt-out workflow")

			coordinator, err := NewFlow(context, prompter)
			if err != nil {
				logger.Errorf("Failed to start workflow: %s", err.Error())
				return failed(err)
			}

			status, err := coordinator.Status()
			if err == nil && status != 1 {
				return failed("operator not registered")
			}

			if err := coordinator.OptOut(); err != nil {
				logger.Errorf("Failed to execute workflow: %s", err.Error())
				return failed(err)
			}

			logger.Info("Completed AVS opt-out workflow")
			return nil
		},
	}

	return optOutCmd
}
