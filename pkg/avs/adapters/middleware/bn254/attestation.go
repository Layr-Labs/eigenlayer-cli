package bn254

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fp"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

type G1Point struct {
	*bn254.G1Affine
}

func newFpElement(x *big.Int) fp.Element {
	var p fp.Element
	p.SetBigInt(x)
	return p
}

func NewG1Point(x, y *big.Int) *G1Point {
	return &G1Point{
		&bn254.G1Affine{
			X: newFpElement(x),
			Y: newFpElement(y),
		},
	}
}

func (p *G1Point) Serialize() []byte {
	res := p.RawBytes()
	return res[:]
}

type G2Point struct {
	*bn254.G2Affine
}

func (p *G2Point) Serialize() []byte {
	res := p.RawBytes()
	return res[:]
}

type Signature struct {
	*G1Point
}

type PrivateKey = fr.Element

type KeyPair struct {
	PrivKey *PrivateKey
	PubKey  *G1Point
}

func (k *KeyPair) SignMessage(message [32]byte) *Signature {
	H := MapToCurve(message)
	sig := new(bn254.G1Affine).ScalarMultiplication(H, k.PrivKey.BigInt(new(big.Int)))
	return &Signature{&G1Point{sig}}
}

func (k *KeyPair) SignHashedToCurveMessage(g1HashedMsg *G1Point) *Signature {
	sig := new(bn254.G1Affine).ScalarMultiplication(g1HashedMsg.G1Affine, k.PrivKey.BigInt(new(big.Int)))
	return &Signature{&G1Point{sig}}
}

func (k *KeyPair) GetPubKeyG2() *G2Point {
	return &G2Point{MulByGeneratorG2(k.PrivKey)}
}

func (k *KeyPair) GetPubKeyG1() *G1Point {
	return k.PubKey
}
