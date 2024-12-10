package appointee

import (
	"context"
	"errors"
	"testing"

	"github.com/Layr-Labs/eigensdk-go/logging"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/urfave/cli/v2"
)

type mockListPermissionsReader struct {
	listPermissionsFunc func(
		ctx context.Context,
		appointed gethcommon.Address,
		userAddress gethcommon.Address,
	) ([]gethcommon.Address, [][4]byte, error)
}

func newMockListPermissionsReader(
	users []gethcommon.Address,
	permissions [][4]byte,
	err error,
) *mockListPermissionsReader {
	return &mockListPermissionsReader{
		listPermissionsFunc: func(ctx context.Context, appointed, userAddress gethcommon.Address) ([]gethcommon.Address, [][4]byte, error) {
			return users, permissions, err
		},
	}
}

func (m *mockListPermissionsReader) ListUserPermissions(
	ctx context.Context,
	appointed gethcommon.Address,
	userAddress gethcommon.Address,
) ([]gethcommon.Address, [][4]byte, error) {
	return m.listPermissionsFunc(ctx, appointed, userAddress)
}

func generateMockListPermissionsReader(
	users []gethcommon.Address,
	permissions [][4]byte,
	err error,
) func(logging.Logger, *listUserPermissionsConfig) (PermissionsReader, error) {
	return func(logger logging.Logger, config *listUserPermissionsConfig) (PermissionsReader, error) {
		return newMockListPermissionsReader(users, permissions, err), nil
	}
}

func TestListPermissions_Success(t *testing.T) {
	expectedUsers := []gethcommon.Address{
		gethcommon.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"),
	}
	expectedPermissions := [][4]byte{
		{0x1A, 0x2B, 0x3C, 0x4D},
	}

	app := cli.NewApp()
	app.Commands = []*cli.Command{
		ListPermissionsCmd(generateMockListPermissionsReader(expectedUsers, expectedPermissions, nil)),
	}

	args := []string{
		"TestListPermissions_Success",
		"list-permissions",
		"--account-address", "0x1234567890abcdef1234567890abcdef12345678",
		"--appointee-address", "0x9876543210fedcba9876543210fedcba98765432",
		"--network", "holesky",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
	}

	err := app.Run(args)
	assert.NoError(t, err)
}

func TestListPermissions_ReaderError(t *testing.T) {
	expectedError := "Error fetching permissions"
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		ListPermissionsCmd(generateMockListPermissionsReader(nil, nil, errors.New(expectedError))),
	}

	args := []string{
		"TestListPermissions_ReaderError",
		"list-permissions",
		"--account-address", "0x1234567890abcdef1234567890abcdef12345678",
		"--appointee-address", "0x9876543210fedcba9876543210fedcba98765432",
		"--network", "holesky",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
	}

	err := app.Run(args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedError)
}

func TestListPermissions_NoPermissions(t *testing.T) {
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		ListPermissionsCmd(generateMockListPermissionsReader([]gethcommon.Address{}, [][4]byte{}, nil)),
	}

	args := []string{
		"TestListPermissions_NoPermissions",
		"list-permissions",
		"--account-address", "0x1234567890abcdef1234567890abcdef12345678",
		"--appointee-address", "0x9876543210fedcba9876543210fedcba98765432",
		"--network", "holesky",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
	}

	err := app.Run(args)
	assert.NoError(t, err)
}
