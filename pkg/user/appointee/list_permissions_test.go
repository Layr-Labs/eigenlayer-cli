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
		accountAddress gethcommon.Address,
		appointeeAddress gethcommon.Address,
	) ([]gethcommon.Address, [][4]byte, error)
}

func newMockListPermissionsReader(
	appointeeAddresses []gethcommon.Address,
	permissions [][4]byte,
	err error,
) *mockListPermissionsReader {
	return &mockListPermissionsReader{
		listPermissionsFunc: func(ctx context.Context, accountAddress, appointeeAddress gethcommon.Address) ([]gethcommon.Address, [][4]byte, error) {
			return appointeeAddresses, permissions, err
		},
	}
}

func (m *mockListPermissionsReader) ListAppointeePermissions(
	ctx context.Context,
	accountAddress gethcommon.Address,
	appointeeAddress gethcommon.Address,
) ([]gethcommon.Address, [][4]byte, error) {
	return m.listPermissionsFunc(ctx, accountAddress, appointeeAddress)
}

func generateMockListPermissionsReader(
	appointeeAddresses []gethcommon.Address,
	permissions [][4]byte,
	err error,
) func(logging.Logger, *listAppointeePermissionsConfig) (PermissionsReader, error) {
	return func(logger logging.Logger, config *listAppointeePermissionsConfig) (PermissionsReader, error) {
		return newMockListPermissionsReader(appointeeAddresses, permissions, err), nil
	}
}

func TestListPermissions_Success(t *testing.T) {
	expectedAppointees := []gethcommon.Address{
		gethcommon.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"),
	}
	expectedPermissions := [][4]byte{
		{0x1A, 0x2B, 0x3C, 0x4D},
	}

	app := cli.NewApp()
	app.Commands = []*cli.Command{
		ListPermissionsCmd(generateMockListPermissionsReader(expectedAppointees, expectedPermissions, nil)),
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
