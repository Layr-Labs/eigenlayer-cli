package container

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/verifiable"

	"github.com/docker/docker/client"
	"github.com/urfave/cli/v2"
)

var shaPrefix = "sha256:"

type SignContainerCmd interface {
	executeSignature(config *verifiable.SignMessageConfig) error
}

func NewSignContainerCmd(prompter utils.Prompter) *cli.Command {
	setCmd := &cli.Command{
		Name:  "sign",
		Usage: "Sign a container using a specified key or remote signer.",
		Action: func(c *cli.Context) error {
			return executeSignature(c, prompter)
		},
		After: telemetry.AfterRunAction(),
		Flags: getContainerSignerFlags(),
	}

	return setCmd
}

func getImageSHA(imageId string) (string, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return "", fmt.Errorf("failed to connect to Docker daemon: %w", err)
	}
	defer func(dockerClient *client.Client) {
		closeErr := dockerClient.Close()
		if closeErr != nil {
			log.Printf("failed to close Docker daemon: %v", closeErr)
		}
	}(dockerClient)

	imageInspect, _, err := dockerClient.ImageInspectWithRaw(context.Background(), imageId)
	if err != nil {
		return "", fmt.Errorf("failed to inspect image %s: %w", imageId, err)
	}

	// TODO: what is this?
	if !strings.HasPrefix(imageInspect.ID, shaPrefix) {
		return "", fmt.Errorf("unexpected image ID format: %s", imageInspect.ID)
	}

	return strings.TrimPrefix(imageInspect.ID, shaPrefix), nil
}

func validateAndGenerateConfig(cCtx *cli.Context) (*verifiable.SignMessageConfig, error) {
	logger := common.GetLogger(cCtx)

	signerConfig, err := common.GetSignerConfig(cCtx, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create signer config: %w", err)
	}

	imageId := cCtx.String(imageIdFlag.Name)
	location := cCtx.String(repositoryLocationFlag.Name)

	return &verifiable.SignMessageConfig{
		SignerConfig:       signerConfig,
		RepositoryLocation: location,
		ImageId:            imageId,
	}, nil
}

func executeSignature(cliCtx *cli.Context, prompter utils.Prompter) error {
	signatureConfig, err := validateAndGenerateConfig(cliCtx)
	if err != nil {
		return err
	}
	signerFn, signerPk, err := common.GetMessageSigner(signatureConfig.SignerConfig, prompter)
	if err != nil {
		return err
	}
	signerPkHex := signerPk.Hex()
	imageSha, err := getImageSHA(signatureConfig.ImageId)
	if err != nil {
		return err
	}
	signedHash, err := signerFn([]byte(imageSha))
	if err != nil {
		return err
	}
	artifact := verifiable.NewOCISignatureArtifact(
		imageSha,
		signedHash,
		signerPkHex,
		signatureConfig.RepositoryLocation,
	)
	jsonOutput, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize signature artifact: %w", err)
	}

	fmt.Println(string(jsonOutput))
	return nil
}

func getContainerSignerFlags() []cli.Flag {
	cmdFlags := []cli.Flag{
		&flags.VerboseFlag,
		&imageIdFlag,
		&repositoryLocationFlag,
		&flags.EcdsaPrivateKeyFlag,
		&flags.PathToKeyStoreFlag,
		&flags.Web3SignerUrlFlag,
	}
	sort.Sort(cli.FlagsByName(cmdFlags))
	return cmdFlags
}
