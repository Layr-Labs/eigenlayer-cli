package admin

import (
	"context"
	"errors"
	"testing"

	"github.com/Layr-Labs/eigensdk-go/logging"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

type mockIsPendingAdminReader struct {
	isPendingAdminFunc func(ctx context.Context, accountAddress gethcommon.Address, pendingAdminAddress gethcommon.Address) (bool, error)
}

func (m *mockIsPendingAdminReader) IsPendingAdmin(
	ctx context.Context,
	accountAddress gethcommon.Address,
	pendingAdminAddress gethcommon.Address,
) (bool, error) {
	return m.isPendingAdminFunc(ctx, accountAddress, pendingAdminAddress)
}

func generateMockIsPendingAdminReader(result bool, err error) func(logging.Logger, *isPendingAdminConfig) (IsPendingAdminReader, error) {
	return func(logger logging.Logger, config *isPendingAdminConfig) (IsPendingAdminReader, error) {
		return &mockIsPendingAdminReader{
			isPendingAdminFunc: func(ctx context.Context, accountAddress gethcommon.Address, pendingAdminAddress gethcommon.Address) (bool, error) {
				return result, err
			},
		}, nil
	}
}

func TestIsPendingCmd_Success(t *testing.T) {
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		IsPendingCmd(generateMockIsPendingAdminReader(true, nil)),
	}

	args := []string{
		"TestIsPendingCmd_Success",
		"is-pending-admin",
		"--account-address", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--pending-admin-address", "0x1234567890abcdef1234567890abcdef12345678",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
		"--network", "holesky",
	}

	err := app.Run(args)
	assert.NoError(t, err)
}

func TestIsPendingCmd_NotPending(t *testing.T) {
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		IsPendingCmd(generateMockIsPendingAdminReader(false, nil)),
	}

	args := []string{
		"TestIsPendingCmd_NotPending",
		"is-pending-admin",
		"--account-address", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--pending-admin-address", "0x1234567890abcdef1234567890abcdef12345678",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
		"--network", "holesky",
	}

	err := app.Run(args)
	assert.NoError(t, err)
}

func TestIsPendingCmd_GeneratorError(t *testing.T) {
	expectedError := "failed to create pending admin reader"
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		IsPendingCmd(func(logger logging.Logger, config *isPendingAdminConfig) (IsPendingAdminReader, error) {
			return nil, errors.New(expectedError)
		}),
	}

	args := []string{
		"TestIsPendingCmd_GeneratorError",
		"is-pending-admin",
		"--account-address", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--pending-admin-address", "0x1234567890abcdef1234567890abcdef12345678",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
		"--network", "holesky",
	}

	err := app.Run(args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedError)
}

func TestIsPendingCmd_IsPendingAdminError(t *testing.T) {
	expectedError := "error checking pending admin status"
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		IsPendingCmd(generateMockIsPendingAdminReader(false, errors.New(expectedError))),
	}

	args := []string{
		"TestIsPendingCmd_IsPendingAdminError",
		"is-pending-admin",
		"--account-address", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--pending-admin-address", "0x1234567890abcdef1234567890abcdef12345678",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
		"--network", "holesky",
	}

	err := app.Run(args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedError)
}
