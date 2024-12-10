package appointee

import (
	"context"
	"errors"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	"testing"

	"github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

type mockSetUserPermissionWriter struct {
	setPermissionFunc func(ctx context.Context, request elcontracts.SetPermissionRequest) error
}

func (m *mockSetUserPermissionWriter) SetPermission(
	ctx context.Context,
	request elcontracts.SetPermissionRequest,
) error {
	return m.setPermissionFunc(ctx, request)
}

func generateMockSetWriter(err error) func(logging.Logger, *setConfig) (SetUserPermissionWriter, error) {
	return func(logger logging.Logger, config *setConfig) (SetUserPermissionWriter, error) {
		return &mockSetUserPermissionWriter{
			setPermissionFunc: func(ctx context.Context, request elcontracts.SetPermissionRequest) error {
				return err
			},
		}, nil
	}
}

func TestSetCmd_Success(t *testing.T) {
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		SetCmd(generateMockSetWriter(nil)),
	}

	args := []string{
		"TestSetCmd_Success",
		"set",
		"--account-address", "0x1234567890abcdef1234567890abcdef12345678",
		"--appointee-address", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--target-address", "0x9876543210fedcba9876543210fedcba98765432",
		"--selector", "0x1A2B3C4D",
		"--network", "holesky",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
		"--ecdsa-private-key", "0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd",
	}

	err := app.Run(args)
	assert.NoError(t, err)
}

func TestSetCmd_GeneratorError(t *testing.T) {
	expectedError := "failed to create permission writer"
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		SetCmd(func(logger logging.Logger, config *setConfig) (SetUserPermissionWriter, error) {
			return nil, errors.New(expectedError)
		}),
	}

	args := []string{
		"TestSetCmd_GeneratorError",
		"set",
		"--account-address", "0x1234567890abcdef1234567890abcdef12345678",
		"--appointee-address", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--target-address", "0x9876543210fedcba9876543210fedcba98765432",
		"--selector", "0x1A2B3C4D",
		"--network", "holesky",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
		"--ecdsa-private-key", "0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd",
	}

	err := app.Run(args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedError)
}

func TestSetCmd_SetPermissionError(t *testing.T) {
	expectedError := "error setting permission"
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		SetCmd(generateMockSetWriter(errors.New(expectedError))),
	}

	args := []string{
		"TestSetCmd_SetPermissionError",
		"set",
		"--account-address", "0x1234567890abcdef1234567890abcdef12345678",
		"--appointee-address", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--target-address", "0x9876543210fedcba9876543210fedcba98765432",
		"--selector", "0x1A2B3C4D",
		"--network", "holesky",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
		"--ecdsa-private-key", "0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd",
	}

	err := app.Run(args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedError)
}
