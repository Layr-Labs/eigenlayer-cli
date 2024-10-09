package keys

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Layr-Labs/bn254-keystore-go/curve"
	"github.com/Layr-Labs/bn254-keystore-go/keystore"
	"github.com/Layr-Labs/bn254-keystore-go/mnemonic"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/Layr-Labs/eigensdk-go/crypto/bls"
	sdkEcdsa "github.com/Layr-Labs/eigensdk-go/crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"

	"github.com/tyler-smith/go-bip39"
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

This command also support piping the password from stdin.
For example: echo "password" | eigenlayer keys create --key-type ecdsa keyname

This command will create keys in $HOME/.eigenlayer/operator_keys/ location
		`,
		Flags: []cli.Flag{
			&KeyTypeFlag,
			&InsecureFlag,
		},
		After: telemetry.AfterRunAction(),
		Action: func(ctx *cli.Context) error {
			args := ctx.Args()
			if args.Len() != 1 {
				return fmt.Errorf("%w: accepts 1 arg, received %d", ErrInvalidNumberOfArgs, args.Len())
			}

			keyName := args.Get(0)
			if err := validateKeyName(keyName); err != nil {
				return err
			}

			// Check if input is available in the pipe and read the password from it
			stdInPassword, readFromPipe := utils.GetStdInPassword()

			keyType := ctx.String(KeyTypeFlag.Name)
			insecure := ctx.Bool(InsecureFlag.Name)

			switch keyType {
			case KeyTypeECDSA:
				// Passing empty string to generate a new mnemonic
				privateKey, pkMnemonic, err := generateEcdsaKeyWithMnemonic("")
				if err != nil {
					return err
				}
				return saveEcdsaKey(keyName, p, privateKey, insecure, stdInPassword, readFromPipe, pkMnemonic)
			case KeyTypeBLS:
				password, err := getPassword(
					p,
					insecure,
					stdInPassword,
					readFromPipe,
					"Enter password to encrypt the bls private key:",
				)
				if err != nil {
					return err
				}
				blsKeyPair, err := keystore.NewKeyPair(password, mnemonic.English)
				if err != nil {
					return err
				}
				return saveBlsKeyERC2335(keyName, blsKeyPair)
			default:
				return ErrInvalidKeyType
			}
		},
	}
	return createCmd
}

func generateEcdsaKeyWithMnemonic(mnemonic string) (*ecdsa.PrivateKey, string, error) {
	if mnemonic == "" {
		// Generate entropy for mnemonic
		entropy, err := bip39.NewEntropy(128)
		if err != nil {
			return nil, "", fmt.Errorf("failed to generate entropy: %v", err)
		}
		// Generate mnemonic
		mnemonic, err = bip39.NewMnemonic(entropy)
		if err != nil {
			return nil, "", fmt.Errorf("failed to generate mnemonic: %v", err)
		}
	}

	// Create HD wallet
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create wallet from mnemonic: %v", err)
	}

	// Derive the Ethereum account using the standard derivation path
	path := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/0")
	account, err := wallet.Derive(path, false)
	if err != nil {
		return nil, "", fmt.Errorf("failed to derive account: %v", err)
	}

	// Get private key
	privateKey, err := wallet.PrivateKey(account)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get private key: %v", err)
	}

	return privateKey, mnemonic, nil
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

func saveBlsKeyERC2335(
	keyName string,
	keyPair *keystore.KeyPair,
) error {
	homePath, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	keyFileName := keyName + ".bls.key"
	fileLoc := filepath.Clean(filepath.Join(homePath, OperatorKeystoreSubFolder, keyFileName))
	if checkIfKeyExists(fileLoc + ".json") {
		return errors.New("key name already exists. Please choose a different name")
	}

	encryptedKeystore, err := keyPair.Encrypt(keystore.KDFScrypt, curve.BN254)
	if err != nil {
		return err
	}

	err = encryptedKeystore.SaveWithPubKeyHex(filepath.Join(homePath, OperatorKeystoreSubFolder), keyFileName)
	if err != nil {
		return err
	}

	privateKeyHex := hex.EncodeToString(keyPair.PrivateKey)
	publicKeyHex := encryptedKeystore.PubKey

	fmt.Printf("\nKey location: %s\nPublic Key: %s\n\n", fileLoc+".json", publicKeyHex)
	return displayWithLess(privateKeyHex, KeyTypeBLS, keyPair.Mnemonic)
}

func getPassword(
	p utils.Prompter,
	insecure bool,
	stdInPassword string,
	readFromPipe bool,
	helpMessage string,
) (string, error) {
	var password string
	var err error
	if !readFromPipe {
		password, err = getPasswordFromPrompt(p, insecure, helpMessage)
		if err != nil {
			return "", err
		}
	} else {
		password = stdInPassword
		if !insecure {
			err = validatePassword(password)
			if err != nil {
				return "", err
			}
		}
	}

	return password, nil
}

func saveBlsKey(
	keyName string,
	p utils.Prompter,
	keyPair *bls.KeyPair,
	insecure bool,
	stdInPassword string,
	readFromPipe bool,
) error {
	homePath, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	keyFileName := keyName + ".bls.key.json"
	fileLoc := filepath.Clean(filepath.Join(homePath, OperatorKeystoreSubFolder, keyFileName))
	if checkIfKeyExists(fileLoc) {
		return errors.New("key name already exists. Please choose a different name")
	}

	password, err := getPassword(
		p,
		insecure,
		stdInPassword,
		readFromPipe,
		"Enter password to encrypt the bls private key:",
	)
	if err != nil {
		return err
	}

	err = keyPair.SaveToFile(fileLoc, password)
	if err != nil {
		return err
	}

	privateKeyHex := keyPair.PrivKey.String()
	publicKeyHex := keyPair.PubKey.String()

	fmt.Printf("\nKey location: %s\nPublic Key: %s\n\n", fileLoc, publicKeyHex)
	return displayWithLess(privateKeyHex, KeyTypeBLS, "")
}

func saveEcdsaKey(
	keyName string,
	p utils.Prompter,
	privateKey *ecdsa.PrivateKey,
	insecure bool,
	stdInPassword string,
	readFromPipe bool,
	mnemonic string,
) error {
	homePath, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	keyFileName := keyName + ".ecdsa.key.json"
	fileLoc := filepath.Clean(filepath.Join(homePath, OperatorKeystoreSubFolder, keyFileName))
	if checkIfKeyExists(fileLoc) {
		return errors.New("key name already exists. Please choose a different name")
	}

	password, err := getPassword(
		p,
		insecure,
		stdInPassword,
		readFromPipe,
		"Enter password to encrypt the ecdsa private key:",
	)
	if err != nil {
		return err
	}

	err = sdkEcdsa.WriteKey(fileLoc, privateKey, password)
	if err != nil {
		return err
	}

	privateKeyHex := hex.EncodeToString(privateKey.D.Bytes())

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return errors.New("error casting public key to ECDSA public key")
	}
	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	publicKeyHex := hexutil.Encode(publicKeyBytes)[4:]
	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()

	fmt.Printf("\nKey location: %s\nPublic Key hex: %s\nEthereum Address: %s\n\n", fileLoc, publicKeyHex, address)
	return displayWithLess(privateKeyHex, KeyTypeECDSA, mnemonic)
}

func padLeft(str string, length int) string {
	for len(str) < length {
		str = "0" + str
	}
	return str
}

func displayWithLess(privateKeyHex string, keyType string, mnemonic string) error {
	var message, border, keyLine string
	tabSpace := "    "

	// Pad with 0 to match size of 64 bytes
	if keyType == KeyTypeECDSA {
		privateKeyHex = padLeft(privateKeyHex, 64)
	}
	keyContent := tabSpace + privateKeyHex + tabSpace
	borderLength := len(keyContent) + 4
	border = strings.Repeat("/", borderLength)
	paddingLine := "//" + strings.Repeat(" ", borderLength-4) + "//"

	keyLine = fmt.Sprintf("//%s//", keyContent)

	if keyType == KeyTypeECDSA {
		message = fmt.Sprintf(`
