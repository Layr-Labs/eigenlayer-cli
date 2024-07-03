package rewards

import "github.com/urfave/cli/v2"

var (
	SubmitClaimFlag = cli.BoolFlag{
		Name:    "submit-claim",
		Aliases: []string{"s"},
		Usage:   "Use this flag to submit the claim",
		EnvVars: []string{"SUBMIT_CLAIM"},
	}

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
)
