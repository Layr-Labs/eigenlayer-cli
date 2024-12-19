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

type mockListPendingAdminsReader struct {
	listPendingAdminsFunc func(ctx context.Context, userAddress gethcommon.Address) ([]gethcommon.Address, error)
}

func (m *mockListPendingAdminsReader) ListPendingAdmins(
	ctx context.Context,
	userAddress gethcommon.Address,
) ([]gethcommon.Address, error) {
	return m.listPendingAdminsFunc(ctx, userAddress)
}

func generateMockListPendingAdminsReader(
	admins []gethcommon.Address,
	err error,
) func(logging.Logger, *listPendingAdminsConfig) (ListPendingAdminsReader, error) {
	return func(logger logging.Logger, config *listPendingAdminsConfig) (ListPendingAdminsReader, error) {
		return &mockListPendingAdminsReader{
			listPendingAdminsFunc: func(ctx context.Context, userAddress gethcommon.Address) ([]gethcommon.Address, error) {
				return admins, err
			},
		}, nil
	}
}

func TestListPendingCmd_Success(t *testing.T) {
	expectedAdmins := []gethcommon.Address{
		gethcommon.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"),
		gethcommon.HexToAddress("0xabcdef1234567890abcdef1234567890abcdef12"),
	}

	app := cli.NewApp()
	app.Commands = []*cli.Command{
		ListPendingCmd(generateMockListPendingAdminsReader(expectedAdmins, nil)),
	}

	args := []string{
		"TestListPendingCmd_Success",
		"list-pending-admins",
		"--account-address", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
		"--network", "holesky",
		"--permission-controller-address", "0xe4dB7125ef7a9D99F809B6b7788f75c8D84d8455",
	}

	err := app.Run(args)
	assert.NoError(t, err)
}

func TestListPendingCmd_GeneratorError(t *testing.T) {
	expectedError := "failed to create pending admins reader"
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		ListPendingCmd(func(logger logging.Logger, config *listPendingAdminsConfig) (ListPendingAdminsReader, error) {
			return nil, errors.New(expectedError)
		}),
	}

	args := []string{
		"TestListPendingCmd_GeneratorError",
		"list-pending-admins",
		"--account-address", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
		"--network", "holesky",
		"--permission-controller-address", "0xe4dB7125ef7a9D99F809B6b7788f75c8D84d8455",
	}

	err := app.Run(args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedError)
}

func TestListPendingCmd_ListPendingError(t *testing.T) {
	expectedError := "failed to fetch pending admins"
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		ListPendingCmd(generateMockListPendingAdminsReader(nil, errors.New(expectedError))),
	}

	args := []string{
		"TestListPendingCmd_ListPendingError",
		"list-pending-admins",
		"--account-address", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
		"--network", "holesky",
		"--permission-controller-address", "0xe4dB7125ef7a9D99F809B6b7788f75c8D84d8455",
	}

	err := app.Run(args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedError)
}
