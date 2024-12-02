package appointee

import "github.com/urfave/cli/v2"

var (
	AccountAddressFlag = cli.StringFlag{
		Name:    "account-address",
		Aliases: []string{"aa"},
		Usage:   "The Ethereum address of the user. Example: --account-address \"0x...\"",
		EnvVars: []string{"ACCOUNT_ADDRESS"},
	}
)

var (
	AppointeeAddressFlag = cli.StringFlag{
		Name:    "appointee-address",
		Aliases: []string{"appa"},
		Usage:   "The Ethereum address of the user. Example: --appointee-address \"0x...\"",
		EnvVars: []string{"APPOINTEE_ADDRESS"},
	}
)

var (
	CallerAddressFlag = cli.StringFlag{
		Name:    "caller-address",
		Aliases: []string{"ca"},
		Usage:   "The Ethereum address of the caller. Example: --caller-address \"0x...\"",
		EnvVars: []string{"CALLER_ADDRESS"},
	}
)

var (
	SelectorFlag = cli.StringFlag{
		Name:    "selector",
		Aliases: []string{"sa"},
		Usage:   "The selector for managing permissions to protocol operations. A selector is a smart contract method.",
		EnvVars: []string{"SELECTOR"},
	}
)

var (
	TargetAddressFlag = cli.StringFlag{
		Name:    "target-address",
		Aliases: []string{"ta"},
		Usage:   "The contract address for managing permissions to protocol operations.",
		EnvVars: []string{"TARGET_ADDRESS"},
	}
)

var (
	BatchSetFileFlag = cli.StringFlag{
		Name:    "batch-set-file",
		Aliases: []string{"bsf"},
		Usage:   "A YAML file for executing a batch of set permission operations.",
		EnvVars: []string{"BATCH_SET_FILE"},
	}
)

var (
	PermissionManagerAddressFlag = cli.StringFlag{
		Name:    "permission-manager-address",
		Aliases: []string{"pma"},
		Usage:   "The Ethereum address of the Permission Manager. Example: --permission-manager-address \"0x...\"",
		EnvVars: []string{"PERMISSION_MANAGER_ADDRESS"},
	}
)
