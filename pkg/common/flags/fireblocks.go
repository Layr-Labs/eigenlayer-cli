package flags

import "github.com/urfave/cli/v2"

var (
	// FireblocksAPIKeyFlag is the flag to set the Fireblocks API key
	FireblocksAPIKeyFlag = cli.StringFlag{
		Name:    "fireblocks-api-key",
		Aliases: []string{"ff"},
		Usage:   "Fireblocks API key",
		EnvVars: []string{"FIREBLOCKS_API_KEY"},
	}

	// FireblocksSecretKeyFlag is the flag to set the Fireblocks secret key
	FireblocksSecretKeyFlag = cli.StringFlag{
		Name:    "fireblocks-secret-key",
		Aliases: []string{"fs"},
		Usage:   "Fireblocks secret key. If you are using AWS Secret Manager, this should be the secret name.",
		EnvVars: []string{"FIREBLOCKS_SECRET_KEY"},
	}

	// FireblocksBaseUrlFlag is the flag to set the Fireblocks base URL
	FireblocksBaseUrlFlag = cli.StringFlag{
		Name:    "fireblocks-base-url",
		Aliases: []string{"fb"},
		Usage:   "Fireblocks base URL",
		EnvVars: []string{"FIREBLOCKS_BASE_URL"},
	}

	// FireblocksVaultAccountNameFlag is the flag to set the Fireblocks vault account name
	FireblocksVaultAccountNameFlag = cli.StringFlag{
		Name:    "fireblocks-vault-account-name",
		Aliases: []string{"fv"},
		Usage:   "Fireblocks vault account name",
		EnvVars: []string{"FIREBLOCKS_VAULT_ACCOUNT_NAME"},
	}

	// FireblocksAWSRegionFlag is the flag to set the Fireblocks AWS region
	FireblocksAWSRegionFlag = cli.StringFlag{
		Name:    "fireblocks-aws-region",
		Aliases: []string{"fa"},
		Usage:   "AWS region if secret is stored in AWS KMS",
		EnvVars: []string{"FIREBLOCKS_AWS_REGION"},
		Value:   "us-east-1",
	}

	// FireblocksTimeoutFlag is the flag to set the Fireblocks timeout
	FireblocksTimeoutFlag = cli.Int64Flag{
		Name:    "fireblocks-timeout",
		Aliases: []string{"ft"},
		Usage:   "Fireblocks timeout",
		EnvVars: []string{"FIREBLOCKS_TIMEOUT"},
		Value:   30,
	}

	// FireblocksSecretStorageTypeFlag is the flag to set the Fireblocks secret storage type
	FireblocksSecretStorageTypeFlag = cli.StringFlag{
		Name:    "fireblocks-secret-storage-type",
		Aliases: []string{"fst"},
		Usage:   "Fireblocks secret storage type. Supported values are 'plaintext' and 'aws_secret_manager'",
		EnvVars: []string{"FIREBLOCKS_SECRET_STORAGE_TYPE"},
	}
)
