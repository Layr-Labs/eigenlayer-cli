package appointee

import (
	"math/big"

	gethcommon "github.com/ethereum/go-ethereum/common"
)

type canCallConfig struct {
	Network                  string
	RPCUrl                   string
	UserAddress              gethcommon.Address
	CallerAddress            gethcommon.Address
	Target                   gethcommon.Address
	Selector                 [4]byte
	PermissionManagerAddress gethcommon.Address
	ChainID                  *big.Int
	Environment              string
}
