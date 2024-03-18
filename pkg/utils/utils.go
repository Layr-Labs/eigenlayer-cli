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
		ELAVSDirectoryAddress:      "",
	},
	GoerliChainId: {
		BlockExplorerUrl:           "https://goerli.etherscan.io/tx",
		ELDelegationManagerAddress: "0x1b7b8F6b258f95Cf9596EabB9aa18B62940Eb0a8",
		ELAVSDirectoryAddress:      "0x0AC9694c271eFbA6059e9783769e515E8731f935",
	},
	HoleskyChainId: {
		BlockExplorerUrl:           "https://holesky.etherscan.io/tx",
		ELDelegationManagerAddress: "0xA44151489861Fe9e3055d95adC98FbD462B948e7",
		ELAVSDirectoryAddress:      "0x055733000064333CaDDbC92763c58BF0192fFeBf",
	},
	LocalChainId: {
		BlockExplorerUrl:           "",
		ELDelegationManagerAddress: "",
		ELAVSDirectoryAddress:      "",
	},
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
