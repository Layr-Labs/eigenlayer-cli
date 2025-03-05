package avs

import (
	"fmt"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/urfave/cli/v2"
)

func SpecsCmd() *cli.Command {
	specsCmd := &cli.Command{
		Name:  "specs",
		Usage: "Manage AVS specifications",
		Subcommands: []*cli.Command{
			ListCmd(),
			ResetCmd(),
			UpdateCmd(),
		},
	}

	return specsCmd
}

func ListCmd() *cli.Command {
	listCmd := &cli.Command{
		Name:      "list",
		Usage:     "List available AVS specifications",
		UsageText: "list",
		Description: `
		This command will list the specifications available in the local
		repository stored in $HOME/.eigenlayer/avs/specs.
		
		Use the 'avs specs update' command to download the latest set of
		specifications and update the local repository.
		`,
		After: telemetry.AfterRunAction(),
		Flags: []cli.Flag{
			&flags.VerboseFlag,
		},
		Action: func(context *cli.Context) error {
			logger := common.GetLogger(context)
			repo, err := NewRepository(context)
			if err != nil {
				logger.Errorf("Failed to initialize repository: %s", err.Error())
				return failed(err)
			}

			specifications, err := repo.List()
			if err != nil {
				logger.Errorf("Failed to list specifications: %s", err.Error())
				return failed(err)
			}

			for _, specification := range *specifications {
				fmt.Printf("%-30s%s\n", specification.Name, specification.Description)
			}

			return nil
		},
	}

	return listCmd
}

func ResetCmd() *cli.Command {
	resetCmd := &cli.Command{
		Name:      "reset",
		Usage:     "Reset local specification repository",
		UsageText: "reset",
		Description: `
		This command will reset the local specification repository stored in
		$HOME/.eigenlayer/avs/specs. All existing content will be removed and
		the repository will be repopulated with the default set of
		specifications that were available at compile time.
		`,
		After: telemetry.AfterRunAction(),
		Flags: []cli.Flag{
			&flags.VerboseFlag,
		},
		Action: func(context *cli.Context) error {
			logger := common.GetLogger(context)
			repo, err := NewRepository(context)
			if err != nil {
				logger.Errorf("Failed to initialize repository: %s", err.Error())
				return failed(err)
			}

			return repo.Reset()
		},
	}

	return resetCmd
}

func UpdateCmd() *cli.Command {
	updateCmd := &cli.Command{
		Name:      "update",
		Usage:     "Download the latest set of specifications",
		UsageText: "update",
		Description: `
		This command will update the local specification repository stored in
		$HOME/.eigenlayer/avs/specs by downloading the latest set of
		specifications form the official remote repository.
		`,
		After: telemetry.AfterRunAction(),
		Flags: []cli.Flag{
			&flags.VerboseFlag,
		},
		Action: func(context *cli.Context) error {
			logger := common.GetLogger(context)
			repo, err := NewRepository(context)
			if err != nil {
				logger.Errorf("Failed to initialize repository: %s", err.Error())
				return failed(err)
			}

			if err := repo.Update(); err != nil {
				logger.Errorf("Failed to update repository: %s", err.Error())
				return failed(err)
			}

			return nil
		},
	}

	return updateCmd
}
