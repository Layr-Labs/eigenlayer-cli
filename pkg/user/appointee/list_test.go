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

type mockListAppointeesReader struct {
	listAppointeesFunc func(
		ctx context.Context,
		accountAddress gethcommon.Address,
		target gethcommon.Address,
		selector [4]byte,
	) ([]gethcommon.Address, error)
}

func newMockListAppointeesReader(appointeeAddresses []gethcommon.Address, err error) *mockListAppointeesReader {
	return &mockListAppointeesReader{
		listAppointeesFunc: func(ctx context.Context, accountAddress, target gethcommon.Address, selector [4]byte) ([]gethcommon.Address, error) {
			return appointeeAddresses, err
		},
	}
}

func (m *mockListAppointeesReader) ListAppointees(
	ctx context.Context,
	accountAddress,
	target gethcommon.Address,
	selector [4]byte,
) ([]gethcommon.Address, error) {
	return m.listAppointeesFunc(ctx, accountAddress, target, selector)
}

func generateMockListReader(
	appointeeAddresses []gethcommon.Address,
	err error,
) func(logging.Logger, *listAppointeesConfig) (ListAppointeesReader, error) {
	return func(logger logging.Logger, config *listAppointeesConfig) (ListAppointeesReader, error) {
		return newMockListAppointeesReader(appointeeAddresses, err), nil
	}
}

func TestListAppointees_Success(t *testing.T) {
	expectedAppointees := []gethcommon.Address{
		gethcommon.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"),
		gethcommon.HexToAddress("0x9876543210fedcba9876543210fedcba98765432"),
	}

	app := cli.NewApp()
	app.Commands = []*cli.Command{
		ListCmd(generateMockListReader(expectedAppointees, nil)),
	}

	args := []string{
		"TestListAppointees_Success",
		"list",
		"--account-address", "0x1234567890abcdef1234567890abcdef12345678",
		"--target-address", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--selector", "0x1A2B3C4D",
		"--network", "holesky",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
	}

	err := app.Run(args)
	assert.NoError(t, err)
}

func TestListAppointees_ReaderError(t *testing.T) {
	expectedError := "Error fetching appointees"
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		ListCmd(generateMockListReader(nil, errors.New(expectedError))),
	}

	args := []string{
		"TestListAppointees_ReaderError",
		"list",
		"--account-address", "0x1234567890abcdef1234567890abcdef12345678",
		"--target-address", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--selector", "0x1A2B3C4D",
		"--network", "holesky",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
	}

	err := app.Run(args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedError)
}

func TestListAppointees_InvalidSelector(t *testing.T) {
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		ListCmd(generateMockListReader(nil, nil)),
	}

	args := []string{
		"TestListAppointees_InvalidSelector",
		"list",
		"--account-address", "0x1234567890abcdef1234567890abcdef12345678",
		"--target-address", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--selector", "invalid",
		"--network", "holesky",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
	}

	err := app.Run(args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "selector must be a 4-byte hex string prefixed with '0x'")
}

func TestListAppointees_NoAppointees(t *testing.T) {
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		ListCmd(generateMockListReader([]gethcommon.Address{}, nil)),
	}

	args := []string{
		"TestListAppointees_NoAppointees",
		"list",
		"--account-address", "0x1234567890abcdef1234567890abcdef12345678",
		"--target-address", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--selector", "0x1A2B3C4D",
		"--network", "holesky",
		"--eth-rpc-url", "https://ethereum-holesky.publicnode.com/",
	}

	err := app.Run(args)
	assert.NoError(t, err)
}
