package flags

import "github.com/urfave/cli/v2"

var (
	Web3SignerUrlFlag = cli.StringFlag{
		Name:    "web3signer-url",
		Aliases: []string{"w"},
		Usage:   "URL of the Web3Signer",
		EnvVars: []string{"WEB3SIGNER_URL"},
	}
)
