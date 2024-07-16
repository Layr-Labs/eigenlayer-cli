package rewards

import "math/big"

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
