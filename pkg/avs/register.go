package avs

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/urfave/cli/v2"
)

func RegisterCmd(prompter utils.Prompter) *cli.Command {
	registerCmd := &cli.Command{
		Name:      "register",
		Usage:     "Register an operator with an AVS",
		UsageText: "register <operator-config-file> <avs-config-file>",
		Description: `
		This command will register an operator with an AVS.
		`,
		After: telemetry.AfterRunAction(),
		Flags: []cli.Flag{
			&ArgFlag,
			&flags.DryRunFlag,
			&flags.VerboseFlag,
		},
		Action: func(context *cli.Context) error {
			logger := common.GetLogger(context)
			logger.Info("Starting AVS registration workflow")

			coordinator, err := NewFlow(context, prompter)
			if err != nil {
				logger.Errorf("Failed to start workflow: %s", err.Error())
				return failed(err)
			}

			status, err := coordinator.Status()
			if err == nil && status == 1 {
				return failed("operator already registered")
			}

			if err := coordinator.Register(); err != nil {
				logger.Errorf("Failed to execute workflow: %s", err.Error())
				return failed(err)
			}

			logger.Info("Completed AVS registration workflow")
			return nil
		},
	}

	return registerCmd
}
