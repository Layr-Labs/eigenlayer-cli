package admin

import (
	"sort"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/urfave/cli/v2"
)

func RemoveCmd() *cli.Command {
	removeCmd := &cli.Command{
		Name:      "remove-admin",
		Usage:     "user admin remove-admin --account-address <AccountAddress> --admin-address <AdminAddress>",
		UsageText: "The remove command allows you to remove an admin user.",
		Description: `
		The remove command allows you to remove an admin user.
		`,
		After: telemetry.AfterRunAction(),
		Flags: removeFlags(),
	}

	return removeCmd
}

func removeFlags() []cli.Flag {
	cmdFlags := []cli.Flag{
		&flags.VerboseFlag,
		&AccountAddressFlag,
		&AdminAddressFlag,
	}
	sort.Sort(cli.FlagsByName(cmdFlags))
	return append(cmdFlags, flags.GetSignerFlags()...)
}
