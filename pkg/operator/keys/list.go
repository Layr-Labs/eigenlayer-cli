package keys

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
	"strings"
)

func ListCmd() *cli.Command {
	listCmd := &cli.Command{
		Name:      "list",
		Usage:     "List all the keys created by this create command",
		UsageText: "list",
		Description: `
This command will list both ecdsa and bls key created using create command

It will only list keys created in the default folder (./operator_keys/)
		`,
		Action: func(context *cli.Context) error {
			homePath, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			keyStorePath := filepath.Clean(filepath.Join(homePath, OperatorKeystoreSubFolder))
			files, err := os.ReadDir(keyStorePath)
			if err != nil {
				return err
			}

			for _, file := range files {
				keySplits := strings.Split(file.Name(), ".")
				fileName := keySplits[0]
				keyType := keySplits[1]
				fmt.Println("Key Name: " + fileName)
				switch keyType {
				case KeyTypeECDSA:
					fmt.Println("Key Type: ECDSA")
					keyFilePath := filepath.Join(keyStorePath, file.Name())
					address, err := GetAddress(filepath.Clean(keyFilePath))
					if err != nil {
						return err
					}
					fmt.Println("Address: 0x" + address)
					fmt.Println("Key location: " + keyFilePath)
					fmt.Println("====================================================================================")
					fmt.Println()
				case KeyTypeBLS:
					fmt.Println("Key Type: BLS")
					keyFilePath := filepath.Join(keyStorePath, file.Name())
					pubKey, err := GetPubKey(filepath.Clean(keyFilePath))
					if err != nil {
						return err
					}
					fmt.Println("Public Key: " + pubKey)
					fmt.Println("Key location: " + keyFilePath)
					fmt.Println("====================================================================================")
					fmt.Println()
				}

			}
			return nil
		},
	}
	return listCmd
}

func GetPubKey(keyStoreFile string) (string, error) {
	keyJson, err := os.ReadFile(keyStoreFile)
	if err != nil {
		return "", err
	}

	m := make(map[string]interface{})
	if err := json.Unmarshal(keyJson, &m); err != nil {
		return "", err
	}

	if pubKey, ok := m["pubKey"].(string); !ok {
		return "", fmt.Errorf("pubKey not found in key file")
	} else {
		return pubKey, nil
	}
}

func GetAddress(keyStoreFile string) (string, error) {
	keyJson, err := os.ReadFile(keyStoreFile)
	if err != nil {
		return "", err
	}

	m := make(map[string]interface{})
	if err := json.Unmarshal(keyJson, &m); err != nil {
		return "", err
	}

	if address, ok := m["address"].(string); !ok {
		return "", fmt.Errorf("address not found in key file")
	} else {
		return address, nil
	}
}

// GetECDSAPrivateKey - Keeping it right now as we might need this function to export
// the keys
func GetECDSAPrivateKey(keyStoreFile string, password string) (*ecdsa.PrivateKey, error) {
	keyStoreContents, err := os.ReadFile(keyStoreFile)
	if err != nil {
		return nil, err
	}

	sk, err := keystore.DecryptKey(keyStoreContents, password)
	if err != nil {
		return nil, err
	}

	return sk.PrivateKey, nil
}
