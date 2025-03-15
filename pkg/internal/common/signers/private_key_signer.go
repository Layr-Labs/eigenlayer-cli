package signers

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/crypto"
)

type PrivateKeySigner struct {
	privateKey *ecdsa.PrivateKey
}

func (m *PrivateKeySigner) Sign(data []byte) ([]byte, error) {
	return crypto.Sign(data, m.privateKey)
}

func NewPrivateKeySigner(privateKey *ecdsa.PrivateKey) (*PrivateKeySigner, error) {
	return &PrivateKeySigner{privateKey: privateKey}, nil
}
