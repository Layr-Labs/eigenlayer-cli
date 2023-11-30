package keys

import "errors"

var (
	ErrInvalidNumberOfArgs           = errors.New("invalid number of arguments")
	ErrEmptyKeyName                  = errors.New("key name cannot be empty")
	ErrEmptyPrivateKey               = errors.New("private key cannot be empty")
	ErrKeyContainsWhitespaces        = errors.New("key name cannot contain spaces")
	ErrPrivateKeyContainsWhitespaces = errors.New("private key cannot contain spaces")
	ErrInvalidKeyType                = errors.New("invalid key type. key type must be either 'ecdsa' or 'bls'")
	ErrInvalidPassword               = errors.New("invalid password")
	ErrInvalidHexPrivateKey          = errors.New("invalid hex private key")
)
