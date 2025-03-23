package registry

import (
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

type RegistryClient interface {
	Push(tag name.Tag, img v1.Image) error
	Get(tag name.Tag) (*remote.Descriptor, error)
}
