package appointee

import (
	"sort"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/urfave/cli/v2"
)

func RemoveCmd() *cli.Command {
	removeCmd := &cli.Command{
		Name:      "remove",
		Usage:     "user appointee remove --account-address <AccountAddress> --appointee-address <AppointeeAddress> --target-address <TargetAddress> --selector <Selector>",
		UsageText: "Remove a user's permission",
		Description: `
		Remove a user's permission'.
		`,
		After: telemetry.AfterRunAction(),
		Flags: removeCommandFlags(),
	}

	return removeCmd
}

func removeCommandFlags() []cli.Flag {
	cmdFlags := []cli.Flag{
		&flags.VerboseFlag,
		&AccountAddressFlag,
		&AppointeeAddressFlag,
		&TargetAddressFlag,
		&SelectorFlag,
	}
	sort.Sort(cli.FlagsByName(cmdFlags))
	return append(cmdFlags, flags.GetSignerFlags()...)
}
