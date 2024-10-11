package rewards

import "github.com/urfave/cli/v2"

var (
	TokenAddressesFlag = cli.StringFlag{
		Name:     "token-addresses",
		Aliases:  []string{"t"},
		Usage:    "Specify the addresses of the tokens to claim. Comma separated list of addresses. Omit to claim all rewards.",
		EnvVars:  []string{"TOKEN_ADDRESSES"},
		Required: false,
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
		Usage:   "Specify the timestamp. Only 'latest' and 'latest_active' are supported. 'latest' can be an inactive root which you can't claim yet.",
		Value:   "latest_active",
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
		Usage:   "Environment to use. Currently supports 'preprod' ,'testnet' and 'prod'. If not provided, it will be inferred based on network",
		EnvVars: []string{"ENVIRONMENT"},
	}

	ClaimerAddressFlag = cli.StringFlag{
		Name:     "claimer-address",
		Aliases:  []string{"a"},
		Usage:    "Address of the claimer",
		Required: false,
		EnvVars:  []string{"REWARDS_CLAIMER_ADDRESS"},
	}

	EarnerAddressFlag = cli.StringFlag{
		Name:     "earner-address",
		Aliases:  []string{"ea"},
		Usage:    "Address of the earner",
		Required: true,
		EnvVars:  []string{"REWARDS_EARNER_ADDRESS"},
	}

	NumberOfDaysFlag = cli.IntFlag{
		Name:    "number-of-days",
		Aliases: []string{"nd"},
		Usage:   "Number of past days to show rewards for. It should be negative. Only used for 'all' claim type",
		Value:   -21,
		EnvVars: []string{"REWARDS_NUMBER_OF_DAYS"},
		Action: func(context *cli.Context, i int) error {
			if i >= 0 {
				return cli.Exit("Number of days should be negative to represent past days", 1)
			}
			return nil
		},
	}

	AVSAddressesFlag = cli.StringFlag{
		Name:    "avs-addresses",
		Aliases: []string{"a"},
		Usage:   "Comma seperated addresses of the AVS",
		EnvVars: []string{"AVS_ADDRESSES"},
	}

	ClaimTypeFlag = cli.StringFlag{
		Name:    "claim-type",
		Aliases: []string{"ct"},
		Usage:   "Type of claim you want to see. Can be 'all', 'unclaimed', or 'claimed'",
		Value:   "all",
		EnvVars: []string{"REWARDS_CLAIM_TYPE"},
	}

	SilentFlag = cli.BoolFlag{
		Name:    "silent",
		Aliases: []string{"s"},
		Usage:   "Suppress output except for claim",
		EnvVars: []string{"REWARDS_CLAIM_SILENT"},
	}
)
