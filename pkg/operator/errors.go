package operator

import "errors"

var (
	ErrInvalidNumberOfArgs = errors.New("invalid number of arguments")
	ErrInvalidYamlFile     = errors.New("invalid yaml file")
)
