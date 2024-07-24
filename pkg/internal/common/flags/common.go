package flags

import "github.com/urfave/cli/v2"

func GetSignerFlags() []cli.Flag {
	return []cli.Flag{
		&EcdsaPrivateKeyFlag,
		&PathToKeyStoreFlag,
		&FireblocksAPIKeyFlag,
		&FireblocksSecretKeyFlag,
		&FireblocksBaseUrlFlag,
		&FireblocksVaultAccountNameFlag,
		&FireblocksAWSRegionFlag,
		&FireblocksTimeoutFlag,
		&FireblocksSecretStorageTypeFlag,
		&Web3SignerUrlFlag,
	}
}
