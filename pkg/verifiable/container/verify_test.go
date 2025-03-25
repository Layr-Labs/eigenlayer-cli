package container

import (
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"testing"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/registry"
	eigensdkLogger "github.com/Layr-Labs/eigensdk-go/logging"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

type mockLogger struct {
	eigensdkLogger.Logger
	logs []string
}

func (m *mockLogger) Infof(format string, args ...interface{}) {
	m.logs = append(m.logs, fmt.Sprintf(format, args...))
}

type mockRegistry struct {
	tag        name.Tag
	signature  string
	publicKey  string
	tagErr     error
	sigCompErr error
	pushErr    error

	pushedDigest     []byte
	pushedSignature  []byte
	pushedPublicKey  string
	pushedSignerAddr string
	pushedTag        name.Tag
}

func (m *mockRegistry) PushSignature(
	digestBytes []byte,
	signature []byte,
	publicKeyHex string,
	signerAddressHex string,
	tag name.Tag,
) error {
	m.pushedDigest = digestBytes
	m.pushedSignature = signature
	m.pushedPublicKey = publicKeyHex
	m.pushedSignerAddr = signerAddressHex
	m.pushedTag = tag
	return m.pushErr
}

func (m *mockRegistry) GetSignatureTag(registry string, digest string) (name.Tag, error) {
	return m.tag, m.tagErr
}

func (m *mockRegistry) GetSignatureComponents(tag name.Tag) (string, string, error) {
	return m.signature, m.publicKey, m.sigCompErr
}

func TestVerifySignatureCmd_Execute_Success(t *testing.T) {
	privateKey, _ := crypto.GenerateKey()
	digest := "a6eb5617ec3be5f0f523829e371ede989e8a3d15336a3030594d349fb14c92e8"
	digestBytes, err := hex.DecodeString(digest)
	assert.NoError(t, err)
	signature, err := crypto.Sign(digestBytes, privateKey)
	assert.NoError(t, err)
	sigBase64 := base64.StdEncoding.EncodeToString(signature)
	pubHex := hex.EncodeToString(crypto.FromECDSAPub(&privateKey.PublicKey))

	tag, err := name.NewTag("ghcr.io/user/container:sha256-" + digest)
	assert.NoError(t, err)

	mockReg := &mockRegistry{
		tag:       tag,
		signature: sigBase64,
		publicKey: pubHex,
	}

	set := flagSet(map[string]string{
		containerDigestFlag.Name:    digest,
		repositoryLocationFlag.Name: "ghcr.io/testing/registry-name",
	})
	ctx := cli.NewContext(nil, set, nil)

	cmd := verifySignatureCmd{registry: mockReg}
	err = cmd.Execute(ctx)

	assert.NoError(t, err)
}

func TestVerifySignatureCmd_Execute_VerificationFails(t *testing.T) {
	privateKey, _ := crypto.GenerateKey()
	privateKeyHex := hex.EncodeToString(crypto.FromECDSA(privateKey))
	hash := crypto.Keccak256Hash([]byte("mismatch"))
	signature, _ := crypto.Sign(hash.Bytes(), privateKey)
	sigBase64 := base64.StdEncoding.EncodeToString(signature)
	digestHex := hex.EncodeToString(hash.Bytes())

	otherKey, _ := crypto.GenerateKey()
	otherPubHex := hex.EncodeToString(crypto.FromECDSAPub(&otherKey.PublicKey))

	tag, err := name.NewTag("ghcr.io/user/container:sha256-" + digestHex)
	assert.NoError(t, err)

	mockReg := &mockRegistry{
		tag:       tag,
		signature: sigBase64,
		publicKey: otherPubHex,
	}

	set := flagSet(map[string]string{
		containerDigestFlag.Name:       digestHex,
		repositoryLocationFlag.Name:    "ghcr.io/testing/registry-name",
		flags.EcdsaPrivateKeyFlag.Name: privateKeyHex,
	})
	ctx := cli.NewContext(nil, set, nil)

	cmd := verifySignatureCmd{registry: mockReg}
	err = cmd.Execute(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "verification failed")
}

func TestVerifySignatureCmd_Execute_GetSignatureTagError(t *testing.T) {
	digestHex := hex.EncodeToString(crypto.Keccak256([]byte("fail tag")))
	mockReg := &mockRegistry{
		tagErr: errors.New("failed to fetch tag"),
	}

	set := flag.NewFlagSet("test", 0)
	_ = set.Set(repositoryLocationFlag.Name, "ghcr.io/user/container")
	_ = set.Set(containerDigestFlag.Name, digestHex)
	ctx := cli.NewContext(nil, set, nil)

	cmd := verifySignatureCmd{registry: mockReg}
	err := cmd.Execute(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch tag")
}

func TestVerifySignatureCmd_Execute_GetSignatureComponentsError(t *testing.T) {
	digestHex := hex.EncodeToString(crypto.Keccak256([]byte("fail components")))
	tag, _ := name.NewTag("ghcr.io/user/container:sha256-" + digestHex)

	mockReg := &mockRegistry{
		tag:        tag,
		sigCompErr: errors.New("component retrieval failed"),
	}

	set := flag.NewFlagSet("test", 0)
	_ = set.Set(repositoryLocationFlag.Name, "ghcr.io/user/container")
	_ = set.Set(containerDigestFlag.Name, digestHex)
	ctx := cli.NewContext(nil, set, nil)

	cmd := verifySignatureCmd{registry: mockReg}
	err := cmd.Execute(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "component retrieval failed")
}

func TestVerifySignature_ValidMatch(t *testing.T) {
	logger := &mockLogger{}
	privateKey, err := crypto.GenerateKey()
	assert.NoError(t, err)

	message := []byte("to verify")
	hash := crypto.Keccak256Hash(message)

	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	assert.NoError(t, err)

	sigBase64 := base64.StdEncoding.EncodeToString(signature)
	digestHex := hex.EncodeToString(hash.Bytes())
	publicKeyHex := hex.EncodeToString(crypto.FromECDSAPub(&privateKey.PublicKey))

	valid := verifySignature(logger, digestHex, sigBase64, publicKeyHex)
	assert.True(t, valid)
	assert.Contains(t, logger.logs[0], "Signature is valid and matches")
}

func TestVerifySignature_ValidButMismatchedPublicKey(t *testing.T) {
	logger := &mockLogger{}
	privateKey, _ := crypto.GenerateKey()
	hash := crypto.Keccak256Hash([]byte("mismatch"))

	signature, _ := crypto.Sign(hash.Bytes(), privateKey)
	sigBase64 := base64.StdEncoding.EncodeToString(signature)
	digestHex := hex.EncodeToString(hash.Bytes())

	otherKey, _ := crypto.GenerateKey()
	otherPubHex := hex.EncodeToString(crypto.FromECDSAPub(&otherKey.PublicKey))

	valid := verifySignature(logger, digestHex, sigBase64, otherPubHex)
	assert.False(t, valid)
	assert.Contains(t, logger.logs[0], "Signature is valid but does not match")
}

func TestVerifySignature_InvalidBase64Signature(t *testing.T) {
	logger := &mockLogger{}
	digest := hex.EncodeToString(crypto.Keccak256([]byte("invalid")))

	valid := verifySignature(logger, digest, "%invalid%", "publicKey")
	assert.False(t, valid)
	assert.Contains(t, logger.logs[0], "Failed to recover public key")
}

func TestDefaultFlags(t *testing.T) {
	expectedFlags := verifiableContainerCmdFlags()

	verifyCmd := NewVerifyContainerCmd(registry.OciRegistryController{})
	actualFlags := verifyCmd.Flags

	for _, expected := range expectedFlags {
		found := false
		for _, actual := range actualFlags {
			found = found || expected == actual
		}
		assert.True(t, found, "Default flag not found in command.")
	}
	assert.Equal(t, len(expectedFlags), len(actualFlags), "Extra flag found in verify command.")
}

func TestGetSignaturePublicKey_ValidInput(t *testing.T) {
	privateKey, err := crypto.GenerateKey()
	assert.NoError(t, err)

	message := []byte("mock message")
	hash := crypto.Keccak256Hash(message)

	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	assert.NoError(t, err)

	signatureBase64 := base64.StdEncoding.EncodeToString(signature)
	digestHex := hex.EncodeToString(hash.Bytes())

	pubKey, err := getSignaturePublicKey(signatureBase64, digestHex)
	assert.NoError(t, err)
	assert.NotNil(t, pubKey)

	expectedPubKey := privateKey.Public().(*ecdsa.PublicKey)
	assert.Equal(t, expectedPubKey.X, pubKey.X)
	assert.Equal(t, expectedPubKey.Y, pubKey.Y)
}

func TestGetSignaturePublicKey_InvalidBase64(t *testing.T) {
	_, err := getSignaturePublicKey("!invalid-base64!", "abc123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error decoding signature to bytes")
}

func TestGetSignaturePublicKey_InvalidSignatureLength(t *testing.T) {
	shortSig := make([]byte, 64)
	sigBase64 := base64.StdEncoding.EncodeToString(shortSig)
	_, err := getSignaturePublicKey(sigBase64, "invalidDigest")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid signature length")
}

func TestGetSignaturePublicKey_InvalidDigest(t *testing.T) {
	privateKey, _ := crypto.GenerateKey()
	hash := crypto.Keccak256Hash([]byte("msg"))
	signature, _ := crypto.Sign(hash.Bytes(), privateKey)
	sigBase64 := base64.StdEncoding.EncodeToString(signature)

	_, err := getSignaturePublicKey(sigBase64, "invalidHex")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error decoding digest to bytes")
}

func flagSet(values map[string]string) *flag.FlagSet {
	set := flag.NewFlagSet("test", 0)
	for key, value := range values {
		set.String(key, value, "")
		_ = set.Set(key, value)
	}
	return set
}
