package bls_signer

import (
	"math/big"

	"github.com/Layr-Labs/eigensdk-go/crypto/bls"
	"github.com/consensys/gnark-crypto/ecc/bn254"
)

var _ BLSSigner = (*BLSLocalSigner)(nil)

// BLSSigner using a locally stored ECDSA private key.
type BLSLocalSigner struct {
	keyFile  string
	password string
}

func NewBLSLocalSigner(keyFile string, password string) (*BLSLocalSigner, error) {
	return &BLSLocalSigner{keyFile: keyFile, password: password}, nil
}

func (m *BLSLocalSigner) Sign(digest []byte) (
	blsG1PublicKeys [][2]*big.Int,
	aggG2PublicKey [2][2]*big.Int,
	signature [2]*big.Int,
	err error,
) {
	key, err := bls.ReadPrivateKeyFromFile(m.keyFile, m.password)
	if err != nil {
		return
	}

	blsPrivateKey := key.PrivKey.Marshal()
	scheme := &BN254Scheme{}

	pubKeyG2, err := scheme.GetPublicKey(blsPrivateKey, false, false)
	if err != nil {
		return
	}
	var pubKey []byte
	pubKey, err = scheme.GetPublicKey(blsPrivateKey, false, true)
	if err != nil {
		return
	}

	pubKeyX := new(big.Int).SetBytes(pubKey[:32])
	pubKeyY := new(big.Int).SetBytes(pubKey[32:])
	blsG1PublicKeys = append(blsG1PublicKeys, [2]*big.Int{pubKeyX, pubKeyY})

	signatures, err := scheme.Sign(blsPrivateKey, digest, false)
	if err != nil {
		return
	}

	aggG2 := new(bn254.G2Affine)
	if _, err = aggG2.SetBytes(pubKeyG2); err != nil {
		return blsG1PublicKeys, aggG2PublicKey, signature, err
	}
	aggG2PublicKey[0][0] = aggG2.X.A1.BigInt(big.NewInt(0))
	aggG2PublicKey[0][1] = aggG2.X.A0.BigInt(big.NewInt(0))
	aggG2PublicKey[1][0] = aggG2.Y.A1.BigInt(big.NewInt(0))
	aggG2PublicKey[1][1] = aggG2.Y.A0.BigInt(big.NewInt(0))

	aggSig := new(bn254.G1Affine)
	if _, err = aggSig.SetBytes(signatures); err != nil {
		return blsG1PublicKeys, aggG2PublicKey, signature, err
	}
	signature[0] = aggSig.X.BigInt(big.NewInt(0))
	signature[1] = aggSig.Y.BigInt(big.NewInt(0))

	return
}

func (m *BLSLocalSigner) LoadBLSKey() (
	keyPair *bls.KeyPair,
	err error,
) {

	key, err := bls.ReadPrivateKeyFromFile(m.keyFile, m.password)
	if err != nil {
		return nil, err
	}

	return key, nil
}
