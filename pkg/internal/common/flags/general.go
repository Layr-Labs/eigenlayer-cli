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
		Aliases: []string{"e"},
		Usage:   "expiry in seconds",
		EnvVars: []string{"EXPIRY"},
		Value:   3600,
	}
)
