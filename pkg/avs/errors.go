package avs

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidNumberOfArgs         = errors.New("invalid number of arguments")
	ErrEmptyArgValue               = errors.New("argument value cannot be empty")
	ErrArgValueContainsWhitespaces = errors.New("argument value cannot contain spaces")
	ErrFailedToLoadFile            = errors.New("failed to load file")
	ErrInvalidConfigFile           = errors.New("invalid configuration file, should have yaml or yml extension")
	ErrInvalidConfig               = errors.New("value not found for given configuration key")
	ErrInvalidConfigType           = errors.New("unsupported configuration value type")
)

func failed(err interface{}) error {
	switch t := err.(type) {
	case error:
		_ = t
		tokens := strings.Split(err.(error).Error(), ":")
		return fmt.Errorf("failed: %s", strings.TrimSpace(tokens[len(tokens)-1]))
	case string:
		tokens := strings.Split(err.(string), ":")
		return fmt.Errorf("failed: %s", strings.TrimSpace(tokens[len(tokens)-1]))
	}

	return nil
}
