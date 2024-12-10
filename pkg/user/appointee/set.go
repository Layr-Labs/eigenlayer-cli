package appointee

import (
	"sort"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/urfave/cli/v2"
)

func SetCmd() *cli.Command {
	setCmd := &cli.Command{
		Name:      "set",
		Usage:     "user appointee set --account-address <AccountAddress> --appointee-address <AppointeeAddress> --target-address <TargetAddress> --selector <Selector>",
		UsageText: "Grant a user a permission.",
		Description: `
		Grant a user a permission.'.
		`,
		After: telemetry.AfterRunAction(),
		Flags: setCommandFlags(),
	}

	return setCmd
}

func setCommandFlags() []cli.Flag {
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
