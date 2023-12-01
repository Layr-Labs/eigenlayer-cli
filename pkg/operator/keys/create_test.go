package keys

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/urfave/cli/v2"

	prompterMock "github.com/Layr-Labs/eigenlayer-cli/pkg/utils/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCreateCmd(t *testing.T) {
	homePath, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		args       []string
		err        error
		keyPath    string
		promptMock func(p *prompterMock.MockPrompter)
	}{
		{
			name: "key-name flag not set",
			args: []string{},
			err:  errors.New("Required flag \"key-type\" not set"),
		},
		{
			name: "more than one argument",
			args: []string{"--key-type", "ecdsa", "arg1", "arg2"},
			err:  fmt.Errorf("%w: accepts 1 arg, received 2", ErrInvalidNumberOfArgs),
		},
		{
			name: "empty name argument",
			args: []string{"--key-type", "ecdsa", ""},
			err:  ErrEmptyKeyName,
		},
		{
			name: "keyname with whitespaces",
			args: []string{"--key-type", "ecdsa", "hello world"},
			err:  ErrKeyContainsWhitespaces,
		},
		{
			name: "invalid key type",
			args: []string{"--key-type", "invalid", "do_not_use_this_name"},
			err:  ErrInvalidKeyType,
		},
		{
			name: "invalid password based on validation function - ecdsa",
			args: []string{"--key-type", "ecdsa", "do_not_use_this_name"},
			err:  ErrInvalidPassword,
			promptMock: func(p *prompterMock.MockPrompter) {
				p.EXPECT().InputHiddenString(gomock.Any(), gomock.Any(), gomock.Any()).Return("", ErrInvalidPassword)
			},
		},
		{
			name: "invalid password based on validation function - bls",
			args: []string{"--key-type", "bls", "do_not_use_this_name"},
			err:  ErrInvalidPassword,
			promptMock: func(p *prompterMock.MockPrompter) {
				p.EXPECT().InputHiddenString(gomock.Any(), gomock.Any(), gomock.Any()).Return("", ErrInvalidPassword)
			},
		},
		{
			name: "valid ecdsa key creation",
			args: []string{"--key-type", "ecdsa", "do_not_use_this_name"},
			err:  nil,
			promptMock: func(p *prompterMock.MockPrompter) {
				p.EXPECT().InputHiddenString(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)
			},
			keyPath: filepath.Join(homePath, OperatorKeystoreSubFolder, "/do_not_use_this_name.ecdsa.key.json"),
		},
		{
			name: "valid bls key creation",
			args: []string{"--key-type", "bls", "do_not_use_this_name"},
			err:  nil,
			promptMock: func(p *prompterMock.MockPrompter) {
				p.EXPECT().InputHiddenString(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)
			},
			keyPath: filepath.Join(homePath, OperatorKeystoreSubFolder, "/do_not_use_this_name.bls.key.json"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				_ = os.Remove(tt.keyPath)
			})
			controller := gomock.NewController(t)
			p := prompterMock.NewMockPrompter(controller)
			if tt.promptMock != nil {
				tt.promptMock(p)
			}

			createCmd := CreateCmd(p)
			app := cli.NewApp()

			// We do this because the in the parsing of arguments it ignores the first argument
			// for commands, so we add a blank string as the first argument
			// I suspect it does this because it is expecting the first argument to be the name of the command
			// But when we are testing the command, we don't want to have to specify the name of the command
			// since we are creating the command ourselves
			// https://github.com/urfave/cli/blob/c023d9bc5a3122830c9355a0a8c17137e0c8556f/command.go#L323
			args := append([]string{""}, tt.args...)

			cCtx := cli.NewContext(app, nil, &cli.Context{Context: context.Background()})
			err := createCmd.Run(cCtx, args...)

			if tt.err == nil {
				assert.NoError(t, err)
				_, err := os.Stat(tt.keyPath)

				// Check if the error indicates that the file does not exist
				if os.IsNotExist(err) {
					assert.Failf(t, "file does not exist", "file %s does not exist", tt.keyPath)
				}
			} else {
				assert.EqualError(t, err, tt.err.Error())
			}
		})
	}
}
