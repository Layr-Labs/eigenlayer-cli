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
		ELRewardsCoordinatorAddress: "0x7750d328b314EfFa365A0402CcfD489B80B0adda",
		WebAppUrl:                   "https://app.eigenlayer.xyz/operator",
		ProofStoreBaseURL:           "https://eigenlabs-rewards-mainnet-ethereum.s3.amazonaws.com",
	},
	HoleskyChainId: {
		BlockExplorerUrl:            "https://holesky.etherscan.io/tx",
		ELDelegationManagerAddress:  "0xA44151489861Fe9e3055d95adC98FbD462B948e7",
		ELAVSDirectoryAddress:       "0x055733000064333CaDDbC92763c58BF0192fFeBf",
		ELRewardsCoordinatorAddress: "0xAcc1fb458a1317E886dB376Fc8141540537E68fE",
		WebAppUrl:                   "https://holesky.eigenlayer.xyz/operator",
		ProofStoreBaseURL:           "https://eigenlabs-rewards-testnet-holesky.s3.amazonaws.com",
	},
	AnvilChainId: {
		BlockExplorerUrl:            "",
		ELDelegationManagerAddress:  "0xDc64a140Aa3E981100a9becA4E685f962f0cF6C9",
		ELAVSDirectoryAddress:       "0x0165878A594ca255338adfa4d48449f69242Eb8F",
		ELRewardsCoordinatorAddress: "0x610178dA211FEF7D417bC0e6FeD39F05609AD788",
		WebAppUrl:                   "",
		ProofStoreBaseURL:           "",
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

func GetProofStoreBaseURL(network string) string {
	chainId := NetworkNameToChainId(network)
	chainMetadata, ok := ChainMetadataMap[chainId.Int64()]
	if !ok {
		return ""
	} else {
		return chainMetadata.ProofStoreBaseURL
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
