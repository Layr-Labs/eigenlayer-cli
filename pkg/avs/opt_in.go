package avs

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/urfave/cli/v2"
)

func OptInCmd(prompter utils.Prompter) *cli.Command {
	optInCmd := &cli.Command{
		Name:      "opt-in",
		Usage:     "Opt-in an operator with an AVS",
		UsageText: "opt-in <operator-config-file> <avs-config-file>",
		Description: `
		This command will opt-in an operator with an AVS.
		`,
		After: telemetry.AfterRunAction(),
		Flags: []cli.Flag{
			&ArgFlag,
			&flags.DryRunFlag,
			&flags.VerboseFlag,
		},
		Action: func(context *cli.Context) error {
			logger := common.GetLogger(context)
			logger.Info("Starting AVS opt-in workflow")

			coordinator, err := NewFlow(context, prompter)
			if err != nil {
				logger.Errorf("Failed to start workflow: %s", err.Error())
				return failed(err)
			}

			status, err := coordinator.Status()
			if err == nil && status != 1 {
				return failed("operator not registered")
			}

			if err := coordinator.OptIn(); err != nil {
				logger.Errorf("Failed to execute workflow: %s", err.Error())
				return failed(err)
			}

			logger.Info("Completed AVS opt-in workflow")
			return nil
		},
	}

	return optInCmd
}
