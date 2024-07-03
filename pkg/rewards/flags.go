package rewards

import "github.com/urfave/cli/v2"

var (
	TokenAddressesFlag = cli.StringFlag{
		Name:     "token-addresses",
		Aliases:  []string{"t"},
		Usage:    "Use this flag to specify the addresses of the tokens to claim. Comma separated list of addresses",
		EnvVars:  []string{"TOKEN_ADDRESSES"},
		Required: true,
	}

	RewardsCoordinatorAddressFlag = cli.StringFlag{
		Name:    "rewards-coordinator-address",
		Aliases: []string{"r"},
		Usage:   "Use this flag to specify the address of the rewards coordinator. If not provided, the address will be used based on provided network",
		EnvVars: []string{"REWARDS_COORDINATOR_ADDRESS"},
	}

	ClaimTimestampFlag = cli.StringFlag{
		Name:        "claim-timestamp",
		Aliases:     []string{"c"},
		Usage:       "Use this flag to specify the timestamp. Only 'latest' is supported",
		DefaultText: "latest",
		EnvVars:     []string{"CLAIM_TIMESTAMP"},
	}

	RecipientAddressFlag = cli.StringFlag{
		Name:    "recipient-address",
		Aliases: []string{"ra"},
		Usage:   "Use this flag to specify the address of the recipient. If this is not provided, the earner address will be used",
		EnvVars: []string{"RECIPIENT_ADDRESS"},
	}
)
