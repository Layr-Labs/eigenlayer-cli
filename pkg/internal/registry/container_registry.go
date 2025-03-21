package registry

import (
	"github.com/google/go-containerregistry/pkg/name"
)

const (
	EigenPublicKey        = "dev.eigen.signer.public-key"
	EigenSignatureKey     = "dev.eigen.signature"
	EigenSignerAddressKey = "dev.eigen.signer.address"
)

type ContainerRegistry interface {
	PushSignature(
		digestBytes []byte,
		signature []byte,
		publicKeyHex string,
		signerAddressHex string,
		tag name.Tag,
	) error
	TagSignature(registry string, digest string) (name.Tag, error)
}
