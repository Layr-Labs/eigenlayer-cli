package types

import (
	"math/big"

	eigensdkTypes "github.com/Layr-Labs/eigensdk-go/types"
)

type SignerType string

const (
	PrivateKeySigner    SignerType = "private_key"
	LocalKeystoreSigner SignerType = "local_keystore"
)

type OperatorConfigNew struct {
	Operator                   eigensdkTypes.Operator `yaml:"operator"`
	ELDelegationManagerAddress string                 `yaml:"el_delegation_manager_address"`
	ELAVSDirectoryAddress      string
	EthRPCUrl                  string     `yaml:"eth_rpc_url"`
	PrivateKeyStorePath        string     `yaml:"private_key_store_path"`
	SignerType                 SignerType `yaml:"signer_type"`
	ChainId                    big.Int    `yaml:"chain_id"`
}

func (config OperatorConfigNew) MarshalYAML() (interface{}, error) {
	return struct {
		Operator                   eigensdkTypes.Operator `yaml:"operator"`
		ELDelegationManagerAddress string                 `yaml:"el_delegation_manager_address"`
		EthRPCUrl                  string                 `yaml:"eth_rpc_url"`
		PrivateKeyStorePath        string                 `yaml:"private_key_store_path"`
		SignerType                 SignerType             `yaml:"signer_type"`
		ChainID                    int64                  `yaml:"chain_id"`
	}{
		Operator:                   config.Operator,
		ELDelegationManagerAddress: config.ELDelegationManagerAddress,
		EthRPCUrl:                  config.EthRPCUrl,
		PrivateKeyStorePath:        config.PrivateKeyStorePath,
		SignerType:                 config.SignerType,
		ChainID:                    config.ChainId.Int64(),
	}, nil
}
