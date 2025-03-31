package service

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/service/release"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/urfave/cli/v2"
)

func NewReleaseManagementServiceCmd(prompter utils.Prompter) *cli.Command {
	var rmsCmd = &cli.Command{
		Name:  "rms",
		Usage: "Release Management Service operations.",
		Subcommands: []*cli.Command{
			release.NewListOperatorReleasesCmd(prompter),
			release.NewListAvsReleaseKeysCmd(prompter),
		},
	}

	return rmsCmd
}
