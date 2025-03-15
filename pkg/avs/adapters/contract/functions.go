package contract

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/avs/adapters/contract/bls_signer"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/avs/adapters/middleware/bn254"
	"github.com/Layr-Labs/eigensdk-go/crypto/bls"
	"github.com/ethereum/go-ethereum/accounts/keystore"
)

func arg(c *Coordinator, params map[string]string, key string, kind string, required bool) (interface{}, error) {
	param, ok := params[key]
	if required && !ok {
		return nil, fmt.Errorf("required function parameter missing: %s", key)
	}

	if param == "" {
		return nil, nil
	}

	value, err := c.load(param, kind, required)
	if err != nil {
		return nil, err
	}

	return value, nil
}

// array_uint8: convert to []uint8
//
// input:
// list	comma separated list of values, each value will be converted to a uint8
//
// output: []uint8 with the converted input
func array_uint8(c *Coordinator, params map[string]string) ([]uint8, error) {
	values, err := arg(c, params, "list", "string", true)
	if err != nil {
		return []uint8{}, err
	}

	tokens := strings.Split(values.(string), ",")
	var result []uint8
	for _, i := range tokens {
		value, err := strconv.Atoi(i)
		if err != nil {
			return []uint8{}, err
		}

		result = append(result, uint8(value))
	}

	return result, nil
}

type BN254G1Point struct {
	X *big.Int
	Y *big.Int
}

type BN254G2Point struct {
	X [2]*big.Int
	Y [2]*big.Int
}

// bls_sign: sign using bls key
//
// input:
// type		signer type, valid values are (local_keystore)
// file		when type=local_keystore, the bls key file path
// password	when type=local_keystore, the bls key file password
// hash		hash to use as digest for signing
// salt		alt to include in the result
// expiry	expiry to include in the result
//
// output: map[string]interface{} with the following items
// null[bn254g1]		all zero value bn254 g1 point
// null[bn254g2]		all zero value bn254 g2 point
// g1					g1 public keys
// g1[bn254g1]			g1 public key as a bn254 g1 point
// g2					aggregated g2 public key
// g2[bn254g2]			aggregated g2 public key as a bn254 g3 point
// signature 			signature
// signature[bn254g1]	signature as a bn254 g1 point
// salt					input salt
// expiry				input expiry
func bls_sign(c *Coordinator, params map[string]string) (interface{}, error) {
	signerType, err := arg(c, params, "type", "string", true)
	if err != nil {
		return nil, err
	}

	var signer bls_signer.BLSSigner
	switch signerType {
	case "local_keystore":
		keyFile, err := arg(c, params, "file", "string", true)
		if err != nil {
			return nil, err
		}

		password, err := arg(c, params, "password", "string", false)
		if err != nil {
			return nil, err
		}

		if password == nil {
			password = ""
		}

		signer, err = bls_signer.NewBLSLocalSigner(keyFile.(string), password.(string))
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported signer type: %s", signerType)
	}

	hash, err := arg(c, params, "hash", "", true)
	if err != nil {
		return nil, err
	}

	var digest []byte
	switch vt := hash.(type) {
	case []byte:
		_ = vt
		digest = append(digest, hash.([]byte)...)
	case [32]byte:
		digest = make([]byte, 32)
		bytes := hash.([32]byte)
		copy(digest, bytes[:])
	default:
		return nil, fmt.Errorf("unsupported hash type: %T", hash)
	}

	g1, g2, signature, err := signer.Sign(digest)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	result["null[bn254g1]"] = BN254G1Point{X: big.NewInt(0), Y: big.NewInt(0)}
	result["null[bn254g2]"] = BN254G2Point{
		X: [2]*big.Int{big.NewInt(0), big.NewInt(0)},
		Y: [2]*big.Int{big.NewInt(0), big.NewInt(0)},
	}
	result["g1"] = g1
	result["g1[bn254g1]"] = BN254G1Point{X: g1[0][0], Y: g1[0][1]}
	result["g2"] = g2
	result["g2[bn254g2]"] = BN254G2Point{X: g2[0], Y: g2[1]}
	result["signature"] = signature
	result["signature[bn254g1]"] = BN254G1Point{X: signature[0], Y: signature[1]}

	salt, err := arg(c, params, "salt", "", false)
	if err != nil {
		return nil, err
	}

	if salt != nil {
		saltBytes, ok := salt.([32]byte)
		if !ok {
			return nil, fmt.Errorf("failed type conversion on salt")
		}

		result["salt"] = saltBytes
	}

	expiry, err := arg(c, params, "expiry", "", false)
	if err != nil {
		return nil, err
	}

	if expiry != nil {
		expiryBigInt, ok := expiry.(*big.Int)
		if !ok {
			return nil, fmt.Errorf("failed type conversion on expiry")
		}

		result["expiry"] = expiryBigInt
	}

	return result, nil
}

