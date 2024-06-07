package types

import (
	"math/big"

	eigensdkTypes "github.com/Layr-Labs/eigensdk-go/types"
)

type SignerType string

type SecretStorageType string

const (
	PrivateKeySigner    SignerType = "private_key"
	LocalKeystoreSigner SignerType = "local_keystore"
	FireBlocksSigner    SignerType = "fireblocks"
	Web3Signer          SignerType = "web3"

	AWSSecretManager SecretStorageType = "aws_secret_manager"
	PlainText        SecretStorageType = "plaintext"
)

type FireblocksConfig struct {
	APIKey           string `yaml:"api_key"`
	SecretKey        string `yaml:"secret_key"`
	BaseUrl          string `yaml:"base_url"`
	VaultAccountName string `yaml:"vault_account_name"`

	SecretStorageType SecretStorageType `yaml:"secret_storage_type"`

	// AWSRegion is the region where the secret is stored in AWS Secret Manager
	AWSRegion string `yaml:"aws_region"`

	// Timeout for API in seconds
	Timeout int64 `yaml:"timeout"`
}

type Web3SignerConfig struct {
	Url string `yaml:"url"`
}

type OperatorConfig struct {
	Operator                   eigensdkTypes.Operator `yaml:"operator"`
	ELDelegationManagerAddress string                 `yaml:"el_delegation_manager_address"`
	ELAVSDirectoryAddress      string
	EthRPCUrl                  string           `yaml:"eth_rpc_url"`
	PrivateKeyStorePath        string           `yaml:"private_key_store_path"`
	SignerType                 SignerType       `yaml:"signer_type"`
	ChainId                    big.Int          `yaml:"chain_id"`
	FireblocksConfig           FireblocksConfig `yaml:"fireblocks"`
	Web3SignerConfig           Web3SignerConfig `yaml:"web3"`
}

func (config OperatorConfig) MarshalYAML() (interface{}, error) {
	return struct {
		Operator                   eigensdkTypes.Operator `yaml:"operator"`
		ELDelegationManagerAddress string                 `yaml:"el_delegation_manager_address"`
		EthRPCUrl                  string                 `yaml:"eth_rpc_url"`
		PrivateKeyStorePath        string                 `yaml:"private_key_store_path"`
		SignerType                 SignerType             `yaml:"signer_type"`
		ChainID                    int64                  `yaml:"chain_id"`
		FireblocksConfig           FireblocksConfig       `yaml:"fireblocks"`
		Web3SignerConfig           Web3SignerConfig       `yaml:"web3"`
	}{
		Operator:                   config.Operator,
		ELDelegationManagerAddress: config.ELDelegationManagerAddress,
		EthRPCUrl:                  config.EthRPCUrl,
		PrivateKeyStorePath:        config.PrivateKeyStorePath,
		SignerType:                 config.SignerType,
		ChainID:                    config.ChainId.Int64(),
		FireblocksConfig:           config.FireblocksConfig,
		Web3SignerConfig:           config.Web3SignerConfig,
	}, nil
}
