package container

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/command"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/urfave/cli/v2"
	"sort"
)

func NewVerifiableContainerCommand(
	baseCmd command.BaseCommand,
	name string,
	usage string,
	usageText string,
	description string,
	commandFlags []cli.Flag,
) *cli.Command {
	withContainerFlags := append(commandFlags, &containerDigestFlag)
	sort.Sort(cli.FlagsByName(withContainerFlags))

	c := &cli.Command{
		Name:  name,
		Usage: usage,
		Flags: withContainerFlags,
		Action: func(cCtx *cli.Context) error {
			return baseCmd.Execute(cCtx)
		},
		After: telemetry.AfterRunAction(),
	}

	if usageText != "" {
		c.UsageText = usageText
	}

	if description != "" {
		c.Description = description
	}

	return c
}
