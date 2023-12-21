package keys

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/Layr-Labs/eigensdk-go/crypto/bls"
	"github.com/Layr-Labs/eigensdk-go/crypto/ecdsa"
	"github.com/urfave/cli/v2"
)

func ExportCmd(p utils.Prompter) *cli.Command {
	exportCmd := &cli.Command{
		Name:      "export",
		Usage:     "Used to export existing keys from local keystore",
		UsageText: "export --key-type <key-type> [flags] [keyname]",
		Description: `Used to export ecdsa and bls key from local keystore

keyname - This will be the name of the key to be imported. If the path of keys is
different from default path created by "create"/"import" command, then provide the
full path using --key-path flag.

If both keyname is provided and --key-path flag is provided, then keyname will be used. 

use --key-type ecdsa/bls to export ecdsa/bls key. 
- ecdsa - exported key should be plaintext hex encoded private key
- bls - exported key should be plaintext bls private key

It will prompt for password to encrypt the key.

This command will import keys from $HOME/.eigenlayer/operator_keys/ location

But if you want it to export from a different location, use --key-path flag`,

		Flags: []cli.Flag{
			&KeyTypeFlag,
			&KeyPathFlag,
		},
		Action: func(c *cli.Context) error {
			keyType := c.String(KeyTypeFlag.Name)

			keyName := c.Args().Get(0)
			if err := validateKeyName(keyName); err != nil {
				return err
			}

			keyPath := c.String(KeyPathFlag.Name)
			if len(keyPath) == 0 && len(keyName) == 0 {
				return errors.New("one of keyname or --key-path is required")
			}

			if len(keyPath) > 0 && len(keyName) > 0 {
				return errors.New("keyname and --key-path both are provided. Please provide only one")
			}

			filePath, err := getKeyPath(keyPath, keyName, keyType)

			confirm, err := p.Confirm("This will show your private key. Are you sure you want to export?")
			if err != nil {
				return err
			}
			if !confirm {
				return nil
			}

			password, err := p.InputHiddenString("Enter password to decrypt the key", "", func(s string) error {
				return nil
			})
			if err != nil {
				return err
			}
			fmt.Println("exporting key from: ", filePath)

			privateKey, err := getPrivateKey(keyType, filePath, password)
			if err != nil {
				return err
			}
			fmt.Println("Private key: ", privateKey)
			return nil
		},
	}

	return exportCmd
}

func getPrivateKey(keyType string, filePath string, password string) (string, error) {
	switch keyType {
	case KeyTypeECDSA:
		key, err := ecdsa.ReadKey(filePath, password)
		if err != nil {
			return "", err
		}
		return hex.EncodeToString(key.D.Bytes()), nil
	case KeyTypeBLS:
		key, err := bls.ReadPrivateKeyFromFile(filePath, password)
		if err != nil {
			return "", err
		}
		return key.PrivKey.String(), nil
	default:
		return "", ErrInvalidKeyType
	}
}

func getKeyPath(keyPath string, keyName string, keyType string) (string, error) {
	homePath, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	var filePath string
	if len(keyName) > 0 {
		switch keyType {
		case KeyTypeECDSA:
			filePath = filepath.Join(homePath, OperatorKeystoreSubFolder, keyName+".ecdsa.key.json")
		case KeyTypeBLS:
			filePath = filepath.Join(homePath, OperatorKeystoreSubFolder, keyName+".bls.key.json")
		default:
			return "", ErrInvalidKeyType
		}

	} else {
		filePath = filepath.Clean(keyPath)
	}

	return filePath, nil
}
