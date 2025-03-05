package command

import (
	"sort"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"

	"github.com/urfave/cli/v2"
)

func NewWriteableCallDataCommand(
	baseCmd BaseCommand,
	name string,
	usage string,
	usageText string,
	description string,
	commandFlags []cli.Flag,
) *cli.Command {
	withWriteFlags := append(commandFlags, flags.WriteFlags...)
	allFlags := append(withWriteFlags, flags.GetSignerFlags()...)
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
