package utils

import (
	"bufio"
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/types"

	"gopkg.in/yaml.v2"
)

var ChainMetadataMap = map[int64]types.ChainMetadata{
	MainnetChainId: {
		BlockExplorerUrl:           "https://etherscan.io/tx",
		ELDelegationManagerAddress: "0x39053D51B77DC0d36036Fc1fCc8Cb819df8Ef37A",
		ELAVSDirectoryAddress:      "0x135dda560e946695d6f155dacafc6f1f25c1f5af",
		WebAppUrl:                  "https://app.eigenlayer.xyz/operator",
	},
	HoleskyChainId: {
		BlockExplorerUrl:           "https://holesky.etherscan.io/tx",
		ELDelegationManagerAddress: "0xA44151489861Fe9e3055d95adC98FbD462B948e7",
		ELAVSDirectoryAddress:      "0x055733000064333CaDDbC92763c58BF0192fFeBf",
		WebAppUrl:                  "https://holesky.eigenlayer.xyz/operator",
	},
	LocalChainId: {
		BlockExplorerUrl:           "",
		ELDelegationManagerAddress: "",
		ELAVSDirectoryAddress:      "",
		WebAppUrl:                  "",
	},
}

func ChainIdToNetworkName(chainId int64) string {
	switch chainId {
	case MainnetChainId:
		return "mainnet"
	case HoleskyChainId:
		return "holesky"
	case LocalChainId:
		return "local"
	default:
		return "unknown"
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
