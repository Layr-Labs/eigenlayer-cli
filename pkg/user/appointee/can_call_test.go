package appointee

import (
	"context"
	"errors"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"testing"
)

type mockElChainReader struct {
	canCallFunc func(
		ctx context.Context,
		userAddress gethcommon.Address,
		callerAddress gethcommon.Address,
		target gethcommon.Address,
		selector [4]byte,
	) (bool, error)
}

func newMockElChainReader() mockElChainReader {
	return mockElChainReader{
		canCallFunc: func(ctx context.Context, userAddress, callerAddress, target gethcommon.Address, selector [4]byte) (bool, error) {
			return true, nil
		},
	}
}

func newErrorMockElChainReader(expectedError string) mockElChainReader {
	return mockElChainReader{
		canCallFunc: func(ctx context.Context, userAddress, callerAddress, target gethcommon.Address, selector [4]byte) (bool, error) {
			return false, errors.New(expectedError)
		},
	}
}

func (m *mockElChainReader) UserCanCall(
	ctx context.Context,
	userAddress, callerAddress,
	target gethcommon.Address,
	selector [4]byte,
) (bool, error) {
	return m.canCallFunc(ctx, userAddress, callerAddress, target, selector)
}

func TestCanCallCmd_Success(t *testing.T) {
	mockReader := newMockElChainReader()

	app := cli.NewApp()
	app.Commands = []*cli.Command{CanCallCmd()}
	app.Metadata = make(map[string]interface{})
	app.Metadata["elReader"] = UserCanCallReader(&mockReader)

	args := []string{
		"TestCanCallCmd_Success",
		"can-call",
		"--user-address", "0x1234567890abcdef1234567890abcdef12345678",
		"--caller-address", "0x9876543210fedcba9876543210fedcba98765432",
		"--target", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--selector", "test",
		"--network", "holesky",
	}

	err := app.Run(args)
	assert.NoError(t, err)
}

func TestCanCallCmd_UserCanCallError(t *testing.T) {
	errString := "Error while executing call from reader"
	mockReader := newErrorMockElChainReader(errString)

	app := cli.NewApp()
	app.Commands = []*cli.Command{CanCallCmd()}
	app.Metadata = make(map[string]interface{})
	app.Metadata["elReader"] = UserCanCallReader(&mockReader)

	args := []string{
		"TestCanCallCmd_UserCanCallError",
		"can-call",
		"--user-address", "0x1234567890abcdef1234567890abcdef12345678",
		"--caller-address", "0x9876543210fedcba9876543210fedcba98765432",
		"--target", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--selector", "test",
	}

	err := app.Run(args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), errString)
}

func TestCanCallCmd_InvalidSelector(t *testing.T) {
	mockReader := newMockElChainReader()

	app := cli.NewApp()
	app.Commands = []*cli.Command{CanCallCmd()}
	app.Metadata = make(map[string]interface{})
	app.Metadata["elReader"] = UserCanCallReader(&mockReader)

	args := []string{
		"TestCanCallCmd_InvalidSelector",
		"can-call",
		"--user-address", "0x1234567890abcdef1234567890abcdef12345678",
		"--caller-address", "0x9876543210fedcba9876543210fedcba98765432",
		"--target", "0xabcdef1234567890abcdef1234567890abcdef12",
		"--selector", "too-long",
	}

	err := app.Run(args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read and validate user can call config: selector must be 4 characters long")
}
