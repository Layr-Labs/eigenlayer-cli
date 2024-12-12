package admin

import (
	"math/big"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/types"
	gethcommon "github.com/ethereum/go-ethereum/common"
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

type acceptAdminConfig struct {
	Network                  string
	RPCUrl                   string
	AccountAddress           gethcommon.Address
	PermissionManagerAddress gethcommon.Address
	SignerConfig             types.SignerConfig
	ChainID                  *big.Int
	Environment              string
}

type addPendingAdminConfig struct {
	Network                  string
	RPCUrl                   string
	AccountAddress           gethcommon.Address
	AdminAddress             gethcommon.Address
	PermissionManagerAddress gethcommon.Address
	SignerConfig             types.SignerConfig
	ChainID                  *big.Int
	Environment              string
}

type removeAdminConfig struct {
	Network                  string
	RPCUrl                   string
	AccountAddress           gethcommon.Address
	AdminAddress             gethcommon.Address
	PermissionManagerAddress gethcommon.Address
	SignerConfig             types.SignerConfig
	ChainID                  *big.Int
	Environment              string
}

type removePendingAdminConfig struct {
	Network                  string
	RPCUrl                   string
	AccountAddress           gethcommon.Address
	AdminAddress             gethcommon.Address
	PermissionManagerAddress gethcommon.Address
	SignerConfig             types.SignerConfig
	ChainID                  *big.Int
	Environment              string
}
