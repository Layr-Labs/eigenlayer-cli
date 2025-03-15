package bls_signer

import (
	"math/big"

	"github.com/Layr-Labs/eigensdk-go/crypto/bls"
)

type BLSSigner interface {
	Sign(digest []byte) (blsG1PublicKeys [][2]*big.Int, aggG2PublicKey [2][2]*big.Int, signature [2]*big.Int, err error)
	LoadBLSKey() (keyPair *bls.KeyPair, err error)
}
