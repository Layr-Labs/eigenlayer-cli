package admin

import (
	"context"
	"errors"
	"testing"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	"github.com/Layr-Labs/eigensdk-go/logging"
	gethcommon "github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

type mockRemoveAdminWriter struct {
	removeAdminFunc func(ctx context.Context, request elcontracts.RemoveAdminRequest) (*gethtypes.Receipt, error)
}

func (m *mockRemoveAdminWriter) RemoveAdmin(
	ctx context.Context,
	request elcontracts.RemoveAdminRequest,
) (*gethtypes.Receipt, error) {
	return m.removeAdminFunc(ctx, request)
}

func generateMockRemoveAdminWriter(
	receipt *gethtypes.Receipt,
	err error,
) func(logging.Logger, *removeAdminConfig) (RemoveAdminWriter, error) {
	return func(logger logging.Logger, config *removeAdminConfig) (RemoveAdminWriter, error) {
		return &mockRemoveAdminWriter{
			removeAdminFunc: func(ctx context.Context, request elcontracts.RemoveAdminRequest) (*gethtypes.Receipt, error) {
				return receipt, err
			},
		}, nil
	}
}

func TestRemoveCmd_Success(t *testing.T) {
	mockReceipt := &gethtypes.Receipt{
		TxHash: gethcommon.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
	}

	app := cli.NewApp()
	app.Commands = []*cli.Command{
		RemoveCmd(generateMockRemoveAdminWriter(mockReceipt, nil)),
	}

	args := []string{
		"TestRemoveCmd_Success",
		"remove-admin",
		"--account-address", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--admin-address", "0x1234567890abcdef1234567890abcdef12345678",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
		"--network", "holesky",
		"--ecdsa-private-key", "0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd",
	}

	err := app.Run(args)
	assert.NoError(t, err)
}

func TestRemoveCmd_GeneratorError(t *testing.T) {
	expectedError := "failed to create admin writer"
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		RemoveCmd(func(logger logging.Logger, config *removeAdminConfig) (RemoveAdminWriter, error) {
			return nil, errors.New(expectedError)
		}),
	}

	args := []string{
		"TestRemoveCmd_GeneratorError",
		"remove-admin",
		"--account-address", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--admin-address", "0x1234567890abcdef1234567890abcdef12345678",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
		"--network", "holesky",
		"--ecdsa-private-key", "0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd",
	}

	err := app.Run(args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedError)
}

func TestRemoveCmd_RemoveAdminError(t *testing.T) {
	expectedError := "error removing admin"
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		RemoveCmd(generateMockRemoveAdminWriter(nil, errors.New(expectedError))),
	}

	args := []string{
		"TestRemoveCmd_RemoveAdminError",
		"remove-admin",
		"--account-address", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--admin-address", "0x1234567890abcdef1234567890abcdef12345678",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
		"--network", "holesky",
		"--ecdsa-private-key", "0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd",
	}

	err := app.Run(args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedError)
}
