package utils

import "github.com/ethereum/go-ethereum/common"

const (
	EmojiCheckMark = "✅"
	EmojiCrossMark = "❌"
	EmojiWarning   = "⚠️"
	EmojiInfo      = "ℹ️"
	EmojiWait      = "⏳"
	EmojiLink      = "🔗"
	EmojiInternet  = "🌐"

	MainnetChainId = 1
	HoleskyChainId = 17000
	SepoliaChainId = 11155111
	HoodiChainId   = 560048
	AnvilChainId   = 31337

	CallDataOutputType string = "calldata"
	PrettyOutputType   string = "pretty"
	JsonOutputType     string = "json"

	MainnetNetworkName = "mainnet"
	HoleskyNetworkName = "holesky"
	SepoliaNetworkName = "sepolia"
	AnvilNetworkName   = "anvil"
	UnknownNetworkName = "unknown"

	MainnetBlockExplorerUrl = "https://etherscan.io/"
	HoleskyBlockExplorerUrl = "https://holesky.etherscan.io"
	SepoliaBlockExplorerUrl = "https://sepolia.etherscan.io"
	HoodiBlockExplorerUrl   = "https://hoodi.etherscan.io"
)

var (
	ZeroAddress = common.HexToAddress("")
)
