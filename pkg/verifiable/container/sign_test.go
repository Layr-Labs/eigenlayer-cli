package container

import (
	"encoding/hex"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/go-containerregistry/pkg/name"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"testing"
)

func TestSignContainerCmd_Execute_Success(t *testing.T) {
	privateKey, err := crypto.GenerateKey()
	assert.NoError(t, err)

	privateKeyHex := hex.EncodeToString(crypto.FromECDSA(privateKey))
	signerAddress := crypto.PubkeyToAddress(privateKey.PublicKey)

	digest := "a6eb5617ec3be5f0f523829e371ede989e8a3d15336a3030594d349fb14c92e8"
	repo := "ghcr.io/example/container"
	tag, _ := name.NewTag(repo + ":signed")

	mockReg := &mockRegistry{
		tag: tag,
	}
	set := flagSet(map[string]string{
		containerDigestFlag.Name:       digest,
		repositoryLocationFlag.Name:    "ghcr.io/testing/registry-name",
		flags.EcdsaPrivateKeyFlag.Name: privateKeyHex,
	})
	ctx := cli.NewContext(nil, set, nil)
	cmd := signContainerCmd{prompter: nil, registry: mockReg}
	err = cmd.Execute(ctx)
	assert.NoError(t, err)

	assert.Equal(t, digest, hex.EncodeToString(mockReg.pushedDigest))
	assert.Equal(t, signerAddress.Hex(), mockReg.pushedSignerAddr)
	assert.Equal(t, tag, mockReg.pushedTag)

	expectedPubHex := hex.EncodeToString(crypto.FromECDSAPub(&privateKey.PublicKey))
	assert.Equal(t, expectedPubHex, mockReg.pushedPublicKey)
}

func TestValidateAndGenerateConfig_MissingSignerConfig(t *testing.T) {
	set := flagSet(map[string]string{
		containerDigestFlag.Name:    "a6eb5617ec3be5f0f523829e371ede989e8a3d15336a3030594d349fb14c92e8",
		repositoryLocationFlag.Name: "ghcr.io/testing/registry-name",
	})

	ctx := cli.NewContext(nil, set, nil)
	cfg, err := validateAndGenerateConfig(ctx, common.GetLogger(ctx))

	assert.Nil(t, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create signer config")
}

func TestExtractPublicKeyHexFromSignature_Valid(t *testing.T) {
	validDigest := "a6eb5617ec3be5f0f523829e371ede989e8a3d15336a3030594d349fb14c92e8"
	privateKey, err := crypto.GenerateKey()
	assert.NoError(t, err)

	digestBytes, err := hex.DecodeString(validDigest)
	assert.NoError(t, err)
	digestHex := hex.EncodeToString(digestBytes)

	signature, err := crypto.Sign(digestBytes, privateKey)
	assert.NoError(t, err)

	pubKeyHex, err := extractPublicKeyHexFromSignature(digestHex, signature)
	assert.NoError(t, err)

	expectedPubKeyHex := hex.EncodeToString(crypto.FromECDSAPub(&privateKey.PublicKey))
	assert.Equal(t, expectedPubKeyHex, pubKeyHex)
}

func TestExtractPublicKeyHexFromSignature_InvalidSigLength(t *testing.T) {
	digestHex := hex.EncodeToString([]byte("a6eb5617ec3be5f0f523829e371ede989e8a3d15336a3030594d349fb14c92e8"))
	shortSig := make([]byte, 64)

	_, err := extractPublicKeyHexFromSignature(digestHex, shortSig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid signature length")
}

func TestExtractPublicKeyHexFromSignature_InvalidDigestHex(t *testing.T) {
	sig := make([]byte, 65)
	_, err := extractPublicKeyHexFromSignature("zzzz", sig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to recover public key") // due to digest decode failure
}

func TestExtractPublicKeyHexFromSignature_SignatureRecoveryFails(t *testing.T) {
	digest := []byte("a6eb5617ec3be5f0f523829e371ede989e8a3d15336a3030594d349fb14c92e8")
	digestHex := hex.EncodeToString(digest)

	badSig := make([]byte, 65)
	copy(badSig[:64], make([]byte, 64))
	badSig[64] = 99

	_, err := extractPublicKeyHexFromSignature(digestHex, badSig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to recover public key")
}