// bls_curve_sign: sign using bls key
//
// input:
// type		signer type, valid values are (local_keystore)
// file		when type=local_keystore, the bls key file path
// password	when type=local_keystore, the bls key file password
// hash		hash to use as digest for signing
// salt		alt to include in the result
// expiry	expiry to include in the result
//
// output: map[string]interface{} with the following items
// null[bn254g1]		all zero value bn254 g1 point
// null[bn254g2]		all zero value bn254 g2 point
// g1[bn254g1]			g1 public key as a bn254 g1 point
// g2[bn254g2]			aggregated g2 public key as a bn254 g3 point
// signature[bn254g1]	signature as a bn254 g1 point
// salt					input salt
// expiry				input expiry
func bls_curve_sign(c *Coordinator, params map[string]string) (interface{}, error) {
	signerType, err := arg(c, params, "type", "string", true)
	if err != nil {
		return nil, err
	}

	var keyPair *bn254.KeyPair
	switch signerType {
	case "local_keystore":
		keyFile, err := arg(c, params, "file", "string", true)
		if err != nil {
			return nil, err
		}

		password, err := arg(c, params, "password", "string", false)
		if err != nil {
			return nil, err
		}
		if password == nil {
			password = ""
		}

		kp, err := bls.ReadPrivateKeyFromFile(keyFile.(string), password.(string))
		if err != nil {
			return nil, err
		}

		keyPair = &bn254.KeyPair{
			PrivKey: kp.PrivKey,
			PubKey: &bn254.G1Point{
				G1Affine: kp.PubKey.G1Affine,
			},
		}

	default:
		return nil, fmt.Errorf("unsupported signer type: %s", signerType)
	}

	hash, err := arg(c, params, "hash", "", true)
	if err != nil {
		return nil, err
	}

	message := bn254.NewG1Point(new(big.Int).SetBytes(hash.([]byte)[:32]), new(big.Int).SetBytes(hash.([]byte)[32:]))
	signature := keyPair.SignHashedToCurveMessage(message)

	signedMessageHashParam := BN254G1Point{
		X: signature.X.BigInt(big.NewInt(0)),
		Y: signature.Y.BigInt(big.NewInt(0)),
	}

	g1 := keyPair.GetPubKeyG1()
	g1Point := BN254G1Point{
		X: g1.X.BigInt(new(big.Int)),
		Y: g1.Y.BigInt(new(big.Int)),
	}

	g2 := keyPair.GetPubKeyG2()
	g2Point := BN254G2Point{
		X: [2]*big.Int{g2.X.A1.BigInt(new(big.Int)), g2.X.A0.BigInt(new(big.Int))},
		Y: [2]*big.Int{g2.Y.A1.BigInt(new(big.Int)), g2.Y.A0.BigInt(new(big.Int))},
	}

	result := make(map[string]interface{})
	result["null[bn254g1]"] = BN254G1Point{X: big.NewInt(0), Y: big.NewInt(0)}
	result["null[bn254g2]"] = BN254G2Point{
		X: [2]*big.Int{big.NewInt(0), big.NewInt(0)},
		Y: [2]*big.Int{big.NewInt(0), big.NewInt(0)},
	}
	result["g1[bn254g1]"] = g1Point
	result["g2[bn254g2]"] = g2Point
	result["signature[bn254g1]"] = signedMessageHashParam

	salt, err := arg(c, params, "salt", "", false)
	if err != nil {
		return nil, err
	}

	if salt != nil {
		saltBytes, ok := salt.([32]byte)
		if !ok {
			return nil, fmt.Errorf("failed type conversion on salt")
		}

		result["salt"] = saltBytes
	}

	expiry, err := arg(c, params, "expiry", "", false)
	if err != nil {
		return nil, err
	}

	if expiry != nil {
		expiryBigInt, ok := expiry.(*big.Int)
		if !ok {
			return nil, fmt.Errorf("failed type conversion on expiry")
		}

		result["expiry"] = expiryBigInt
	}

	return result, nil

}

// chain_id: get the id of the chain
//
// input: none
//
// output: id if the chain
func chain_id(c *Coordinator, params map[string]string) (int64, error) {
	client, err := c.Client()
	if err != nil {
		return -1, err
	}

	chainId, err := client.ChainID(context.Background())
	if err != nil {
		return -1, err
	}

	return chainId.Int64(), nil
}

type ECDSAPublicKey struct {
	X *big.Int
	Y *big.Int
}

// ecdsa_public_key: load a ecdsa public key from a file
//
// input:
// type		    signer type, valid values are (local_keystore)
// file			file path
// password 	pass-phrase for the file
// format		output format ([]byte or default = struct{X,Y})
//
// output: public key loaded from file
func ecdsa_public_key(c *Coordinator, params map[string]string) (interface{}, error) {
	file, err := arg(c, params, "file", "string", true)
	if err != nil {
		return nil, err
	}

	password, err := arg(c, params, "password", "string", false)
	if err != nil {
		return nil, err
	}

	if password == nil {
		password = ""
	}

	contents, err := os.ReadFile(file.(string))
	if err != nil {
		return nil, err
	}

	key, err := keystore.DecryptKey(contents, password.(string))
	if err != nil {
		return nil, err
	}

	publicKey := key.PrivateKey.PublicKey
	format, err := arg(c, params, "format", "string", false)
	if err != nil {
		return nil, err
	}

	switch format {
	case "[]byte":
		return append(publicKey.X.Bytes(), publicKey.Y.Bytes()...), nil

	default:
		return &ECDSAPublicKey{
			X: publicKey.X,
			Y: publicKey.Y,
		}, nil
	}
}

