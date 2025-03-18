package container

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/urfave/cli/v2"
)

var (
	sha256Prefix       = "sha256:"
	hexPrefix          = "0x"
	annotationsKey     = "Annotations"
	signatureKey       = "io.eigen.signature"
	signerPublicKeyKey = "io.eigen.signer.publickey"
)

func NewVerifyContainerCmd() *cli.Command {
	return &cli.Command{
		Name:  "verify",
		Usage: "Verify a container signature from Github Container Registry.",
		Action: func(c *cli.Context) error {
			return executeVerification(c)
		},
		Flags: getVerifyContainerFlags(),
	}
}

func executeVerification(cliCtx *cli.Context) error {
	repo := cliCtx.String(repositoryLocationFlag.Name)
	tag := cliCtx.String(containerTagFlag.Name)

	signatureTag := fmt.Sprintf("%s.sig:%s", repo, tag)

	if err := pullImage(signatureTag); err != nil {
		fmt.Printf("No signature artifact found for %s in GHCR.\n", repo)
		return err
	}

	signatureArtifact, err := inspectSignatureArtifact(signatureTag)
	if err != nil {
		return fmt.Errorf("failed to inspect signature artifact: %w", err)
	}

	imageTag := fmt.Sprintf("%s:%s", repo, tag)
	imageSha, err := getImageSHA_2(imageTag)
	if err != nil {
		return fmt.Errorf("failed to retrieve image SHA for %s: %w", imageTag, err)
	}

	valid, extractedSigner := verifySignature(imageSha, signatureArtifact)
	if !valid {
		fmt.Println("Signature verification failed! The image was NOT signed by the expected key.")
		return fmt.Errorf("signature mismatch")
	}

	fmt.Printf("Successfully verified: %s:%s was signed by %s\n", repo, tag, extractedSigner)
	return nil
}

func pullImage(imageTag string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return fmt.Errorf("failed to initialize Docker client: %w", err)
	}
	defer cli.Close()

	out, err := cli.ImagePull(context.Background(), imageTag, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", imageTag, err)
	}
	defer out.Close()

	_, _ = io.Copy(io.Discard, out)
	return nil
}

func inspectSignatureArtifact(imageTag string) (map[string]interface{}, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Docker client: %w", err)
	}
	defer dockerClient.Close()

	inspectData, _, err := dockerClient.ImageInspectWithRaw(context.Background(), imageTag)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect image %s: %w", imageTag, err)
	}

	var result map[string]interface{}
	jsonData, err := json.Marshal(inspectData)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize inspection data: %w", err)
	}

	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, fmt.Errorf("failed to parse inspection JSON: %w", err)
	}

	return result, nil
}

func getImageSHA_2(imageTag string) (string, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return "", err
	}
	defer dockerClient.Close()

	inspect, _, err := dockerClient.ImageInspectWithRaw(context.Background(), imageTag)
	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(inspect.ID, sha256Prefix) {
		return "", fmt.Errorf("unexpected image ID format: %s", inspect.ID)
	}

	return strings.TrimPrefix(inspect.ID, sha256Prefix), nil
}

func verifySignature(imageSha string, artifactData map[string]interface{}) (bool, string) {
	annotations, ok := artifactData[annotationsKey].(map[string]interface{})
	if !ok {
		return false, ""
	}

	signedHashHex, ok := annotations[signatureKey].(string)
	if !ok {
		return false, ""
	}
	signerPkHex, ok := annotations[signerPublicKeyKey].(string)
	if !ok {
		return false, ""
	}

	signedHash, err := hex.DecodeString(signedHashHex)
	if err != nil {
		return false, ""
	}

	signerPkBytes, err := hex.DecodeString(strings.TrimPrefix(signerPkHex, hexPrefix))
	if err != nil {
		return false, ""
	}

	valid := crypto.VerifySignature(signerPkBytes, []byte(imageSha), signedHash)
	return valid, signerPkHex
}

func getVerifyContainerFlags() []cli.Flag {
	cmdFlags := []cli.Flag{
		&containerTagFlag,
		&repositoryLocationFlag,
		&flags.VerboseFlag,
	}
	sort.Sort(cli.FlagsByName(cmdFlags))
	return cmdFlags

}
