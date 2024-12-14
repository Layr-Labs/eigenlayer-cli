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

type mockAddPendingAdminWriter struct {
	addPendingAdminFunc func(ctx context.Context, request elcontracts.AddPendingAdminRequest) (*gethtypes.Receipt, error)
}

func (m *mockAddPendingAdminWriter) AddPendingAdmin(
	ctx context.Context,
	request elcontracts.AddPendingAdminRequest,
) (*gethtypes.Receipt, error) {
	return m.addPendingAdminFunc(ctx, request)
}

func generateMockAddPendingAdminWriter(
	receipt *gethtypes.Receipt,
	err error,
) func(logging.Logger, *addPendingAdminConfig) (AddPendingAdminWriter, error) {
	return func(logger logging.Logger, config *addPendingAdminConfig) (AddPendingAdminWriter, error) {
		return &mockAddPendingAdminWriter{
			addPendingAdminFunc: func(ctx context.Context, request elcontracts.AddPendingAdminRequest) (*gethtypes.Receipt, error) {
				return receipt, err
			},
		}, nil
	}
}

func TestAddPendingCmd_Success(t *testing.T) {
	mockReceipt := &gethtypes.Receipt{
		TxHash: gethcommon.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
	}

	app := cli.NewApp()
	app.Commands = []*cli.Command{
		AddPendingCmd(generateMockAddPendingAdminWriter(mockReceipt, nil)),
	}

	args := []string{
		"TestAddPendingCmd_Success",
		"add-pending-admin",
		"--account-address", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--admin-address", "0x1234567890abcdef1234567890abcdef12345678",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
		"--network", "holesky",
		"--ecdsa-private-key", "0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd",
	}

	err := app.Run(args)
	assert.NoError(t, err)
}

func TestAddPendingCmd_GeneratorError(t *testing.T) {
	expectedError := "failed to create admin writer"
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		AddPendingCmd(func(logger logging.Logger, config *addPendingAdminConfig) (AddPendingAdminWriter, error) {
			return nil, errors.New(expectedError)
		}),
	}

	args := []string{
		"TestAddPendingCmd_GeneratorError",
		"add-pending-admin",
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

func TestAddPendingCmd_AddPendingError(t *testing.T) {
	expectedError := "error adding pending admin"
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		AddPendingCmd(generateMockAddPendingAdminWriter(nil, errors.New(expectedError))),
	}

	args := []string{
		"TestAddPendingCmd_AddPendingError",
		"add-pending-admin",
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