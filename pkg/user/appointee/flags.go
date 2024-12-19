package appointee

import "github.com/urfave/cli/v2"

var (
	AccountAddressFlag = cli.StringFlag{
		Name:    "account-address",
		Aliases: []string{"aa"},
		Usage:   "The Ethereum address of the user. Example: --account-address \"0x...\"",
		EnvVars: []string{"ACCOUNT_ADDRESS"},
	}
	AppointeeAddressFlag = cli.StringFlag{
		Name:    "appointee-address",
		Aliases: []string{"appa"},
		Usage:   "The Ethereum address of the user. Example: --appointee-address \"0x...\"",
		EnvVars: []string{"APPOINTEE_ADDRESS"},
	}
	SelectorFlag = cli.StringFlag{
		Name:    "selector",
		Aliases: []string{"s"},
		Usage:   "The selector for managing permissions to protocol operations. A selector is a smart contract method.",
		EnvVars: []string{"SELECTOR"},
	}
	TargetAddressFlag = cli.StringFlag{
		Name:    "target-address",
		Aliases: []string{"ta"},
		Usage:   "The contract address for managing permissions to protocol operations.",
		EnvVars: []string{"TARGET_ADDRESS"},
	}
	PermissionControllerAddressFlag = cli.StringFlag{
		Name:    "permission-controller-address",
		Aliases: []string{"pca"},
		Usage:   "The Ethereum address of the Permission Manager. Example: --permission-controller-address \"0x...\"",
		EnvVars: []string{"PERMISSION_CONTROLLER_ADDRESS"},
	}
)
