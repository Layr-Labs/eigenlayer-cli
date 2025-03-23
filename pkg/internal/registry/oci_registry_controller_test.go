package registry

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/stretchr/testify/assert"
)

type RegistryClientFunc struct {
	PushFunc func(name.Tag, v1.Image) error
	GetFunc  func(name.Tag) (*remote.Descriptor, error)
}

func (m RegistryClientFunc) Push(tag name.Tag, img v1.Image) error {
	return m.PushFunc(tag, img)
}

func (m RegistryClientFunc) Get(tag name.Tag) (*remote.Descriptor, error) {
	return m.GetFunc(tag)
}

func TestOciRegistryController_GetSignatureTag(t *testing.T) {
	controller := NewOciRegistryController(nil)

	digest := "a6eb5617ec3be5f0f523829e371ede989e8a3d15336a3030594d349fb14c92e8"
	location := "ghcr.io/user/container"

	tag, err := controller.GetSignatureTag(location, digest)
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%s:sha256-%s.sig", location, digest), tag.String())
}

func TestOciRegistryController_GetSignatureComponents(t *testing.T) {
	expectedSig := base64.StdEncoding.EncodeToString([]byte("toSign"))
	expectedPubKey := "publicKey"

	man := v1.Manifest{
		Annotations: map[string]string{
			EigenSignatureKey: expectedSig,
			EigenPublicKey:    expectedPubKey,
		},
	}
	manBytes, _ := json.Marshal(man)

	mockClient := RegistryClientFunc{
		GetFunc: func(tag name.Tag) (*remote.Descriptor, error) {
			return &remote.Descriptor{Manifest: manBytes}, nil
		},
		PushFunc: func(name.Tag, v1.Image) error {
			return nil
		},
	}

	controller := NewOciRegistryController(mockClient)
	tag, _ := name.NewTag("ghcr.io/test/image:tag")

	sig, pub, err := controller.GetSignatureComponents(tag)
	assert.NoError(t, err)
	assert.Equal(t, expectedSig, sig)
	assert.Equal(t, expectedPubKey, pub)
}

func TestOciRegistryController_PushSignature(t *testing.T) {
	var calledTag name.Tag
	var calledImage v1.Image

	mockClient := RegistryClientFunc{
		PushFunc: func(tag name.Tag, img v1.Image) error {
			calledTag = tag
			calledImage = img
			return nil
		},
		GetFunc: func(tag name.Tag) (*remote.Descriptor, error) {
			return nil, nil
		},
	}

	controller := NewOciRegistryController(mockClient)

	digest := crypto.Keccak256([]byte("test"))
	sig := make([]byte, 65)
	pubKey := "publicKey"
	signer := "signerAddress"
	tag, _ := name.NewTag("ghcr.io/test/image:sha256-containerHash.sig")

	err := controller.PushSignature(digest, sig, pubKey, signer, tag)
	assert.NoError(t, err)
	assert.Equal(t, tag, calledTag)

	cfg, err := calledImage.ConfigFile()
	assert.NoError(t, err)
	annotations := cfg.Config.Labels
	assert.Equal(t, base64.StdEncoding.EncodeToString(sig), annotations[EigenSignatureKey])
	assert.Equal(t, pubKey, annotations[EigenPublicKey])
	assert.Equal(t, signer, annotations[EigenSignerAddressKey])
}

func TestOciRegistryController_PushSignature_PushFails(t *testing.T) {
	mockClient := RegistryClientFunc{
		PushFunc: func(tag name.Tag, img v1.Image) error {
			return fmt.Errorf("registry unavailable")
		},
		GetFunc: func(tag name.Tag) (*remote.Descriptor, error) { return nil, nil },
	}

	controller := NewOciRegistryController(mockClient)
	tag, _ := name.NewTag("ghcr.io/test/image:tag")
	err := controller.PushSignature([]byte("data"), make([]byte, 65), "pk", "addr", tag)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "registry unavailable")
}

func TestOciRegistryController_GetSignatureComponents_FetchFails(t *testing.T) {
	mockClient := RegistryClientFunc{
		GetFunc: func(tag name.Tag) (*remote.Descriptor, error) {
			return nil, fmt.Errorf("failed to fetch")
		},
		PushFunc: func(tag name.Tag, img v1.Image) error { return nil },
	}

	controller := NewOciRegistryController(mockClient)
	tag, _ := name.NewTag("ghcr.io/test/image:tag")

	_, _, err := controller.GetSignatureComponents(tag)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch")
}

func TestOciRegistryController_GetSignatureComponents_MissingAnnotations(t *testing.T) {
	tests := []struct {
		name        string
		annotations map[string]string
		expectError string
	}{
		{
			name:        "missing signature",
			annotations: map[string]string{EigenPublicKey: "abcd"},
			expectError: "signature not found",
		},
		{
			name:        "missing public key",
			annotations: map[string]string{EigenSignatureKey: "xyz"},
			expectError: "public key not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			man := v1.Manifest{Annotations: tt.annotations}
			manBytes, _ := json.Marshal(man)

			mockClient := RegistryClientFunc{
				GetFunc: func(tag name.Tag) (*remote.Descriptor, error) {
					return &remote.Descriptor{Manifest: manBytes}, nil
				},
				PushFunc: func(tag name.Tag, img v1.Image) error { return nil },
			}

			controller := NewOciRegistryController(mockClient)
			tag, _ := name.NewTag("ghcr.io/test/image:tag")

			_, _, err := controller.GetSignatureComponents(tag)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectError)
		})
	}
}
