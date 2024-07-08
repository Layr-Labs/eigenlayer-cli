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
	LocalChainId   = 31337

	MainnetNetworkName = "mainnet"
	HoleskyNetworkName = "holesky"
	PreprodNetworkName = "preprod"
	LocalNetworkName   = "local"
	UnknownNetworkName = "unknown"
)

var (
	ZeroAddress = common.HexToAddress("")
)
