package command

import (
	"sort"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"

	"github.com/urfave/cli/v2"
)

func NewServiceCommand(
	baseCmd BaseCommand,
	name string,
	usage string,
	usageText string,
	description string,
	commandFlags []cli.Flag,
) *cli.Command {
	allFlags := append(commandFlags, getServiceFlags()...)
	sort.Sort(cli.FlagsByName(allFlags))

	command := &cli.Command{
		Name:  name,
		Usage: usage,
		Flags: allFlags,
		Action: func(cCtx *cli.Context) error {
			return baseCmd.Execute(cCtx)
		},
		After: telemetry.AfterRunAction(),
	}

	if usageText != "" {
		command.UsageText = usageText
	}

	if description != "" {
		command.Description = description
	}

	return command
}

func getServiceFlags() []cli.Flag {
	return []cli.Flag{
		&flags.NetworkFlag,
		&flags.EnvironmentFlag,
	}
}
