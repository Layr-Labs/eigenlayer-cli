package rewards

import (
	"math/big"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/types"
	gethcommon "github.com/ethereum/go-ethereum/common"
)

type RewardResponse struct {
	Rewards []Reward `json:"rewards"`
}

type Reward struct {
	StrategyAddress    string              `json:"strategyAddress"`
	RewardsPerStrategy []RewardPerStrategy `json:"rewards"`
}

type RewardPerStrategy struct {
	AVSAddress string  `json:"avsAddress"`
	Tokens     []Token `json:"tokens"`
}

type Token struct {
	TokenAddress string `json:"tokenAddress"`
	WeiAmount    string `json:"weiAmount"`
}

type NormalizedReward struct {
	StrategyAddress string   `csv:"strategyAddress"`
	AVSAddress      string   `csv:"avsAddress"`
	TokenAddress    string   `csv:"tokenAddress"`
	WeiAmount       *big.Int `csv:"weiAmount"`
}

type UnclaimedRewardResponse struct {
	BlockHeight string              `json:"blockHeight"`
	Rewards     []RewardPerStrategy `json:"rewards"`
}

type NormalizedUnclaimedReward struct {
	TokenAddress string   `csv:"tokenAddress"`
	WeiAmount    *big.Int `csv:"weiAmount"`
}

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
	EarnerAddress gethcommon.Address
	NumberOfDays  int64
	Network       string
	Environment   string
	ClaimType     ClaimType
	ChainID       *big.Int
	Output        string
}
