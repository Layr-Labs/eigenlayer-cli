package admin

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	"github.com/urfave/cli/v2"
)

func AddCmd(p utils.Prompter) *cli.Command {
	addAdminCmd := &cli.Command{
		Name:      "addAdmin",
		Usage:     "Add an Admin",
		UsageText: "",
		Description: `
		addAdmin command gives a new user admin privileges.
		`,
		After: telemetry.AfterRunAction(),
		Action: func(cCtx *cli.Context) error {
			return Add(cCtx, p)
		},
		Flags: []cli.Flag{
			&flags.VerboseFlag,
		},
	}

	return addAdminCmd
}

func Add(cCtx *cli.Context, p utils.Prompter) error {

}
