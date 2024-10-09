package keys

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"github.com/Layr-Labs/bn254-keystore-go/keystore"
	"math/big"
	"strings"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

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
			if args.Len() != 2 {
				return fmt.Errorf("%w: accepts 2 arg, received %d", ErrInvalidNumberOfArgs, args.Len())
			}

			keyName := args.Get(0)
			if err := validateKeyName(keyName); err != nil {
				return err
			}

			privateKey := args.Get(1)
			if privateKey == "" {
				return ErrEmptyPrivateKey
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
				var pkBytes []byte
				var keyPair *keystore.KeyPair
				password, err := getPassword(p, insecure, stdInPassword, readFromPipe, "Enter password to encrypt the bls private key:")
				if len(pkSlice) == 1 {
					pkInt, ok := new(big.Int).SetString(privateKey, 10)
					if ok {
						// It's a bigInt
						pkBytes = pkInt.Bytes()
					} else {
						// It's a hex string
						pkHex := common.Trim0x(privateKey)
						pkBytes, err = hex.DecodeString(pkHex)
						if err != nil {
							return err
						}
					}
					keyPair = &keystore.KeyPair{
						PrivateKey: pkBytes,
						Password:   password,
					}
				} else {
					keyPair, err = keystore.NewKeyPairFromMnemonic(privateKey, password)
					if err != nil {
						return err
					}
				}

				return saveBlsKeyERC2335(keyName, keyPair)
			default:
				return ErrInvalidKeyType
			}
		},
	}
	return importCmd
}