ECDSA Private Key (Hex):

%s
%s
%s
%s
%s

🔐 Please backup the above private key hex in a safe place 🔒

`, border, paddingLine, keyLine, paddingLine, border)
	} else if keyType == KeyTypeBLS {
		message = fmt.Sprintf(`
BLS Private Key (Hex):

%s
%s
%s
%s
%s

🔐 Please backup the above private key hex in a safe place 🔒

`, border, paddingLine, keyLine, paddingLine, border)
	}

	if mnemonic != "" {
		// format mnemonic to be displayed in above format
		mnemonicContent := tabSpace + mnemonic + tabSpace
		borderLength := len(mnemonicContent) + 4
		border = strings.Repeat("/", borderLength)
		paddingLine := "//" + strings.Repeat(" ", borderLength-4) + "//"

		keyLine = fmt.Sprintf("//%s//", mnemonicContent)
		mnemonicDisplay := fmt.Sprintf(`
Mnemonic:

%s
%s
%s
%s
%s

🔐 Please backup the above mnemonic in a safe place 🔒

`, border, paddingLine, keyLine, paddingLine, border)
		message = fmt.Sprintf("%s%s", message, mnemonicDisplay)
	}

	cmd := exec.Command("less", "-R")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("error creating stdin pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting less command: %w", err)
	}

	if _, err := stdin.Write([]byte(message)); err != nil {
		return fmt.Errorf("error writing message to less command: %w", err)
	}

	if err := stdin.Close(); err != nil {
		return fmt.Errorf("error closing stdin pipe: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("error waiting for less command: %w", err)
	}

	return nil
}

func getPasswordFromPrompt(p utils.Prompter, insecure bool, prompt string) (string, error) {
	password, err := p.InputHiddenString(prompt, "",
		func(s string) error {
			if insecure {
				return nil
			}
			return validatePassword(s)
		},
	)
	if err != nil {
		return "", err
	}
	_, err = p.InputHiddenString("Please confirm your password:", "",
		func(s string) error {
			if s != password {
				return errors.New("passwords are not matched")
			}
			return nil
		},
	)
	if err != nil {
		return "", err
	}
	return password, nil
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
