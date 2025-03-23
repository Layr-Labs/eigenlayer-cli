package registry

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
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

type OciRegistryController struct {
	registryClient RegistryClient
}

func NewOciRegistryController(client RegistryClient) RegistryController {
	return OciRegistryController{registryClient: client}
}

func (g OciRegistryController) PushSignature(
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

	return g.registryClient.Push(tag, finalImage)
}

func (g OciRegistryController) GetSignatureTag(location string, digest string) (name.Tag, error) {
	sigTag := fmt.Sprintf(signatureTagFormat, digest)
	fullRef := fmt.Sprintf(locationTagFormat, location, sigTag)
	return name.NewTag(fullRef)
}

func (g OciRegistryController) GetSignatureComponents(tag name.Tag) (string, string, error) {
	desc, err := g.registryClient.Get(tag)
	if err != nil {
		return "", "", err
	}

	var man v1.Manifest
	if err = json.Unmarshal(desc.Manifest, &man); err != nil {
		return "", "", err
	}
	annotations := man.Annotations
	signature, sigOk := annotations[EigenSignatureKey]
	if !sigOk {
		return "", "", fmt.Errorf("signature not found in annotations")
	}
	publicKey, keyOk := annotations[EigenPublicKey]
	if !keyOk {
		return "", "", fmt.Errorf("public key not found in annotations")
	}
	if len(publicKey) > 2 && publicKey[:2] == "0x" {
		publicKey = publicKey[2:]
	}
	return signature, publicKey, nil
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
