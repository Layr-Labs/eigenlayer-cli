package flags

import "github.com/urfave/cli/v2"

var (
	NetworkFlag = cli.StringFlag{
		Name:    "network",
		Aliases: []string{"n"},
		Usage:   "Network to use. Currently supports 'holesky' and 'mainnet'",
		Value:   "holesky",
		EnvVars: []string{"NETWORK"},
	}

	ETHRpcUrlFlag = cli.StringFlag{
		Name:     "eth-rpc-url",
		Aliases:  []string{"r"},
		Required: true,
		Usage:    "URL of the Ethereum RPC",
		EnvVars:  []string{"ETH_RPC_URL"},
	}

	BeaconRpcUrlFlag = cli.StringFlag{
		Name:    "beacon-rpc-url",
		Aliases: []string{"b"},
		Usage:   "URL of the ETH Beacon RPC",
		EnvVars: []string{"BEACON_RPC_URL"},
	}

	OutputFileFlag = cli.StringFlag{
		Name:    "output-file",
		Aliases: []string{"o"},
		Usage:   "Output file to write the data",
		EnvVars: []string{"OUTPUT_FILE"},
	}

	OutputTypeFlag = cli.StringFlag{
		Name:    "output-type",
		Aliases: []string{"ot"},
		Value:   "pretty",
		Usage:   "Output format of the command. One of 'pretty', 'json' or 'calldata'",
		EnvVars: []string{"OUTPUT_TYPE"},
	}

	PathToKeyStoreFlag = cli.StringFlag{
		Name:    "path-to-key-store",
		Aliases: []string{"k"},
		Usage:   "Path to the key store used to send transactions",
		EnvVars: []string{"PATH_TO_KEY_STORE"},
	}

	BroadcastFlag = cli.BoolFlag{
		Name:    "broadcast",
		Aliases: []string{"b"},
		Usage:   "Use this flag to broadcast the transaction",
		EnvVars: []string{"BROADCAST"},
	}

	DryRunFlag = cli.BoolFlag{
		Name:    "dry-run",
		Aliases: []string{"d"},
		Usage:   "Perform a dry run. This takes precedence over the broadcast flag",
		EnvVars: []string{"DRY_RUN"},
	}

	EcdsaPrivateKeyFlag = cli.StringFlag{
		Name:    "ecdsa-private-key",
		Aliases: []string{"e"},
		Usage:   "ECDSA private key hex to send transaction",
		EnvVars: []string{"ECDSA_PRIVATE_KEY"},
	}

	VerboseFlag = cli.BoolFlag{
		Name:    "verbose",
		Aliases: []string{"v"},
		Usage:   "Enable verbose logging",
		EnvVars: []string{"VERBOSE"},
	}

	SilentFlag = cli.BoolFlag{
		Name:    "silent",
		Aliases: []string{"s"},
		Usage:   "Suppress unnecessary output",
		EnvVars: []string{"SILENT"},
	}

	ExpiryFlag = cli.Int64Flag{
		Name:    "expiry",
		Aliases: []string{"exp"},
		Usage:   "expiry in seconds",
		EnvVars: []string{"EXPIRY"},
		Value:   3600,
	}

	OperatorAddressFlag = cli.StringFlag{
		Name:    "operator-address",
		Aliases: []string{"oa", "operator"},
		Usage:   "Operator address",
		EnvVars: []string{"OPERATOR_ADDRESS"},
	}

	BatchClaimFile = cli.StringFlag{
		Name:    "batch-claim-file",
		Aliases: []string{"bcf"},
		Usage:   "Input file for batch rewards claim",
		EnvVars: []string{"BATCH_CLAIM_FILE"},
	}

	CSVFileFlag = cli.StringFlag{
		Name:    "csv-file",
		Aliases: []string{"csv"},
		Usage:   "CSV file to read data from",
		EnvVars: []string{"CSV_FILE"},
	}

	EnvironmentFlag = cli.StringFlag{
		Name:    "environment",
		Aliases: []string{"env"},
		Usage:   "environment to use. Currently supports 'preprod' ,'testnet' and 'prod'. If not provided, it will be inferred based on network",
		EnvVars: []string{"ENVIRONMENT"},
	}

	DelegationManagerAddressFlag = cli.StringFlag{
		Name:    "delegation-manager-address",
		Aliases: []string{"dma"},
		Usage:   "Optional delegation manager address. This can be used if you are testing against your own deployment of eigenlayer contracts",
		EnvVars: []string{"DELEGATION_MANAGER_ADDRESS"},
	}

	CallerAddressFlag = cli.StringFlag{
		Name:    "caller-address",
		Aliases: []string{"ca"},
		Usage:   "This is the address of the caller who is calling the contract function. If it is not provided, the operator address will be used as the caller address",
		EnvVars: []string{"CALLER_ADDRESS"},
	}
)
