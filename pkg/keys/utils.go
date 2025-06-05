package keys

import (
	"fmt"
	"math/big"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigensdk-go/crypto/bls"
)

// ParseBlsPrivateKey parses a BLS private key from a string either in hex or large integer format.
func ParseBlsPrivateKey(privateKey string) (*bls.KeyPair, error) {
	privateKeyBigInt := new(big.Int)
	_, ok := privateKeyBigInt.SetString(privateKey, 10)
	var blsKeyPair *bls.KeyPair
	var err error
	if ok {
		fmt.Println("Importing from large integer")
		blsKeyPair, err = bls.NewKeyPairFromString(privateKey)
		if err != nil {
			return nil, err
		}
	} else {
		// Try to parse as hex
		fmt.Println("Importing from hex")
		z := new(big.Int)
		privateKey = common.Trim0x(privateKey)
		_, ok := z.SetString(privateKey, 16)
		if !ok {
			return nil, ErrInvalidHexPrivateKey
		}
		blsKeyPair, err = bls.NewKeyPairFromString(z.String())
		if err != nil {
			return nil, err
		}
	}
	return blsKeyPair, nil
}
