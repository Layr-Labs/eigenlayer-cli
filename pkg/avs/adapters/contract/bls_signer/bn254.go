package bls_signer

import (
	"crypto/rand"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fp"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

const (
	sizeFr         = fr.Bytes
	sizeFp         = fp.Bytes
	sizePublicKey  = sizeFp
	sizePrivateKey = sizeFr
	sizeSignature  = 2 * sizeFp
)

var (
	dst   = []byte("0x01")
	order = fr.Modulus()
	one   = new(big.Int).SetInt64(1)
	g     bn254.G1Affine
	g1    bn254.G1Affine
	g2    bn254.G2Affine
	g3    bn254.G2Affine
)

func init() {
	_, _, g, g2 = bn254.Generators()
	g1.Neg(&g)
	g3.Neg(&g2)
}

type BN254Scheme struct {
}

func (s *BN254Scheme) GenerateRandomKey() ([]byte, error) {
	b := make([]byte, fr.Bits/8+8)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}

	k := new(big.Int).SetBytes(b)
	n := new(big.Int).Sub(order, one)
	k.Mod(k, n)
	k.Add(k, one)

	privKey := make([]byte, sizeFr)
	k.FillBytes(privKey)

	return privKey, nil
}

func (s *BN254Scheme) GetPublicKey(privKey []byte, isCompressed, isG1 bool) ([]byte, error) {
	scalar := new(big.Int)
	scalar.SetBytes(privKey[:sizeFr])

	if isG1 {
		pubKey := new(bn254.G1Affine)
		pubKey.ScalarMultiplication(&g, scalar)

		if isCompressed {
			pubKeyRaw := pubKey.Bytes()
			return pubKeyRaw[:sizeFp], nil
		}

		pubKeyRaw := pubKey.RawBytes()
		return pubKeyRaw[:], nil
	}

	pubKey := new(bn254.G2Affine)
	pubKey.ScalarMultiplication(&g2, scalar)
	if isCompressed {
		pubKeyRaw := pubKey.Bytes()
		return pubKeyRaw[:2*sizeFp], nil
	}
	pubKeyRaw := pubKey.RawBytes()
	return pubKeyRaw[:], nil
}

func (s *BN254Scheme) ConvertPublicKey(pubKey []byte, isCompressed, isG1 bool) ([]byte, error) {
	if isG1 {
		publicKey := new(bn254.G1Affine)
		_, err := publicKey.SetBytes(pubKey)
		if err != nil {
			return nil, err
		}

		if isCompressed {
			pubKeyRaw := publicKey.Bytes()
			return pubKeyRaw[:sizeFp], nil
		}

		pubKeyRaw := publicKey.RawBytes()
		return pubKeyRaw[:], nil
	}

	publicKey := new(bn254.G2Affine)
	_, err := publicKey.SetBytes(pubKey)
	if err != nil {
		return nil, err
	}

	if isCompressed {
		pubKeyRaw := publicKey.Bytes()
		return pubKeyRaw[:2*sizeFp], nil
	}

	pubKeyRaw := publicKey.RawBytes()
	return pubKeyRaw[:], nil
}

// MapToCurve implements the simple hash-and-check (also sometimes try-and-increment) algorithm
// see https://hackmd.io/@benjaminion/bls12-381#Hash-and-check
// Note that this function needs to be the same as the one used in the contract:
// https://github.com/Layr-Labs/eigenlayer-middleware/blob/1feb6ae7e12f33ce8eefb361edb69ee26c118b5d/src/libraries/BN254.sol#L292
// we don't use the newer constant time hash-to-curve algorithms as they are gas-expensive to compute onchain
func MapToCurve(digest [32]byte) *bn254.G1Affine {
	one := new(big.Int).SetUint64(1)
	three := new(big.Int).SetUint64(3)
	x := new(big.Int)
	x.SetBytes(digest[:])
	for {
		// y = x^3 + 3
		xP3 := new(big.Int).Exp(x, big.NewInt(3), fp.Modulus())
		y := new(big.Int).Add(xP3, three)
		y.Mod(y, fp.Modulus())

		if y.ModSqrt(y, fp.Modulus()) == nil {
			x.Add(x, one).Mod(x, fp.Modulus())
		} else {
			var fpX, fpY fp.Element
			fpX.SetBigInt(x)
			fpY.SetBigInt(y)
			return &bn254.G1Affine{
				X: fpX,
				Y: fpY,
			}
		}
	}
}

func (s *BN254Scheme) Sign(privKey, message []byte, isG1 bool) ([]byte, error) {
	// Convert the private key to a scalar
	scalar := new(big.Int)
	scalar.SetBytes(privKey[:sizeFr])

	if isG1 {
		// Hash the message into G2
		h, err := bn254.HashToG2(message, dst)
		if err != nil {
			return nil, err
		}

		sig := new(bn254.G2Affine)
		sig.ScalarMultiplication(&h, scalar)
		sigRaw := sig.Bytes()

		return sigRaw[:], nil
	}

	// Hash the message into G1
	var digest [32]byte
	copy(digest[:], message)
	h := MapToCurve(digest)
	sig := new(bn254.G1Affine)
	sig.ScalarMultiplication(h, scalar)
	sigRaw := sig.RawBytes()
	return sigRaw[:], nil
}

func (s *BN254Scheme) VerifySignature(pubKey, message, signature []byte, isG1 bool) (bool, error) {
	// Deserialize the public key
	pub := new(bn254.G1Affine)
	if _, err := pub.SetBytes(pubKey); err != nil {
		return false, err
	}

	// Deserialize the signature
	sig := new(bn254.G2Affine)
	if _, err := sig.SetBytes(signature); err != nil {
		return false, err
	}
	// Hash the message into G2
	h, err := bn254.HashToG2(message, dst)
	if err != nil {
		return false, err
	}

	// Verify the signature
	res, err := bn254.PairingCheck([]bn254.G1Affine{g1, *pub}, []bn254.G2Affine{*sig, h})
	if err != nil {
		return false, err
	}

	return res, nil
}

func (s *BN254Scheme) aggregateG1(points [][]byte) ([]byte, error) {
	agg := new(bn254.G1Affine)
	for _, p := range points {
		point := new(bn254.G1Affine)
		if _, err := point.SetBytes(p); err != nil {
			return nil, err
		}
		agg.Add(agg, point)
	}

	aggRaw := agg.RawBytes()
	return aggRaw[:], nil
}

func (s *BN254Scheme) aggregateG2(points [][]byte) ([]byte, error) {
	agg := new(bn254.G2Affine)
	for _, p := range points {
		point := new(bn254.G2Affine)
		if _, err := point.SetBytes(p); err != nil {
			return nil, err
		}
		agg.Add(agg, point)
	}

	aggRaw := agg.Bytes()
	return aggRaw[:], nil
}

func (s *BN254Scheme) AggregateSignatures(signatures [][]byte, isG1 bool) ([]byte, error) {
	if isG1 {
		return s.aggregateG1(signatures)
	}

	return s.aggregateG2(signatures)
}

func (s *BN254Scheme) AggregatePublicKeys(pubKeys [][]byte, isG1 bool) ([]byte, error) {
	if isG1 {
		return s.aggregateG1(pubKeys)
	}

	return s.aggregateG2(pubKeys)
}

func (s *BN254Scheme) VerifyAggregatedSignature(pubKeys [][]byte, message, signature []byte, isG1 bool) (bool, error) {
	aggPubKey, err := s.AggregatePublicKeys(pubKeys, isG1)
	if err != nil {
		return false, err
	}

	return s.VerifySignature(aggPubKey, message, signature, isG1)
}
