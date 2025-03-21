package container

import (
	"encoding/hex"
	"fmt"
	"github.com/Layr-Labs/eigensdk-go/logging"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/registry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/urfave/cli/v2"
)

const (
	expectedKeyLength = 65
)

type signContainerCmd struct {
	prompter utils.Prompter
	registry registry.ContainerRegistry
}

func NewSignContainerCmd(prompter utils.Prompter, registry registry.ContainerRegistry) *cli.Command {
	delegateCommand := signContainerCmd{prompter: prompter, registry: registry}
	return NewVerifiableContainerCommand(
		delegateCommand,
		"sign",
		"Sign a container using a specified key or remote signer.",
		"",
		"",
		getContainerSignerFlags(),
	)
}

func (s signContainerCmd) Execute(cliCtx *cli.Context) error {
	logger := common.GetLogger(cliCtx)
	cfg, err := validateAndGenerateConfig(cliCtx, logger)
	if err != nil {
		return fmt.Errorf("failed to validate signature config: %w", err)
	}

	signerFn, signerAddress, err := common.GetMessageSigner(cfg.SignerConfig, s.prompter)
	if err != nil {
		return fmt.Errorf("failed to get message signer: %w", err)
	}
	digest := cfg.ContainerDigest
	digestBytes, err := hex.DecodeString(digest)
	if err != nil {
		return fmt.Errorf("failed to sign image: %w", err)
	}

	signature, err := signerFn(digestBytes)
	if err != nil {
		return fmt.Errorf("failed to sign image: %w", err)
	}

	tag, err := s.registry.TagSignature(cfg.RepositoryLocation, digest)
	if err != nil {
		return err
	}
	pubKeyHex, err := extractPublicKeyHexFromSignature(digest, signature)
	if err != nil {
		return fmt.Errorf("failed to extract public key: %w", err)
	}
	signerAddressHex := signerAddress.Hex()
	return s.registry.PushSignature(digestBytes, signature, pubKeyHex, signerAddressHex, tag)
}

func validateAndGenerateConfig(cCtx *cli.Context, logger logging.Logger) (*SignMessageConfig, error) {
	signerConfig, err := common.GetSignerConfig(cCtx, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create signer config: %w", err)
	}

	digest := cCtx.String(containerDigestFlag.Name)
	location := cCtx.String(repositoryLocationFlag.Name)
	ecdsaPublicKey := cCtx.String(ecdsaPublicKeyFlag.Name)

	return &SignMessageConfig{
		SignerConfig:       signerConfig,
		RepositoryLocation: location,
		ContainerDigest:    digest,
		EcdsaPublicKey:     ecdsaPublicKey,
	}, nil
}

func extractPublicKeyHexFromSignature(containerDigest string, sigBytes []byte) (string, error) {
	if len(sigBytes) != expectedKeyLength {
		return "", fmt.Errorf("invalid signature length: expected 65 bytes, got %d", len(sigBytes))
	}

	digestBytes, err := hex.DecodeString(containerDigest)
	if err != nil {

	}
	pubKey, err := crypto.SigToPub(digestBytes, sigBytes)
	if err != nil {
		return "", fmt.Errorf("failed to recover public key: %w", err)
	}

	pubKeyBytes := crypto.FromECDSAPub(pubKey)
	pubKeyHex := hex.EncodeToString(pubKeyBytes)

	return pubKeyHex, nil
}

func getContainerSignerFlags() []cli.Flag {
	cmdFlags := []cli.Flag{
		&flags.VerboseFlag,
		&repositoryLocationFlag,
		&ecdsaPublicKeyFlag,
		&flags.PathToKeyStoreFlag,
		&flags.Web3SignerUrlFlag,
	}
	return cmdFlags
}
