package types

import "crypto/ecdsa"

type SignerConfig struct {
	PrivateKeyStorePath string           `yaml:"private_key_store_path"`
	SignerType          SignerType       `yaml:"signer_type"`
	FireblocksConfig    FireblocksConfig `yaml:"fireblocks"`
	Web3SignerConfig    Web3SignerConfig `yaml:"web3"`
	PrivateKey          *ecdsa.PrivateKey
}

func (s SignerConfig) MarshalYAML() (interface{}, error) {
	return struct {
		PrivateKeyStorePath string           `yaml:"private_key_store_path"`
		SignerType          SignerType       `yaml:"signer_type"`
		FireblocksConfig    FireblocksConfig `yaml:"fireblocks"`
		Web3SignerConfig    Web3SignerConfig `yaml:"web3"`
	}{
		PrivateKeyStorePath: s.PrivateKeyStorePath,
		SignerType:          s.SignerType,
		FireblocksConfig:    s.FireblocksConfig,
		Web3SignerConfig:    s.Web3SignerConfig,
	}, nil
}
