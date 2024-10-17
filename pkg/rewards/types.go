package rewards

import (
	"math/big"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/types"
	gethcommon "github.com/ethereum/go-ethereum/common"
)

type rewardsJson struct {
	Address string `json:"tokenAddress"`
	Amount  string `json:"amount"`
}

type allRewardsJson []rewardsJson

type ClaimConfig struct {
	Network                   string
	RPCUrl                    string
	EarnerAddress             gethcommon.Address
	RecipientAddress          gethcommon.Address
	ClaimerAddress            gethcommon.Address
	Output                    string
	OutputType                string
	Broadcast                 bool
	TokenAddresses            []gethcommon.Address
	RewardsCoordinatorAddress gethcommon.Address
	ClaimTimestamp            string
	ChainID                   *big.Int
	ProofStoreBaseURL         string
	Environment               string
	SignerConfig              *types.SignerConfig
	IsSilent                  bool
}

type SetClaimerConfig struct {
	ClaimerAddress            gethcommon.Address
	Network                   string
	RPCUrl                    string
	Broadcast                 bool
	RewardsCoordinatorAddress gethcommon.Address
	ChainID                   *big.Int
	SignerConfig              *types.SignerConfig
	EarnerAddress             gethcommon.Address
	Output                    string
	OutputType                string
}

type ShowConfig struct {
	EarnerAddress             gethcommon.Address
	RPCUrl                    string
	NumberOfDays              int64
	Network                   string
	Environment               string
	ClaimType                 ClaimType
	ChainID                   *big.Int
	Output                    string
	OutputType                string
	ProofStoreBaseURL         string
	ClaimTimestamp            string
	RewardsCoordinatorAddress gethcommon.Address
}
