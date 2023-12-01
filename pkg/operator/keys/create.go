package keys

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/Layr-Labs/eigensdk-go/crypto/bls"
	sdkEcdsa "github.com/Layr-Labs/eigensdk-go/crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/urfave/cli/v2"
	passwordvalidator "github.com/wagslane/go-password-validator"
)

const (
	OperatorKeystoreSubFolder = ".eigenlayer/operator_keys"

	KeyTypeECDSA = "ecdsa"
	KeyTypeBLS   = "bls"

	// MinEntropyBits For password validation
	MinEntropyBits = 70
)

func CreateCmd(p utils.Prompter) *cli.Command {

	createCmd := &cli.Command{
		Name:      "create",
		Usage:     "Used to create encrypted keys in local keystore",
		UsageText: "create --key-type <key-type> [flags] <keyname>",
		Description: `
Used to create ecdsa and bls key in local keystore

keyname (required) - This will be the name of the created key file. It will be saved as <keyname>.ecdsa.key.json or <keyname>.bls.key.json

use --key-type ecdsa/bls to create ecdsa/bls key. 
It will prompt for password to encrypt the key, which is optional but highly recommended.
If you want to create a key with weak/no password, use --insecure flag. Do NOT use those keys in production

This command will create keys in $HOME/.eigenlayer/operator_keys/ location
		`,
		Flags: []cli.Flag{
			&KeyTypeFlag,
			&InsecureFlag,
		},

		Action: func(ctx *cli.Context) error {
			args := ctx.Args()
			if args.Len() != 1 {
				return fmt.Errorf("%w: accepts 1 arg, received %d", ErrInvalidNumberOfArgs, args.Len())
			}

			keyName := args.Get(0)
			if err := validateKeyName(keyName); err != nil {
				return err
			}

			keyType := ctx.String(KeyTypeFlag.Name)
			insecure := ctx.Bool(InsecureFlag.Name)

			switch keyType {
			case KeyTypeECDSA:
				privateKey, err := crypto.GenerateKey()
				if err != nil {
					return err
				}
				return saveEcdsaKey(keyName, p, privateKey, insecure)
			case KeyTypeBLS:
				blsKeyPair, err := bls.GenRandomBlsKeys()
				if err != nil {
					return err
				}
				return saveBlsKey(keyName, p, blsKeyPair, insecure)
			default:
				return ErrInvalidKeyType
			}
		},
	}
	return createCmd
}

func validateKeyName(keyName string) error {
	if len(keyName) == 0 {
		return ErrEmptyKeyName
	}

	if match, _ := regexp.MatchString("\\s", keyName); match {
		return ErrKeyContainsWhitespaces
	}

	return nil
}

func saveBlsKey(keyName string, p utils.Prompter, keyPair *bls.KeyPair, insecure bool) error {
	homePath, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	keyFileName := keyName + ".bls.key.json"
	fileLoc := filepath.Clean(filepath.Join(homePath, OperatorKeystoreSubFolder, keyFileName))
	if checkIfKeyExists(fileLoc) {
		return errors.New("key name already exists. Please choose a different name")
	}
	password, err := p.InputHiddenString("Enter password to encrypt the bls private key:", "",
		func(s string) error {
			if insecure {
				return nil
			}
			return validatePassword(s)
		},
	)
	if err != nil {
		return err
	}

	err = keyPair.SaveToFile(fileLoc, password)
	if err != nil {
		return err
	}
	// TODO: display it using `less` of `vi` so that it is not saved in terminal history
	fmt.Println("BLS Private Key: " + keyPair.PrivKey.String())
	fmt.Println("Please backup the above private key in safe place.")
	fmt.Println()
	fmt.Println("BLS Pub key: " + keyPair.PubKey.String())
	fmt.Println("Key location: " + fileLoc)
	return nil
}

func saveEcdsaKey(keyName string, p utils.Prompter, privateKey *ecdsa.PrivateKey, insecure bool) error {
	homePath, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	keyFileName := keyName + ".ecdsa.key.json"
	fileLoc := filepath.Clean(filepath.Join(homePath, OperatorKeystoreSubFolder, keyFileName))
	if checkIfKeyExists(fileLoc) {
		return errors.New("key name already exists. Please choose a different name")
	}

	password, err := p.InputHiddenString("Enter password to encrypt the ecdsa private key:", "",
		func(s string) error {
			if insecure {
				return nil
			}
			return validatePassword(s)
		},
	)
	if err != nil {
		return err
	}

	err = sdkEcdsa.WriteKey(fileLoc, privateKey, password)
	if err != nil {
		return err
	}

	privateKeyHex := hex.EncodeToString(privateKey.D.Bytes())
	// TODO: display it using `less` of `vi` so that it is not saved in terminal history
	fmt.Println("ECDSA Private Key (Hex): ", privateKeyHex)
	fmt.Println("Please backup the above private key hex in safe place.")
	fmt.Println()
	fmt.Println("Key location: " + fileLoc)
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return err
	}
	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	fmt.Println("Public Key hex: ", hexutil.Encode(publicKeyBytes)[4:])
	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	fmt.Println("Ethereum Address", address)
	return nil
}

func checkIfKeyExists(fileLoc string) bool {
	_, err := os.Stat(fileLoc)
	return !os.IsNotExist(err)
}

func validatePassword(password string) error {
	err := passwordvalidator.Validate(password, MinEntropyBits)
	if err != nil {
		fmt.Println(
			"if you want to create keys for testing with weak/no password, use --insecure flag. Do NOT use those keys in production",
		)
	}
	return err
}
