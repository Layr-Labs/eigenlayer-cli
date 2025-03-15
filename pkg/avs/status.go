package avs

import (
	"fmt"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/urfave/cli/v2"
)

func StatusCmd(prompter utils.Prompter) *cli.Command {
	statusCmd := &cli.Command{
		Name:      "status",
		Usage:     "Check the registration status of an operator with an AVS",
		UsageText: "status <operator-config> <avs-config>",
		Description: `
		This command will check the registration status of an operator with an
		AVS.
		`,
		After: telemetry.AfterRunAction(),
		Flags: []cli.Flag{
			&ArgFlag,
			&flags.VerboseFlag,
		},
		Action: func(context *cli.Context) error {
			logger := common.GetLogger(context)
			logger.Info("Starting AVS status workflow")

			coordinator, err := NewFlow(context, prompter)
			if err != nil {
				logger.Errorf("Failed to start workflow: %s", err.Error())
				return failed(err)
			}

			status, err := coordinator.Status()
			if err != nil {
				logger.Errorf("Failed to execute workflow: %s", err.Error())
				return failed(err)
			}

			logger.Info("Completed AVS status workflow")

			switch status {
			case 0:
				fmt.Println("0: not registered")
			case 1:
				fmt.Println("1: registered")
			case 2:
				fmt.Println("2: deregistered")
			default:
				fmt.Printf("%d: unknown\n", status)
			}

			return nil
		},
	}

	return statusCmd
}
