package verifiable

import (
	"encoding/hex"
	"fmt"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/types"
	"time"
)

type SignMessageConfig struct {
	SignerConfig       *types.SignerConfig
	RepositoryLocation string
	ImageId            string
}

type SignatureArtifact struct {
	ImageID   string `json:"image_id"`
	ImageSHA  string `json:"image_sha"`
	Signature string `json:"signature"`
	SignerPK  string `json:"signer_public_key"`
}

type OCISignatureArtifact struct {
	SchemaVersion int               `json:"schemaVersion"`
	MediaType     string            `json:"mediaType"`
	Config        OCIConfig         `json:"config"`
	Layers        []OCILayer        `json:"layers"`
	Annotations   map[string]string `json:"annotations"`
}

func NewOCISignatureArtifact(
	imageSha string,
	signedHash []byte,
	signerPkHex string,
	location string,
) OCISignatureArtifact {
	return OCISignatureArtifact{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.manifest.v1+json",
		Config: OCIConfig{
			MediaType: "application/vnd.oci.artifact.signature.v1+json",
			Size:      0,
			Digest:    "sha256:" + imageSha,
		},
		Layers: []OCILayer{
			{
				MediaType: "application/vnd.oci.artifact.signature.v1+json",
				Digest:    "sha256:" + hex.EncodeToString(signedHash),
				Size:      len(signedHash),
			},
		},
		Annotations: map[string]string{
			"org.opencontainers.image.created":     time.Now().UTC().Format(time.RFC3339),
			"org.opencontainers.image.description": "Signature artifact for verifying image integrity",
			"org.opencontainers.image.url":         fmt.Sprintf("docker://%s.sig", location),
			"io.github.signer.publickey":           signerPkHex,
		},
	}
}

type OCIConfig struct {
	MediaType string `json:"mediaType"`
	Size      int    `json:"size"`
	Digest    string `json:"digest"`
}

type OCILayer struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int    `json:"size"`
}
