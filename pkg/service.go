package pkg

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/service"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/urfave/cli/v2"
)

func NewServiceCmd(prompter utils.Prompter) *cli.Command {
	var verifiableCmd = &cli.Command{
		Name:  "service",
		Usage: "Eigen hosted service operations.",
		Subcommands: []*cli.Command{
			service.NewReleaseManagementServiceCmd(prompter),
		},
	}

	return verifiableCmd
}
