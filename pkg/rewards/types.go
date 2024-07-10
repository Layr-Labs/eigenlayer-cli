package rewards

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
