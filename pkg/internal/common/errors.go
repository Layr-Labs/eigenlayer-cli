package common

import "errors"

var (
	ErrInvalidYamlFile = errors.New("invalid yaml file")
	ErrInvalidMetadata = errors.New("invalid metadata")
)
