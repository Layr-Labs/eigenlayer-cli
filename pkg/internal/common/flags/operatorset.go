package flags

import "github.com/urfave/cli/v2"

var (
	OperatorSetIdsFlag = cli.Uint64SliceFlag{
		Name:     "operator-set-ids",
		Usage:    "The IDs of the operator sets to deregister. Comma separated list of operator set ids",
		Required: true,
		EnvVars:  []string{"OPERATOR_SET_IDS"},
		Aliases:  []string{"opsids"},
	}
)
