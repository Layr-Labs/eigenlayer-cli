package container

import (
	eltypes "github.com/Layr-Labs/eigenlayer-cli/pkg/types"
)

const (
	signatureTagFormat        = "sha256-%s.sig"
	registryLocationTagFormat = "%s:%s"
)

type SignMessageConfig struct {
	SignerConfig       *eltypes.SignerConfig
	RepositoryLocation string
	ContainerDigest    string
	EcdsaPublicKey     string
}
