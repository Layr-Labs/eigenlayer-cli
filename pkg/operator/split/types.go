package split

import (
	"math/big"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/types"
	gethcommon "github.com/ethereum/go-ethereum/common"
)

type SetOperatorAVSSplitConfig struct {
	Network                   string
	RPCUrl                    string
	RewardsCoordinatorAddress gethcommon.Address
	ChainID                   *big.Int
	SignerConfig              *types.SignerConfig
	Broadcast                 bool
	OperatorAddress           gethcommon.Address
	AVSAddress                gethcommon.Address
	Split                     uint16
	OutputType                string
	OutputFile                string
	IsSilent                  bool
	OperatorSetId             int
}

type GetOperatorAVSSplitConfig struct {
	Network                   string
	RPCUrl                    string
	RewardsCoordinatorAddress gethcommon.Address
	ChainID                   *big.Int

	OperatorAddress gethcommon.Address
	AVSAddress      gethcommon.Address
	OperatorSetId   int
}
