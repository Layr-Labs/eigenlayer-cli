package operator

import (
	"context"
	"crypto/ecdsa"
	"encoding/asn1"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	yubihsm "github.com/certusone/yubihsm-go"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/eth"
	"github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/certusone/yubihsm-go/commands"
	"github.com/certusone/yubihsm-go/connector"
	"github.com/ethereum/go-ethereum/common"
	gethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

/** A wallet that can sign transactions with a YubiHSM */
type YubiHsmWallet struct {
	operatorKeyId  uint16
	sessionManager *yubihsm.SessionManager
	logger         logging.Logger
	ethClient      eth.Client
}

/**
 * Example:
 * connectionUri: localhost:12345
 * authKeyID: 1
 * password: <secret password>
 * operatorKeyId: 16
 * <other common fields?
 */
func NewYubihsmWallet(connectionUri string, authKeyId uint16, password string, operatorKeyId uint16, logger logging.Logger, ethClient eth.Client) (*YubiHsmWallet, error) {
	connection := connector.NewHTTPConnector(connectionUri)
	sessionManager, err := yubihsm.NewSessionManager(connection, authKeyId, password)
	if err != nil {
		return nil, fmt.Errorf("unable to create a session with yubihsm: %s", err)
	}
	logger.Debugf("connected to yubihsm at %s", connectionUri)

	wallet := &YubiHsmWallet{
		sessionManager: sessionManager,
		operatorKeyId:  operatorKeyId,
		logger:         logger,
		ethClient:      ethClient,
	}

	// Ensure the operator key exists and uses the correct algo
	getPubKeyResponse, err := wallet.issueGetPublicKeyCommand()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve public key for operator key with ID %d, check connection and that the operator key ID is correct", operatorKeyId)
	}
	if getPubKeyResponse.Algorithm != commands.AlgorithmSecp256k1 {
		return nil, fmt.Errorf("key at ID %d was not a secp256k1 key. check key id", wallet.operatorKeyId)
	}
	logger.Debugf("validated public key with ID %d exists and is correct algorithm", wallet.operatorKeyId)

	return wallet, nil
}

func (wallet *YubiHsmWallet) Sign(digestHash []byte, sender common.Address) ([]byte, error) {
	// Sanity check
	wallet.logger.Debugf("signing transaction %s", hex.EncodeToString(digestHash))
	if len(digestHash) != gethcrypto.DigestLength {
		return nil, fmt.Errorf("hash is required to be exactly %d bytes (%d)", gethcrypto.DigestLength, len(digestHash))
	}

	// Sign
	signCommand, err := commands.CreateSignDataEcdsaCommand(wallet.operatorKeyId, digestHash)
	if err != nil {
		return nil, fmt.Errorf("unable to create a sign ecdsa command: %s", err)
	}

	rawResponse, err := wallet.sessionManager.SendEncryptedCommand(signCommand)
	if err != nil {
		return nil, fmt.Errorf("unable to send sign command: %s", err)
	}

	response, ok := rawResponse.(*commands.SignDataEcdsaResponse)
	if !ok {
		return nil, fmt.Errorf("unable to parse sign response")
	}

	// Unpack and parse the raw yubihsm response
	derSignature := response.Signature
	signature, err := normalizeSignature(derSignature, digestHash, sender)
	if err != nil {
		return nil, fmt.Errorf("unable to parse and canonize signature: %s", err.Error())
	}
	wallet.logger.Debugf("created signature %s", hex.EncodeToString(signature))

	return signature, nil
}

func (wallet *YubiHsmWallet) SenderAddress(ctx context.Context) (common.Address, error) {
	// Get raw key bytes from yubihsm
	response, err := wallet.issueGetPublicKeyCommand()
	if err != nil {
		return common.Address{}, fmt.Errorf("unable to retrieve public key with ID %d from yubihsm: %s", wallet.operatorKeyId, err.Error())
	}

	// Parse key bytes
	publicKeyBytes := response.KeyData
	secp256k1CurveParams := secp256k1.S256().Params()
	ecdsaPublicKey := ecdsa.PublicKey{
		Curve: secp256k1CurveParams,
		X:     new(big.Int).SetBytes(publicKeyBytes[:32]),
		Y:     new(big.Int).SetBytes(publicKeyBytes[32:]),
	}

	// Convert into an address with geth's utility libraries.
	return gethcrypto.PubkeyToAddress(ecdsaPublicKey), nil
}

/** Helpers */

func (wallet *YubiHsmWallet) issueGetPublicKeyCommand() (*commands.GetPubKeyResponse, error) {
	command, err := commands.CreateGetPubKeyCommand(wallet.operatorKeyId)
	if err != nil {
		return nil, fmt.Errorf("unable to create get public key command: %s", err)
	}

	rawResponse, err := wallet.sessionManager.SendEncryptedCommand(command)
	if err != nil {
		return nil, fmt.Errorf("unable to send get public key command: %s", err)
	}

	response, ok := rawResponse.(*commands.GetPubKeyResponse)
	if !ok {
		return nil, fmt.Errorf("unable to parse public key response")
	}
	return response, nil
}

// Normalizes a raw yubihsm signature
func normalizeSignature(yubihsmSignature []byte, digest []byte, signerAddress common.Address) ([]byte, error) {
	// YubiHSM2 signatures are in DER format, with R, S
	var parsedSignature struct {
		R *big.Int
		S *big.Int
	}
	if _, err := asn1.Unmarshal(yubihsmSignature, &parsedSignature); err != nil {
		return nil, fmt.Errorf("unable to unmarshal DER encoded signature: %s", err.Error())
	}

	// Create a canonical signature
	canonicalSignature := canonizeSignature(parsedSignature.R, parsedSignature.S)

	// Determine recovery value
	v, err := findV(canonicalSignature, digest, signerAddress)
	if err != nil {
		return nil, err
	}

	return append(canonicalSignature, v), nil
}

// Inspired by: https://github.com/ecadlabs/gotez/blob/v2/crypt/ecdsa.go#L172C1-L188C1
func canonizeSignature(r *big.Int, s *big.Int) []byte {
	order := secp256k1.S256().Params().N
	quo := new(big.Int).Quo(order, new(big.Int).SetInt64(2))
	if s.Cmp(quo) > 0 {
		s = s.Sub(order, s)
	}

	return append(r.Bytes(), s.Bytes()...)
}

func findV(signature []byte, digest []byte, signerAddress common.Address) (byte, error) {
	if testRecoverByte(0x00, signature, digest, signerAddress) {
		return 0x00, nil
	}

	if testRecoverByte(0x01, signature, digest, signerAddress) {
		return 0x01, nil
	}

	return 0xff, fmt.Errorf("unable to find a suitable value for v")
}

func testRecoverByte(v byte, signature []byte, digest []byte, expected common.Address) bool {
	recovered, err := secp256k1.RecoverPubkey(digest, append(signature, v))
	if err != nil {
		return false
	}

	secp256k1CurveParams := secp256k1.S256().Params()
	recoveredPublicKey := ecdsa.PublicKey{
		Curve: secp256k1CurveParams,
		// NOTE: `RecoverPubkey` prefixes key with magic byte 0x04 for uncompressed public key.
		X: new(big.Int).SetBytes(recovered[1:33]),
		Y: new(big.Int).SetBytes(recovered[33:]),
	}

	recoveredAddress := gethcrypto.PubkeyToAddress(recoveredPublicKey)

	return strings.EqualFold(recoveredAddress.Hex(), expected.Hex())
}
