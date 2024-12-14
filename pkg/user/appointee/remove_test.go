package appointee

import (
	"context"
	"errors"
	"testing"

	gethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	"github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/stretchr/testify/assert"

	"github.com/urfave/cli/v2"
)

type mockRemoveAppointeePermissionWriter struct {
	removePermissionFunc      func(ctx context.Context, request elcontracts.RemovePermissionRequest) (*gethtypes.Receipt, error)
	newRemovePermissionTxFunc func(request elcontracts.RemovePermissionRequest) (*gethtypes.Transaction, error)
}

func (m *mockRemoveAppointeePermissionWriter) RemovePermission(
	ctx context.Context,
	request elcontracts.RemovePermissionRequest,
) (*gethtypes.Receipt, error) {
	return m.removePermissionFunc(ctx, request)
}

func (m *mockRemoveAppointeePermissionWriter) NewRemovePermissionTx(
	request elcontracts.RemovePermissionRequest,
) (*gethtypes.Transaction, error) {
	return m.newRemovePermissionTxFunc(request)
}

func generateMockRemoveWriter(err error) func(logging.Logger, *removeConfig) (RemoveAppointeePermissionWriter, error) {
	return func(logger logging.Logger, config *removeConfig) (RemoveAppointeePermissionWriter, error) {
		return &mockRemoveAppointeePermissionWriter{
			removePermissionFunc: func(ctx context.Context, request elcontracts.RemovePermissionRequest) (*gethtypes.Receipt, error) {
				return &gethtypes.Receipt{}, err
			},
			newRemovePermissionTxFunc: func(request elcontracts.RemovePermissionRequest) (*gethtypes.Transaction, error) {
				return &gethtypes.Transaction{}, err
			},
		}, nil
	}
}

func TestRemoveCmd_Success(t *testing.T) {
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		RemoveCmd(generateMockRemoveWriter(nil)),
	}

	args := []string{
		"TestRemoveCmd_Success",
		"remove",
		"--account-address", "0x1234567890abcdef1234567890abcdef12345678",
		"--appointee-address", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--target-address", "0x9876543210fedcba9876543210fedcba98765432",
		"--selector", "0x1A2B3C4D",
		"--network", "holesky",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
		"--ecdsa-private-key", "0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		"--broadcast",
	}

	err := app.Run(args)
	assert.NoError(t, err)
}

func TestRemoveCmd_GeneratorError(t *testing.T) {
	expectedError := "failed to create permission writer"
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		RemoveCmd(func(logger logging.Logger, config *removeConfig) (RemoveAppointeePermissionWriter, error) {
			return nil, errors.New(expectedError)
		}),
	}

	args := []string{
		"TestRemoveCmd_GeneratorError",
		"remove",
		"--account-address", "0x1234567890abcdef1234567890abcdef12345678",
		"--appointee-address", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--target-address", "0x9876543210fedcba9876543210fedcba98765432",
		"--selector", "0x1A2B3C4D",
		"--network", "holesky",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
		"--ecdsa-private-key", "0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		"--broadcast",
	}

	err := app.Run(args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedError)
}
