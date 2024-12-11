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

type isPendingAdminConfig struct {
	Network                  string
	RPCUrl                   string
	AccountAddress           gethcommon.Address
	PendingAdminAddress      gethcommon.Address
	PermissionManagerAddress gethcommon.Address
	ChainID                  *big.Int
	Environment              string
}

type isAdminConfig struct {
	Network                  string
	RPCUrl                   string
	AccountAddress           gethcommon.Address
	AdminAddress             gethcommon.Address
	PermissionManagerAddress gethcommon.Address
	ChainID                  *big.Int
	Environment              string
}
