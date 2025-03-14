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

type OperatorConfig struct {
	Operator                    eigensdkTypes.Operator `yaml:"operator"                      json:"operator"`
	ELDelegationManagerAddress  string                 `yaml:"el_delegation_manager_address" json:"el_delegation_manager_address"`
	ELAVSDirectoryAddress       string
	ELRewardsCoordinatorAddress string
	EthRPCUrl                   string  `yaml:"eth_rpc_url"                   json:"eth_rpc_url"`
	ChainId                     big.Int `yaml:"chain_id"                      json:"chain_id"`
	SignerConfig                SignerConfig
}

func (o *OperatorConfig) MarshalYAML() (interface{}, error) {
	return struct {
		Operator                   eigensdkTypes.Operator `yaml:"operator"`
		ELDelegationManagerAddress string                 `yaml:"el_delegation_manager_address"`
		EthRPCUrl                  string                 `yaml:"eth_rpc_url"`
		ChainId                    int64                  `yaml:"chain_id"`
		PrivateKeyStorePath        string                 `yaml:"private_key_store_path"`
		SignerType                 SignerType             `yaml:"signer_type"`
		Fireblocks                 FireblocksConfig       `yaml:"fireblocks"`
		Web3                       Web3SignerConfig       `yaml:"web3"`
	}{
		Operator:                   o.Operator,
		ELDelegationManagerAddress: o.ELDelegationManagerAddress,
		EthRPCUrl:                  o.EthRPCUrl,
		ChainId:                    o.ChainId.Int64(),
		PrivateKeyStorePath:        o.SignerConfig.PrivateKeyStorePath,
		SignerType:                 o.SignerConfig.SignerType,
		Fireblocks:                 o.SignerConfig.FireblocksConfig,
		Web3:                       o.SignerConfig.Web3SignerConfig,
	}, nil
}

func (o *OperatorConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var aux struct {
		Operator                    eigensdkTypes.Operator `yaml:"operator"`
		ELDelegationManagerAddress  string                 `yaml:"el_delegation_manager_address"`
		ELAVSDirectoryAddress       string                 `yaml:"el_avs_directory_address"`
		ELRewardsCoordinatorAddress string                 `yaml:"el_rewards_coordinator_address"`
		EthRPCUrl                   string                 `yaml:"eth_rpc_url"`
		ChainId                     int64                  `yaml:"chain_id"`
		PrivateKeyStorePath         string                 `yaml:"private_key_store_path"`
		SignerType                  SignerType             `yaml:"signer_type"`
		Fireblocks                  FireblocksConfig       `yaml:"fireblocks"`
		Web3                        Web3SignerConfig       `yaml:"web3"`
	}
	if err := unmarshal(&aux); err != nil {
		return err
	}
	o.Operator = aux.Operator
	o.ELDelegationManagerAddress = aux.ELDelegationManagerAddress
	o.ELAVSDirectoryAddress = aux.ELAVSDirectoryAddress
	o.ELRewardsCoordinatorAddress = aux.ELRewardsCoordinatorAddress
	o.EthRPCUrl = aux.EthRPCUrl

	chainId := new(big.Int)
	chainId.SetInt64(aux.ChainId)
	o.ChainId = *chainId

	o.SignerConfig.PrivateKeyStorePath = aux.PrivateKeyStorePath
	o.SignerConfig.SignerType = aux.SignerType
	o.SignerConfig.FireblocksConfig = aux.Fireblocks
	o.SignerConfig.Web3SignerConfig = aux.Web3
	return nil
}
