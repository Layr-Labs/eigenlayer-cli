package pkg

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/user/admin"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/user/appointee"
	"github.com/urfave/cli/v2"
)

func UsersCmd() *cli.Command {
	var userCmd = &cli.Command{
		Name:  "user",
		Usage: "Manage user permissions",
		Subcommands: []*cli.Command{
			admin.AdminCmd(),
			appointee.AppointeeCmd(),
		},
	}

	return userCmd
}
