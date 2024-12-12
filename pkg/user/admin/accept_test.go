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

type mockAcceptAdminWriter struct {
	acceptAdminFunc func(ctx context.Context, request elcontracts.AcceptAdminRequest) (*gethtypes.Receipt, error)
}

func (m *mockAcceptAdminWriter) AcceptAdmin(
	ctx context.Context,
	request elcontracts.AcceptAdminRequest,
) (*gethtypes.Receipt, error) {
	return m.acceptAdminFunc(ctx, request)
}

func generateMockAcceptAdminWriter(
	receipt *gethtypes.Receipt,
	err error,
) func(logging.Logger, *acceptAdminConfig) (AcceptAdminWriter, error) {
	return func(logger logging.Logger, config *acceptAdminConfig) (AcceptAdminWriter, error) {
		return &mockAcceptAdminWriter{
			acceptAdminFunc: func(ctx context.Context, request elcontracts.AcceptAdminRequest) (*gethtypes.Receipt, error) {
				return receipt, err
			},
		}, nil
	}
}

func TestAcceptCmd_Success(t *testing.T) {
	mockReceipt := &gethtypes.Receipt{
		TxHash: gethcommon.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
	}

	app := cli.NewApp()
	app.Commands = []*cli.Command{
		AcceptCmd(generateMockAcceptAdminWriter(mockReceipt, nil)),
	}

	args := []string{
		"TestAcceptCmd_Success",
		"accept-admin",
		"--account-address", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
		"--network", "holesky",
		"--ecdsa-private-key", "0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd",
	}

	err := app.Run(args)
	assert.NoError(t, err)
}

func TestAcceptCmd_GeneratorError(t *testing.T) {
	expectedError := "failed to create admin writer"
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		AcceptCmd(func(logger logging.Logger, config *acceptAdminConfig) (AcceptAdminWriter, error) {
			return nil, errors.New(expectedError)
		}),
	}

	args := []string{
		"TestAcceptCmd_GeneratorError",
		"accept-admin",
		"--account-address", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
		"--network", "holesky",
		"--ecdsa-private-key", "0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd",
	}

	err := app.Run(args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedError)
}

func TestAcceptCmd_AcceptAdminError(t *testing.T) {
	expectedError := "error accepting admin"
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		AcceptCmd(generateMockAcceptAdminWriter(nil, errors.New(expectedError))),
	}

	args := []string{
		"TestAcceptCmd_AcceptAdminError",
		"accept-admin",
		"--account-address", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
		"--network", "holesky",
		"--ecdsa-private-key", "0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd",
	}

	err := app.Run(args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedError)
}
