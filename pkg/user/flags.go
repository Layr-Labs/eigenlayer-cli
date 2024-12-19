package user

import "github.com/urfave/cli/v2"

var (
	CallerAddressFlag = cli.StringFlag{
		Name:    "caller-address",
		Aliases: []string{"ca"},
		Usage: "This is the address of the caller who is calling the contract function. \n" +
			"If it is not provided, the account address will be used as the caller address",
		EnvVars: []string{"CALLER_ADDRESS"},
	}
)
