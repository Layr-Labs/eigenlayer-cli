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

type mockListAdminsReader struct {
	listAdminsFunc func(ctx context.Context, userAddress gethcommon.Address) ([]gethcommon.Address, error)
}

func (m *mockListAdminsReader) ListAdmins(ctx context.Context, userAddress gethcommon.Address) ([]gethcommon.Address, error) {
	return m.listAdminsFunc(ctx, userAddress)
}

func generateMockListAdminsReader(admins []gethcommon.Address, err error) func(logging.Logger, *listAdminsConfig) (ListAdminsReader, error) {
	return func(logger logging.Logger, config *listAdminsConfig) (ListAdminsReader, error) {
		return &mockListAdminsReader{
			listAdminsFunc: func(ctx context.Context, userAddress gethcommon.Address) ([]gethcommon.Address, error) {
				return admins, err
			},
		}, nil
	}
}

func TestListCmd_Success(t *testing.T) {
	expectedAdmins := []gethcommon.Address{
		gethcommon.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"),
		gethcommon.HexToAddress("0xabcdef1234567890abcdef1234567890abcdef12"),
	}

	app := cli.NewApp()
	app.Commands = []*cli.Command{
		ListCmd(generateMockListAdminsReader(expectedAdmins, nil)),
	}

	args := []string{
		"TestListCmd_Success",
		"list-admins",
		"--account-address", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
		"--network", "holesky",
	}

	err := app.Run(args)
	assert.NoError(t, err)
}

func TestListCmd_NoAdmins(t *testing.T) {
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		ListCmd(generateMockListAdminsReader([]gethcommon.Address{}, nil)),
	}

	args := []string{
		"TestListCmd_NoAdmins",
		"list-admins",
		"--account-address", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
		"--network", "holesky",
	}

	err := app.Run(args)
	assert.NoError(t, err)
}

func TestListCmd_GeneratorError(t *testing.T) {
	expectedError := "failed to create admin reader"
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		ListCmd(func(logger logging.Logger, config *listAdminsConfig) (ListAdminsReader, error) {
			return nil, errors.New(expectedError)
		}),
	}

	args := []string{
		"TestListCmd_GeneratorError",
		"list-admins",
		"--account-address", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
		"--network", "holesky",
	}

	err := app.Run(args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedError)
}

func TestListCmd_ListAdminsError(t *testing.T) {
	expectedError := "failed to fetch admins"
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		ListCmd(generateMockListAdminsReader(nil, errors.New(expectedError))),
	}

	args := []string{
		"TestListCmd_ListAdminsError",
		"list-admins",
		"--account-address", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
		"--network", "holesky",
	}

	err := app.Run(args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedError)
}
