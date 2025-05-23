package appointee

import (
	"context"
	"errors"
	"testing"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	"github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

type mockSetAppointeePermissionWriter struct {
	setPermissionFunc      func(ctx context.Context, request elcontracts.SetPermissionRequest) (*gethtypes.Receipt, error)
	newSetPermissionTxFunc func(txOpts *bind.TransactOpts, request elcontracts.SetPermissionRequest) (*gethtypes.Transaction, error)
}

func (m *mockSetAppointeePermissionWriter) SetPermission(
	ctx context.Context,
	request elcontracts.SetPermissionRequest,
) (*gethtypes.Receipt, error) {
	return m.setPermissionFunc(ctx, request)
}

func (m *mockSetAppointeePermissionWriter) NewSetPermissionTx(
	txOpts *bind.TransactOpts,
	request elcontracts.SetPermissionRequest,
) (*gethtypes.Transaction, error) {
	return m.newSetPermissionTxFunc(txOpts, request)
}

func generateMockSetWriter(err error) func(logging.Logger, *setConfig) (SetAppointeePermissionWriter, error) {
	return func(logger logging.Logger, config *setConfig) (SetAppointeePermissionWriter, error) {
		return &mockSetAppointeePermissionWriter{
			setPermissionFunc: func(ctx context.Context, request elcontracts.SetPermissionRequest) (*gethtypes.Receipt, error) {
				return &gethtypes.Receipt{}, err
			},
			newSetPermissionTxFunc: func(txOpts *bind.TransactOpts, request elcontracts.SetPermissionRequest) (*gethtypes.Transaction, error) {
				return &gethtypes.Transaction{}, err
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
		"--broadcast",
	}

	err := app.Run(args)
	assert.NoError(t, err)
}

func TestSetCmd_GeneratorError(t *testing.T) {
	expectedError := "failed to create permission writer"
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		SetCmd(func(logger logging.Logger, config *setConfig) (SetAppointeePermissionWriter, error) {
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
		"--broadcast",
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
		"--broadcast",
	}

	err := app.Run(args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedError)
}
