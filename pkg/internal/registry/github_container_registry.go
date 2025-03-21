package registry

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/static"
	protobundle "github.com/sigstore/protobuf-specs/gen/pb-go/bundle/v1"
	protocommon "github.com/sigstore/protobuf-specs/gen/pb-go/common/v1"
	sgbundle "github.com/sigstore/sigstore-go/pkg/bundle"
	"github.com/sigstore/sigstore-go/pkg/sign"
)

const (
	signatureTagFormat      = "sha256-%s.sig"
	locationTagFormat       = "%s:%s"
	sigstoreBundleMediaType = "application/vnd.dev.sigstore.bundle.v0.3+json"
)

type GithubContainerRegistry struct{}

func NewGithubContainerRegistry() ContainerRegistry {
	return GithubContainerRegistry{}
}

func (g GithubContainerRegistry) PushSignature(
	digestBytes []byte,
	signature []byte,
	publicKeyHex string,
	signerAddressHex string,
	tag name.Tag,
) error {
	b64Signature := base64.StdEncoding.EncodeToString(signature)

	annotations := map[string]string{
		EigenSignatureKey:     b64Signature,
		EigenPublicKey:        publicKeyHex,
		EigenSignerAddressKey: signerAddressHex,
	}

	sigBundle := newBundle(digestBytes, signature, publicKeyHex)
	bundleBytes, err := json.Marshal(sigBundle)
	layer := static.NewLayer(bundleBytes, sigstoreBundleMediaType)

	annotatedImg := mutate.Annotations(empty.Image, annotations).(v1.Image)
	finalImage, err := mutate.AppendLayers(annotatedImg, layer)
	if err != nil {
		return fmt.Errorf("failed to append layer: %w", err)
	}

	return remote.Write(tag, finalImage, remote.WithAuthFromKeychain(authn.DefaultKeychain))
}

func (g GithubContainerRegistry) TagSignature(location string, digest string) (name.Tag, error) {
	sigTag := fmt.Sprintf(signatureTagFormat, digest)
	fullRef := fmt.Sprintf(locationTagFormat, location, sigTag)
	return name.NewTag(fullRef)
}

func newBundle(digestBytes []byte, signature []byte, pkHex string) *sgbundle.Bundle {
	content := sign.PlainData{
		Data: digestBytes,
	}
	protoBundle := &protobundle.Bundle{MediaType: "application/vnd.dev.sigstore.bundle.v0.3+json"}
	protoBundle.VerificationMaterial = &protobundle.VerificationMaterial{
		Content: &protobundle.VerificationMaterial_PublicKey{
			PublicKey: &protocommon.PublicKeyIdentifier{
				Hint: pkHex,
			},
		},
	}
	content.Bundle(protoBundle, signature, digestBytes, protocommon.HashAlgorithm_SHA2_256)
	bundleResult, err := sgbundle.NewBundle(protoBundle)
	if err != nil {
		log.Fatalf("Failed to create bundle: %v", err)
	}
	return bundleResult
}