// bls_public_key: load a bls public key from a file
//
// input:
// file			file path
// password 	pass-phrase for the file
// format		output format ([]byte or default = struct{X,Y})
//
// output: bls public key loaded from file
func bls_public_key(c *Coordinator, params map[string]string) (interface{}, error) {
	signerType, err := arg(c, params, "type", "string", true)
	if err != nil {
		return nil, err
	}

	var signer bls_signer.BLSSigner
	switch signerType {
	case "local_keystore":
		keyFile, err := arg(c, params, "file", "string", true)
		if err != nil {
			return nil, err
		}

		password, err := arg(c, params, "password", "string", false)
		if err != nil {
			return nil, err
		}
		if password == nil {
			password = ""
		}

		signer, err = bls_signer.NewBLSLocalSigner(keyFile.(string), password.(string))
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported signer type: %s", signerType)
	}

	key, err := signer.LoadBLSKey()
	if err != nil {
		return nil, err
	}

	format, err := arg(c, params, "format", "string", false)
	if err != nil {
		return nil, err
	}

	pubKeyG2 := key.GetPubKeyG2()

	switch format {
	case "[]byte":

		aggG2XA0 := pubKeyG2.X.A0.Marshal()
		aggG2XA1 := pubKeyG2.X.A1.Marshal()
		aggG2YA0 := pubKeyG2.Y.A0.Marshal()
		aggG2YA1 := pubKeyG2.Y.A1.Marshal()

		return append(aggG2XA0, append(aggG2XA1, append(aggG2YA0, aggG2YA1...)...)...), nil
	default:
		return &BN254G2Point{
			X: [2]*big.Int{pubKeyG2.X.A1.BigInt(big.NewInt(0)), pubKeyG2.X.A0.BigInt(big.NewInt(0))},
			Y: [2]*big.Int{pubKeyG2.Y.A1.BigInt(big.NewInt(0)), pubKeyG2.Y.A0.BigInt(big.NewInt(0))},
		}, nil
	}
}

type ECDSASignatureWithSaltAndExpiry struct {
	Signature []byte
	Salt      [32]byte
	Expiry    *big.Int
}

// ecdsa_sign: sign with salt and expiry
//
// input:
// hash		hash to sign
// salt		salt to include in the output
// expiry	expiry to include in the output
//
// output: signature with salt and expiry
func ecdsa_sign(c *Coordinator, params map[string]string) (*ECDSASignatureWithSaltAndExpiry, error) {
	hash, err := arg(c, params, "hash", "", true)
	if err != nil {
		return nil, err
	}

	bytes := hash.([32]byte)
	signature, err := c.TxMgr.Sign(bytes[:])
	if err != nil {
		return nil, err
	}

	signature[64] += 27

	salt, err := arg(c, params, "salt", "", true)
	if err != nil {
		return nil, err
	}

	saltBytesValue, ok := salt.([32]byte)
	if !ok {
		return nil, fmt.Errorf("failed type conversion on salt")
	}

	expiry, err := arg(c, params, "expiry", "", true)
	if err != nil {
		return nil, err
	}

	expiryBigIntValue, ok := expiry.(*big.Int)
	if !ok {
		return nil, fmt.Errorf("failed type conversion on expiry")
	}

	return &ECDSASignatureWithSaltAndExpiry{
		Signature: signature,
		Salt:      saltBytesValue,
		Expiry:    expiryBigIntValue,
	}, nil
}

// expiry: compute expiry
//
// input:
// timeout	timeout seconds to add to current block time
//
// output: expiry timestamp
func expiry(c *Coordinator, params map[string]string) (*big.Int, error) {
	timeout, err := arg(c, params, "timeout", "uint64", true)
	if err != nil {
		return nil, err
	}

	client, err := c.Client()
	if err != nil {
		return nil, err
	}

	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	expiry := header.Time + timeout.(uint64) + 2000
	c.store("last:expiry", big.NewInt(int64(expiry)))

	return big.NewInt(int64(expiry)), nil
}

// salt: compute salt
//
// input:
// seed		seed for the salt
//
// output: computed salt
func salt(c *Coordinator, params map[string]string) ([32]byte, error) {
	seed, err := arg(c, params, "seed", "string", true)
	if err != nil {
		return [32]byte{}, err
	}

	bytes := []byte(seed.(string))
	var salt [32]byte
	copy(salt[:], bytes)
	copy(salt[len(bytes):], big.NewInt(time.Now().UnixNano()).Bytes())

	c.store("last:salt", salt)
	return salt, nil
}
