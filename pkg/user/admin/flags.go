package admin

import "github.com/urfave/cli/v2"

var (
	AccountAddressFlag = cli.StringFlag{
		Name:    "account-address",
		Aliases: []string{"aa"},
		Usage:   "user admin ... --account-address \"0x...\"",
		EnvVars: []string{"ACCOUNT_ADDRESS"},
	}
)

var (
	AdminAddressFlag = cli.StringFlag{
		Name:    "admin-address",
		Aliases: []string{"aa"},
		Usage:   "user admin ... --admin-address \"0x...\"",
		EnvVars: []string{"ADMIN_ADDRESS"},
	}
)

var (
	CallerAddressFlag = cli.StringFlag{
		Name:    "caller-address",
		Aliases: []string{"ca"},
		Usage:   "user admin ... --caller-address \"0x...\"",
		EnvVars: []string{"CALLER_ADDRESS"},
	}
)

var (
	PendingAdminAddressFlag = cli.StringFlag{
		Name:    "pending-admin-address",
		Aliases: []string{"paa"},
		Usage:   "user admin ... --pending-admin-address \"0x...\"",
		EnvVars: []string{"PENDING_ADMIN_ADDRESS"},
	}
)

var (
	PermissionManagerAddressFlag = cli.StringFlag{
		Name:    "permission-manager-address",
		Aliases: []string{"pma"},
		Usage:   "user admin ... --permission-manager-address \"0x...\"",
		EnvVars: []string{"PERMISSION_MANAGER_ADDRESS"},
	}
)
