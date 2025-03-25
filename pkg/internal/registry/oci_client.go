package registry

import (
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

type OciClient struct{}

func (o OciClient) Push(tag name.Tag, img v1.Image) error {
	err := remote.Write(tag, img, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return err
	}
	return nil
}

func (o OciClient) Get(tag name.Tag) (*remote.Descriptor, error) {
	descriptor, err := remote.Get(tag)
	if err != nil {
		return nil, err
	}
	return descriptor, nil
}
