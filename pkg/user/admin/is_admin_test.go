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

type mockIsAdminReader struct {
	isAdminFunc func(ctx context.Context, accountAddress gethcommon.Address, adminAddress gethcommon.Address) (bool, error)
}

func (m *mockIsAdminReader) IsAdmin(
	ctx context.Context,
	accountAddress gethcommon.Address,
	adminAddress gethcommon.Address,
) (bool, error) {
	return m.isAdminFunc(ctx, accountAddress, adminAddress)
}

func generateMockIsAdminReader(result bool, err error) func(logging.Logger, *isAdminConfig) (IsAdminReader, error) {
	return func(logger logging.Logger, config *isAdminConfig) (IsAdminReader, error) {
		return &mockIsAdminReader{
			isAdminFunc: func(ctx context.Context, accountAddress gethcommon.Address, adminAddress gethcommon.Address) (bool, error) {
				return result, err
			},
		}, nil
	}
}

func TestIsAdminCmd_Success(t *testing.T) {
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		IsAdminCmd(generateMockIsAdminReader(true, nil)),
	}

	args := []string{
		"TestIsAdminCmd_Success",
		"is-admin",
		"--account-address", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--caller-address", "0x1234567890abcdef1234567890abcdef12345678",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
		"--network", "holesky",
	}

	err := app.Run(args)
	assert.NoError(t, err)
}

func TestIsAdminCmd_NotAdmin(t *testing.T) {
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		IsAdminCmd(generateMockIsAdminReader(false, nil)),
	}

	args := []string{
		"TestIsAdminCmd_NotAdmin",
		"is-admin",
		"--account-address", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--caller-address", "0x1234567890abcdef1234567890abcdef12345678",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
		"--network", "holesky",
	}

	err := app.Run(args)
	assert.NoError(t, err)
}

func TestIsAdminCmd_GeneratorError(t *testing.T) {
	expectedError := "failed to create admin reader"
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		IsAdminCmd(func(logger logging.Logger, config *isAdminConfig) (IsAdminReader, error) {
			return nil, errors.New(expectedError)
		}),
	}

	args := []string{
		"TestIsAdminCmd_GeneratorError",
		"is-admin",
		"--account-address", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--caller-address", "0x1234567890abcdef1234567890abcdef12345678",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
		"--network", "holesky",
	}

	err := app.Run(args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedError)
}

func TestIsAdminCmd_IsAdminError(t *testing.T) {
	expectedError := "error checking admin status"
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		IsAdminCmd(generateMockIsAdminReader(false, errors.New(expectedError))),
	}

	args := []string{
		"TestIsAdminCmd_IsAdminError",
		"is-admin",
		"--account-address", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--caller-address", "0x1234567890abcdef1234567890abcdef12345678",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
		"--network", "holesky",
	}

	err := app.Run(args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedError)
}
