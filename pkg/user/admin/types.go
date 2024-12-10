package admin

import (
	gethcommon "github.com/ethereum/go-ethereum/common"
	"math/big"
)

type listPendingAdminsConfig struct {
	Network                  string
	RPCUrl                   string
	AccountAddress           gethcommon.Address
	PermissionManagerAddress gethcommon.Address
	ChainID                  *big.Int
	Environment              string
}

type listAdminsConfig struct {
	Network                  string
	RPCUrl                   string
	AccountAddress           gethcommon.Address
	PermissionManagerAddress gethcommon.Address
	ChainID                  *big.Int
	Environment              string
}
