package keys

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Layr-Labs/eigensdk-go/crypto/bls"

	prompterMock "github.com/Layr-Labs/eigenlayer-cli/pkg/utils/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"go.uber.org/mock/gomock"
)

func TestImportCmd(t *testing.T) {
	homePath, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name            string
		args            []string
		err             error
		keyPath         string
		expectedPrivKey string
		promptMock      func(p *prompterMock.MockPrompter)
	}{
		{
			name: "key-name flag not set",
			args: []string{},
			err:  errors.New("Required flag \"key-type\" not set"),
		},
		{
			name: "one argument",
			args: []string{"--key-type", "ecdsa", "arg1"},
			err:  fmt.Errorf("%w: accepts 2 arg, received 1", ErrInvalidNumberOfArgs),
		},

		{
			name: "more than two argument",
			args: []string{"--key-type", "ecdsa", "arg1", "arg2", "arg3"},
			err:  fmt.Errorf("%w: accepts 2 arg, received 3", ErrInvalidNumberOfArgs),
		},
		{
			name: "empty key name argument",
			args: []string{"--key-type", "ecdsa", "", ""},
			err:  ErrEmptyKeyName,
		},
		{
			name: "keyname with whitespaces",
			args: []string{"--key-type", "ecdsa", "hello world", ""},
			err:  ErrKeyContainsWhitespaces,
		},
		{
			name: "empty private key argument",
			args: []string{"--key-type", "ecdsa", "hello", ""},
			err:  ErrEmptyPrivateKey,
		},
		{
			name: "keyname with whitespaces",
			args: []string{"--key-type", "ecdsa", "hello", "hello world"},
			err:  ErrInvalidKeyFormat,
		},
		{
			name: "invalid key type",
			args: []string{"--key-type", "invalid", "hello", "privkey"},
			err:  ErrInvalidKeyType,
		},
		{
			name: "invalid password based on validation function - ecdsa",
			args: []string{
				"--key-type",
				"ecdsa",
				"test",
				"6842fb8f5fa574d0482818b8a825a15c4d68f542693197f2c2497e3562f335f6",
			},
			err: ErrInvalidPassword,
			promptMock: func(p *prompterMock.MockPrompter) {
				p.EXPECT().InputHiddenString(gomock.Any(), gomock.Any(), gomock.Any()).Return("", ErrInvalidPassword)
			},
		},
		{
			name: "invalid password based on validation function - bls",
			args: []string{"--key-type", "bls", "test", "123"},
			err:  ErrInvalidPassword,
			promptMock: func(p *prompterMock.MockPrompter) {
				p.EXPECT().InputHiddenString(gomock.Any(), gomock.Any(), gomock.Any()).Return("", ErrInvalidPassword)
			},
		},
		{
			name: "valid ecdsa key import",
			args: []string{
				"--key-type",
				"ecdsa",
				"test",
				"6842fb8f5fa574d0482818b8a825a15c4d68f542693197f2c2497e3562f335f6",
			},
			err: nil,
			promptMock: func(p *prompterMock.MockPrompter) {
				p.EXPECT().InputHiddenString(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)
				p.EXPECT().InputHiddenString(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)
			},
			expectedPrivKey: "6842fb8f5fa574d0482818b8a825a15c4d68f542693197f2c2497e3562f335f6",
			keyPath:         filepath.Join(homePath, OperatorKeystoreSubFolder, "/test.ecdsa.key.json"),
		},
		{
			name: "valid ecdsa key import with 0x prefix",
			args: []string{
				"--key-type",
				"ecdsa",
				"test",
				"0x6842fb8f5fa574d0482818b8a825a15c4d68f542693197f2c2497e3562f335f6",
			},
			err: nil,
			promptMock: func(p *prompterMock.MockPrompter) {
				p.EXPECT().InputHiddenString(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)
				p.EXPECT().InputHiddenString(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)
			},
			expectedPrivKey: "6842fb8f5fa574d0482818b8a825a15c4d68f542693197f2c2497e3562f335f6",
			keyPath:         filepath.Join(homePath, OperatorKeystoreSubFolder, "/test.ecdsa.key.json"),
		},
		{
			name: "valid ecdsa key import with mnemonic",
			args: []string{
				"--key-type",
				"ecdsa",
				"test",
				"kidney various problem toe ready mass exhibit volume shuffle must glue sketch",
			},
			err: nil,
			promptMock: func(p *prompterMock.MockPrompter) {
				p.EXPECT().InputHiddenString(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)
				p.EXPECT().InputHiddenString(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)
			},
			expectedPrivKey: "aee7f88721a86c9e269f50ba9a8675609ee8eef54947827fcdce818d8aafd3b1",
			keyPath:         filepath.Join(homePath, OperatorKeystoreSubFolder, "/test.ecdsa.key.json"),
		},
		{
			name: "valid bls key import",
			args: []string{
				"--key-type",
				"bls",
				"test",
				"20030410000080487431431153104351076122223465926814327806350179952713280726583",
			},
			err: nil,
			promptMock: func(p *prompterMock.MockPrompter) {
				p.EXPECT().InputHiddenString(gomock.Any(), gomock.Any(), gomock.Any()).Return("p@$$w0rd", nil)
				p.EXPECT().InputHiddenString(gomock.Any(), gomock.Any(), gomock.Any()).Return("p@$$w0rd", nil)
			},
			expectedPrivKey: "20030410000080487431431153104351076122223465926814327806350179952713280726583",
			keyPath:         filepath.Join(homePath, OperatorKeystoreSubFolder, "/test.bls.key.json"),
		},
		{
			name: "valid bls key import for hex key",
			args: []string{
				"--key-type",
				"bls",
				"test",
				"0xfe198b992d97545b3b0174f026f781039f167c13f6d0ce9f511d0d2e973b7f02",
			},
			err: nil,
			promptMock: func(p *prompterMock.MockPrompter) {
				p.EXPECT().InputHiddenString(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)
				p.EXPECT().InputHiddenString(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)
			},
			expectedPrivKey: "5491383829988096583828972342810831790467090979842721151380259607665538989821",
			keyPath:         filepath.Join(homePath, OperatorKeystoreSubFolder, "/test.bls.key.json"),
		},
		{
			name:    "invalid bls key import for hex key",
			args:    []string{"--key-type", "bls", "test", "0xfes"},
			err:     ErrInvalidHexPrivateKey,
			keyPath: filepath.Join(homePath, OperatorKeystoreSubFolder, "/test.bls.key.json"),
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

			importCmd := ImportCmd(p)
			app := cli.NewApp()

			// We do this because the in the parsing of arguments it ignores the first argument
			// for commands, so we add a blank string as the first argument
			// I suspect it does this because it is expecting the first argument to be the name of the command
			// But when we are testing the command, we don't want to have to specify the name of the command
			// since we are creating the command ourselves
			// https://github.com/urfave/cli/blob/c023d9bc5a3122830c9355a0a8c17137e0c8556f/command.go#L323
			args := append([]string{""}, tt.args...)
			cCtx := cli.NewContext(app, nil, &cli.Context{Context: context.Background()})
			err := importCmd.Run(cCtx, args...)

			if tt.err == nil {
				assert.NoError(t, err)
				_, err := os.Stat(tt.keyPath)

				// Check if the error indicates that the file does not exist
				if os.IsNotExist(err) {
					assert.Failf(t, "file does not exist", "file %s does not exist", tt.keyPath)
				}

				if tt.args[1] == KeyTypeECDSA {
					key, err := GetECDSAPrivateKey(tt.keyPath, "")
					assert.NoError(t, err)
					assert.Equal(t, tt.expectedPrivKey, hex.EncodeToString(key.D.Bytes()))
				} else if tt.args[1] == KeyTypeBLS {
					key, err := bls.ReadPrivateKeyFromFile(tt.keyPath, "")
					assert.NoError(t, err)
					assert.Equal(t, tt.expectedPrivKey, key.PrivKey.String())
				}
			} else {
				assert.EqualError(t, err, tt.err.Error())
			}
		})
	}
}
