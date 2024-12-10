package admin

import (
	"sort"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/urfave/cli/v2"
)

func RemovePendingCmd() *cli.Command {
	removeCmd := &cli.Command{
		Name:      "remove-pending-admin",
		Usage:     "user admin remove-pending-admin --account-address <AccountAddress> --admin-address <AdminAddress>",
		UsageText: "Remove a user who is pending admin acceptance.",
		Description: `
		Remove a user who is pending admin acceptance.
		`,
		After: telemetry.AfterRunAction(),
		Flags: removePendingFlags(),
	}

	return removeCmd
}

func removePendingFlags() []cli.Flag {
	cmdFlags := []cli.Flag{
		&flags.VerboseFlag,
		&AccountAddressFlag,
		&AdminAddressFlag,
	}
	sort.Sort(cli.FlagsByName(cmdFlags))
	return append(cmdFlags, flags.GetSignerFlags()...)
}
