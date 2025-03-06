package signers

import (
	"crypto/ecdsa"
	sdkEcdsa "github.com/Layr-Labs/eigensdk-go/crypto/ecdsa"
	"github.com/ethereum/go-ethereum/crypto"
)

type LocalSigner struct {
	privateKey *ecdsa.PrivateKey
}

func (m *LocalSigner) Sign(data []byte) ([]byte, error) {
	return crypto.Sign(data, m.privateKey)
}

func NewLocalSigner(keystorePath string, password string) (*LocalSigner, error) {
	privateKey, err := sdkEcdsa.ReadKey(keystorePath, password)
	if err != nil {
		return nil, err
	}
	return &LocalSigner{privateKey: privateKey}, nil
}
