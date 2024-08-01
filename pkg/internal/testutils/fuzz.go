package testutils

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/crypto"
)

func GenerateRandomEthereumAddressString() string {
	// Generate a new private key
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return ""
	}

	// Derive the public key from the private key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return ""
	}

	// Generate the Ethereum address from the public key
	return crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
}
