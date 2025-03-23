package container

import (
	eltypes "github.com/Layr-Labs/eigenlayer-cli/pkg/types"
)

const (
	expectedSignatureLength = 65
)

type SignMessageConfig struct {
	SignerConfig       *eltypes.SignerConfig
	RepositoryLocation string
	ContainerDigest    string
}
