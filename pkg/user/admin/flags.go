package admin

import "github.com/urfave/cli/v2"

var (
	AccountAddressFlag = cli.StringFlag{
		Name:    "account-address",
		Aliases: []string{"aa"},
		Usage:   "user admin ... --account-address \"0x...\"",
		EnvVars: []string{"ACCOUNT_ADDRESS"},
	}
	AdminAddressFlag = cli.StringFlag{
		Name:    "admin-address",
		Aliases: []string{"ada"},
		Usage:   "user admin ... --admin-address \"0x...\"",
		EnvVars: []string{"ADMIN_ADDRESS"},
	}
	CallerAddressFlag = cli.StringFlag{
		Name:    "caller-address",
		Aliases: []string{"ca"},
		Usage: "This is the address of the caller who is calling the contract function. \n" +
			"If it is not provided, the account address will be used as the caller address",
		EnvVars: []string{"CALLER_ADDRESS"},
	}
	PendingAdminAddressFlag = cli.StringFlag{
		Name:    "pending-admin-address",
		Aliases: []string{"paa"},
		Usage:   "user admin ... --pending-admin-address \"0x...\"",
		EnvVars: []string{"PENDING_ADMIN_ADDRESS"},
	}
	PermissionControllerAddressFlag = cli.StringFlag{
		Name:    "permission-controller-address",
		Aliases: []string{"pca"},
		Usage:   "user admin ... --permission-controller-address \"0x...\"",
		EnvVars: []string{"PERMISSION_CONTROLLER_ADDRESS"},
	}
)
