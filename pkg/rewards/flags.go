package rewards

import "github.com/urfave/cli/v2"

var (
	TokenAddressesFlag = cli.StringFlag{
		Name:     "token-addresses",
		Aliases:  []string{"t"},
		Usage:    "Specify the addresses of the tokens to claim. Comma separated list of addresses",
		EnvVars:  []string{"TOKEN_ADDRESSES"},
		Required: true,
	}

	RewardsCoordinatorAddressFlag = cli.StringFlag{
		Name:    "rewards-coordinator-address",
		Aliases: []string{"rc"},
		Usage:   "Specify the address of the rewards coordinator. If not provided, the address will be used based on provided network",
		EnvVars: []string{"REWARDS_COORDINATOR_ADDRESS"},
	}

	ClaimTimestampFlag = cli.StringFlag{
		Name:    "claim-timestamp",
		Aliases: []string{"c"},
		Usage:   "Specify the timestamp. Only 'latest' is supported",
		Value:   "latest",
		EnvVars: []string{"CLAIM_TIMESTAMP"},
	}

	RecipientAddressFlag = cli.StringFlag{
		Name:    "recipient-address",
		Aliases: []string{"ra"},
		Usage:   "Specify the address of the recipient. If this is not provided, the earner address will be used",
		EnvVars: []string{"RECIPIENT_ADDRESS"},
	}

	ProofStoreBaseURLFlag = cli.StringFlag{
		Name:    "proof-store-base-url",
		Aliases: []string{"psbu"},
		Usage:   "Specify the base URL of the proof store. If not provided, the value based on network will be used",
		EnvVars: []string{"PROOF_STORE_BASE_URL"},
	}

	EnvironmentFlag = cli.StringFlag{
		Name:    "environment",
		Aliases: []string{"env"},
		Usage:   "Environment to use. Currently supports 'preprod' ,`testnet' and 'prod'. If not provided, it will be inferred based on network",
		EnvVars: []string{"ENVIRONMENT"},
	}

	ClaimerAddressFlag = cli.StringFlag{
		Name:     "claimer-address",
		Aliases:  []string{"a"},
		Usage:    "Address of the claimer",
		Required: true,
		EnvVars:  []string{"NODE_OPERATOR_CLAIMER_ADDRESS"},
	}
)
