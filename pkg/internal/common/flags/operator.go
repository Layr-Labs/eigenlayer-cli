package flags

import "github.com/urfave/cli/v2"

var (
	OperatorAddressFlag = cli.StringFlag{
		Name:    "operator-address",
		Usage:   "The address of the operator",
		EnvVars: []string{"OPERATOR_ADDRESS"},
		Aliases: []string{"oa"},
	}
)
