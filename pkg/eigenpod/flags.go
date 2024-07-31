package eigenpod

import "github.com/urfave/cli/v2"

var (
	PodAddressFlag = cli.StringFlag{
		Name:     "pod-address",
		Usage:    "Specify the address of the EigenPod",
		Required: true,
		EnvVars:  []string{"POD_ADDRESS"},
	}
)
