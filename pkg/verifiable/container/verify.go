package container

import (
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"sort"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/registry"

	eigensdkLogger "github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/urfave/cli/v2"
)

type verifySignatureCmd struct{}

func NewVerifyContainerCmd() *cli.Command {
	delegateCmd := verifySignatureCmd{}
	return NewVerifiableContainerCommand(
		delegateCmd,
		"verify",
		"Verify a container signature from Github Container Registry.",
		"",
		"",
		getVerifyContainerFlags(),
	)
}

func (v verifySignatureCmd) Execute(cliCtx *cli.Context) error {
	return executeSignatureVerification(cliCtx)
}

func executeSignatureVerification(cliCtx *cli.Context) error {
	logger := common.GetLogger(cliCtx)
	location := cliCtx.String(repositoryLocationFlag.Name)
	digest := cliCtx.String(containerDigestFlag.Name)
	signature, publicKey, err := getContainerSignatureAndPubKey(logger, location, digest)
	if err != nil {
		return err
	}
	isVerified := verifySignature(logger, digest, signature, publicKey)
	if !isVerified {
		return fmt.Errorf("container signature verification failed")
	}
	return nil
}

func getContainerSignatureAndPubKey(
	logger eigensdkLogger.Logger,
	location string,
	digest string,
) (string, string, error) {
	annotations := getVerificationComponents(location, digest)
	return parseSignatureWithPublicKey(logger, annotations)
}

func getVerificationComponents(location string, digest string) map[string]string {
	sigTag := fmt.Sprintf(signatureTagFormat, digest)
	sigRef := fmt.Sprintf(registryLocationTagFormat, location, sigTag)

	ref, err := name.NewTag(sigRef)
	if err != nil {
		return nil
	}

	desc, err := remote.Get(ref)
	if err != nil {
		return nil
	}

	var manifest v1.Manifest
	if err = json.Unmarshal(desc.Manifest, &manifest); err != nil {
		return nil
	}

	return manifest.Annotations
}

func parseSignatureWithPublicKey(logger eigensdkLogger.Logger, annotations map[string]string) (string, string, error) {
	signature, sigOk := annotations[registry.EigenSignatureKey]
	if !sigOk {
		return "", "", fmt.Errorf("signature not found in annotations")
	}
	signerAddress, addrOk := annotations[registry.EigenSignerAddressKey]
	if !addrOk {
		return "", "", fmt.Errorf("signer address not found in annotations")
	}
	publicKey, keyOk := annotations[registry.EigenPublicKey]
	if !keyOk {
		return "", "", fmt.Errorf("public key not found in annotations")
	}

	logger.Debugf("Fetched Signature: %s", signature)
	logger.Debugf("Fetched Public Key: %s", publicKey)
	logger.Debugf("Fetched Signer Address: %s", signerAddress)
	return signature, publicKey, nil
}

func verifySignature(
	logger eigensdkLogger.Logger,
	containerDigest string,
	signatureBase64 string,
	pubKeyHex string,
) bool {
	publicKey, err := getSignaturePublicKey(signatureBase64, containerDigest)
	if err != nil {
		logger.Fatalf("Failed to recover public key: %v", err)
		return false
	}
	recoveredHex := hex.EncodeToString(crypto.FromECDSAPub(publicKey))

	if len(pubKeyHex) > 2 && pubKeyHex[:2] == "0x" {
		pubKeyHex = pubKeyHex[2:]
	}

	if recoveredHex == pubKeyHex {
		logger.Infof("Signature is valid and matches expected public key!")
		return true
	}

	logger.Infof("Signature is valid but does not match the expected public key.")
	return false
}

func getSignaturePublicKey(signatureBase64 string, containerDigest string) (*ecdsa.PublicKey, error) {
	signatureBytes, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return nil, fmt.Errorf("rrror decoding signature to bytes: %v", err)
	}

	if len(signatureBytes) != 65 {
		return nil, fmt.Errorf("invalid signature length: expected 65 bytes, got %d", len(signatureBytes))
	}

	digestBytes, err := hex.DecodeString(containerDigest)
	if err != nil {
		return nil, fmt.Errorf("error decoding digest to bytes: %v", err)
	}

	return crypto.SigToPub(digestBytes, signatureBytes)
}

func getVerifyContainerFlags() []cli.Flag {
	cmdFlags := []cli.Flag{
		&repositoryLocationFlag,
		&flags.VerboseFlag,
	}
	sort.Sort(cli.FlagsByName(cmdFlags))
	return cmdFlags
}
