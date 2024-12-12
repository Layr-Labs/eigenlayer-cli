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

type mockRemovePendingAdminWriter struct {
	removePendingAdminFunc func(ctx context.Context, request elcontracts.RemovePendingAdminRequest) (*gethtypes.Receipt, error)
}

func (m *mockRemovePendingAdminWriter) RemovePendingAdmin(
	ctx context.Context,
	request elcontracts.RemovePendingAdminRequest,
) (*gethtypes.Receipt, error) {
	return m.removePendingAdminFunc(ctx, request)
}

func generateMockRemovePendingAdminWriter(
	receipt *gethtypes.Receipt,
	err error,
) func(logging.Logger, *removePendingAdminConfig) (RemovePendingAdminWriter, error) {
	return func(logger logging.Logger, config *removePendingAdminConfig) (RemovePendingAdminWriter, error) {
		return &mockRemovePendingAdminWriter{
			removePendingAdminFunc: func(ctx context.Context, request elcontracts.RemovePendingAdminRequest) (*gethtypes.Receipt, error) {
				return receipt, err
			},
		}, nil
	}
}

func TestRemovePendingCmd_Success(t *testing.T) {
	mockReceipt := &gethtypes.Receipt{
		TxHash: gethcommon.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
	}

	app := cli.NewApp()
	app.Commands = []*cli.Command{
		RemovePendingCmd(generateMockRemovePendingAdminWriter(mockReceipt, nil)),
	}

	args := []string{
		"TestRemovePendingCmd_Success",
		"remove-pending-admin",
		"--account-address", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--admin-address", "0x1234567890abcdef1234567890abcdef12345678",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
		"--network", "holesky",
		"--ecdsa-private-key", "0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd",
	}

	err := app.Run(args)
	assert.NoError(t, err)
}

func TestRemovePendingCmd_GeneratorError(t *testing.T) {
	expectedError := "failed to create admin writer"
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		RemovePendingCmd(
			func(logger logging.Logger, config *removePendingAdminConfig) (RemovePendingAdminWriter, error) {
				return nil, errors.New(expectedError)
			},
		),
	}

	args := []string{
		"TestRemovePendingCmd_GeneratorError",
		"remove-pending-admin",
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

func TestRemovePendingCmd_RemovePendingError(t *testing.T) {
	expectedError := "error removing pending admin"
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		RemovePendingCmd(generateMockRemovePendingAdminWriter(nil, errors.New(expectedError))),
	}

	args := []string{
		"TestRemovePendingCmd_RemovePendingError",
		"remove-pending-admin",
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
