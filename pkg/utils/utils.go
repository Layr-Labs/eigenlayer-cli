package utils

import (
	"bufio"
	"errors"
	"log"
	"math/big"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

func ChainIdToNetworkName(chainId int64) string {
	switch chainId {
	case MainnetChainId:
		return MainnetNetworkName
	case HoleskyChainId:
		return HoleskyNetworkName
	case AnvilChainId:
		return AnvilNetworkName
	default:
		return UnknownNetworkName
	}
}

func NetworkNameToChainId(networkName string) *big.Int {
	switch networkName {
	case MainnetNetworkName:
		return big.NewInt(MainnetChainId)
	case HoleskyNetworkName:
		return big.NewInt(HoleskyChainId)
	case AnvilNetworkName:
		return big.NewInt(AnvilChainId)
	default:
		return big.NewInt(-1)
	}
}

func ReadYamlConfig(path string, o interface{}) error {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		log.Fatal("Path ", path, " does not exist")
	}
	b, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return err
	}

	return yaml.Unmarshal(b, o)
}

func GetStdInPassword() (string, bool) {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Input is available in the pipe, read from it
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			return scanner.Text(), true
		}
	}
	return "", false
}
