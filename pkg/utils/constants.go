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
	AnvilChainId   = 31337

	MainnetNetworkName = "mainnet"
	HoleskyNetworkName = "holesky"
	AnvilNetworkName   = "anvil"
	UnknownNetworkName = "unknown"
)

var (
	ZeroAddress = common.HexToAddress("")
)
