package avs

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/urfave/cli/v2"
)

func DeregisterCmd(prompter utils.Prompter) *cli.Command {
	deregisterCmd := &cli.Command{
		Name:      "deregister",
		Usage:     "Deregister an operator with an AVS",
		UsageText: "deregister <operator-config-file> <avs-config-file>",
		Description: `
		This command will deregister an operator with an AVS.
		`,
		After: telemetry.AfterRunAction(),
		Flags: []cli.Flag{
			&ArgFlag,
			&flags.DryRunFlag,
			&flags.VerboseFlag,
		},
		Action: func(context *cli.Context) error {
			logger := common.GetLogger(context)
			logger.Info("Starting AVS deregistration workflow")

			coordinator, err := NewFlow(context, prompter)
			if err != nil {
				logger.Errorf("Failed to start workflow: %s", err.Error())
				return failed(err)
			}

			status, err := coordinator.Status()
			if err == nil && status != 1 {
				return failed("operator not registered")
			}

			if err := coordinator.Deregister(); err != nil {
				logger.Errorf("Failed to execute workflow: %s", err.Error())
				return failed(err)
			}

			logger.Info("Completed AVS deregistration workflow")
			return nil
		},
	}

	return deregisterCmd
}
