package container

import (
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/registry"
	eigensdkLogger "github.com/Layr-Labs/eigensdk-go/logging"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/urfave/cli/v2"
)

type verifySignatureCmd struct {
	registry registry.RegistryController
}

func NewVerifyContainerCmd(registry registry.RegistryController) *cli.Command {
	delegateCmd := verifySignatureCmd{registry: registry}
	return NewVerifiableContainerCommand(
		delegateCmd,
		"verify",
		"Verify a container signature from Github Container Registry.",
		"",
		"",
		[]cli.Flag{},
	)
}

func (v verifySignatureCmd) Execute(cliCtx *cli.Context) error {
	logger := common.GetLogger(cliCtx)
	location := cliCtx.String(repositoryLocationFlag.Name)
	digest := cliCtx.String(containerDigestFlag.Name)
	tag, err := v.registry.GetSignatureTag(location, digest)
	if err != nil {
		return err
	}
	signature, publicKey, err := v.registry.GetSignatureComponents(tag)
	if err != nil {
		return err
	}

	logger.Debugf("Retrieved Signature: %s and Public Key: %s", signature, publicKey)
	isVerified := verifySignature(logger, digest, signature, publicKey)
	if !isVerified {
		return fmt.Errorf("container signature verification failed")
	}
	return nil
}

func verifySignature(
	logger eigensdkLogger.Logger,
	containerDigest string,
	signatureBase64 string,
	providedPublicKey string,
) bool {
	publicKey, err := getSignaturePublicKey(signatureBase64, containerDigest)
	if err != nil {
		logger.Fatalf("Failed to recover public key: %v", err)
		return false
	}
	derivedPublicKey := hex.EncodeToString(crypto.FromECDSAPub(publicKey))

	if derivedPublicKey == providedPublicKey {
		logger.Infof("Signature is valid and matches expected public key!")
		return true
	}

	logger.Infof("Signature is valid but does not match the expected public key.")
	return false
}

func getSignaturePublicKey(signatureBase64 string, containerDigest string) (*ecdsa.PublicKey, error) {
	signatureBytes, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return nil, fmt.Errorf("error decoding signature to bytes: %v", err)
	}

	if len(signatureBytes) != expectedSignatureLength {
		return nil, fmt.Errorf("invalid signature length: expected 65 bytes, got %d", len(signatureBytes))
	}

	digestBytes, err := hex.DecodeString(containerDigest)
	if err != nil {
		return nil, fmt.Errorf("error decoding digest to bytes: %v", err)
	}

	return crypto.SigToPub(digestBytes, signatureBytes)
}
