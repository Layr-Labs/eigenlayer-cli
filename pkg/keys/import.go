package keys

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	"github.com/Layr-Labs/eigensdk-go/crypto/bls"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/urfave/cli/v2"
)

func ImportCmd(p utils.Prompter) *cli.Command {
	importCmd := &cli.Command{
		Name:      "import",
		Usage:     "Used to import existing keys in local keystore",
		UsageText: "import --key-type <key-type> [flags] <keyname> <private-key>",
		Description: `
Used to import ecdsa and bls key in local keystore

keyname (required) - This will be the name of the imported key file. It will be saved as <keyname>.ecdsa.key.json or <keyname>.bls.key.json

use --key-type ecdsa/bls to import ecdsa/bls key. 
- ecdsa - <private-key> should be plaintext hex encoded private key
- bls - <private-key> should be plaintext bls private key

It will prompt for password to encrypt the key, which is optional but highly recommended.
If you want to import a key with weak/no password, use --insecure flag. Do NOT use those keys in production

This command also support piping the password from stdin.
For example: echo "password" | eigenlayer keys import --key-type ecdsa keyname privateKey

This command will import keys in $HOME/.eigenlayer/operator_keys/ location
		`,
		Flags: []cli.Flag{
			&KeyTypeFlag,
			&InsecureFlag,
		},
		After: telemetry.AfterRunAction(),
		Action: func(ctx *cli.Context) error {
			args := ctx.Args()
			if args.Len() != 1 && args.Len() != 2 {
				return fmt.Errorf("%w: accepts 1 or 2 arg, received %d", ErrInvalidNumberOfArgs, args.Len())
			}

			var err error
			keyName := args.Get(0)
			if err = validateKeyName(keyName); err != nil {
				return err
			}

			privateKey := args.Get(1)

			// In case user doesn't provide private key, prompt for it
			// This is to ensure backward compatibility
			if len(privateKey) == 0 {
				privateKey, err = p.InputString(
					"Enter private key or Mnemonic (mnemonic only work for ecdsa right now): ",
					"",
					"Please provide private key or Mnemonic to import key",
					func(s string) error {
						if len(s) == 0 {
							return ErrEmptyPrivateKey
						}
						return nil
					},
				)
				if err != nil {
					return err
				}
			}

			pkSlice := strings.Split(privateKey, " ")
			if len(pkSlice) != 1 && len(pkSlice) != 12 {
				return ErrInvalidKeyFormat
			}

			// Check if input is available in the pipe and read the password from it
			stdInPassword, readFromPipe := utils.GetStdInPassword()

			keyType := ctx.String(KeyTypeFlag.Name)
			insecure := ctx.Bool(InsecureFlag.Name)

			switch keyType {
			case KeyTypeECDSA:
				var privateKeyPair *ecdsa.PrivateKey
				var err error
				if len(pkSlice) == 1 {
					privateKey = common.Trim0x(privateKey)
					privateKeyPair, err = crypto.HexToECDSA(privateKey)
					if err != nil {
						return err
					}
				} else {
					privateKeyPair, _, err = generateEcdsaKeyWithMnemonic(privateKey)
					if err != nil {
						return err
					}
				}
				return saveEcdsaKey(keyName, p, privateKeyPair, insecure, stdInPassword, readFromPipe, "")
			case KeyTypeBLS:
				privateKeyBigInt := new(big.Int)
				_, ok := privateKeyBigInt.SetString(privateKey, 10)
				var blsKeyPair *bls.KeyPair
				var err error
				if ok {
					fmt.Println("Importing from large integer")
					blsKeyPair, err = bls.NewKeyPairFromString(privateKey)
					if err != nil {
						return err
					}
				} else {
					// Try to parse as hex
					fmt.Println("Importing from hex")
					z := new(big.Int)
					privateKey = common.Trim0x(privateKey)
					_, ok := z.SetString(privateKey, 16)
					if !ok {
						return ErrInvalidHexPrivateKey
					}
					blsKeyPair, err = bls.NewKeyPairFromString(z.String())
					if err != nil {
						return err
					}
				}
				return saveBlsKey(keyName, p, blsKeyPair, insecure, stdInPassword, readFromPipe)
			default:
				return ErrInvalidKeyType
			}
		},
	}
	return importCmd
}
