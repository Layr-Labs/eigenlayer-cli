package utils

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/types"

	"gopkg.in/yaml.v2"
)

var ChainMetadataMap = map[int64]types.ChainMetadata{
	MainnetChainId: {
		BlockExplorerUrl:            "https://etherscan.io/tx",
		ELDelegationManagerAddress:  "0x39053D51B77DC0d36036Fc1fCc8Cb819df8Ef37A",
		ELAVSDirectoryAddress:       "0x135dda560e946695d6f155dacafc6f1f25c1f5af",
		ELRewardsCoordinatorAddress: "",
		WebAppUrl:                   "https://app.eigenlayer.xyz/operator",
	},
	HoleskyChainId: {
		BlockExplorerUrl:            "https://holesky.etherscan.io/tx",
		ELDelegationManagerAddress:  "0xA44151489861Fe9e3055d95adC98FbD462B948e7",
		ELAVSDirectoryAddress:       "0x055733000064333CaDDbC92763c58BF0192fFeBf",
		ELRewardsCoordinatorAddress: "0xb22Ef643e1E067c994019A4C19e403253C05c2B0",
		WebAppUrl:                   "https://holesky.eigenlayer.xyz/operator",
	},
	LocalChainId: {
		BlockExplorerUrl:            "",
		ELDelegationManagerAddress:  "",
		ELAVSDirectoryAddress:       "",
		ELRewardsCoordinatorAddress: "",
		WebAppUrl:                   "",
	},
}

func GetRewardCoordinatorAddress(chainID *big.Int) (string, error) {
	chainIDInt := chainID.Int64()
	chainMetadata, ok := ChainMetadataMap[chainIDInt]
	if !ok {
		return "", fmt.Errorf("chain ID %d not supported", chainIDInt)
	} else {
		return chainMetadata.ELRewardsCoordinatorAddress, nil
	}
}

func ChainIdToNetworkName(chainId int64) string {
	switch chainId {
	case MainnetChainId:
		return MainnetNetworkName
	case HoleskyChainId:
		return HoleskyNetworkName
	case LocalChainId:
		return LocalNetworkName
	default:
		return UnknownNetworkName
	}
}

func NetworkNameToChainId(networkName string) *big.Int {
	switch networkName {
	case MainnetNetworkName:
		return big.NewInt(MainnetChainId)
	case HoleskyNetworkName, PreprodNetworkName:
		return big.NewInt(HoleskyChainId)
	case LocalNetworkName:
		return big.NewInt(LocalChainId)
	default:
		return big.NewInt(-1)
	}
}

func GetProofStoreBaseURL(network string) string {
	chainId := NetworkNameToChainId(network)
	chainMetadata, ok := ChainMetadataMap[chainId.Int64()]
	if !ok {
		return ""
	} else {
		return chainMetadata.ProofStoreBaseURL
	}
}

func GetEnvironmentFromNetwork(network string) string {
	switch network {
	case MainnetNetworkName:
		return "prod"
	case HoleskyNetworkName:
		return "testnet"
	default:
		return "preprod"
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
