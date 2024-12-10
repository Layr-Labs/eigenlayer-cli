package appointee

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/types"
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

type listUsersConfig struct {
	Network                  string
	RPCUrl                   string
	AccountAddress           gethcommon.Address
	UserAddress              gethcommon.Address
	Target                   gethcommon.Address
	Selector                 [4]byte
	PermissionManagerAddress gethcommon.Address
	ChainID                  *big.Int
	Environment              string
}

type listUserPermissionsConfig struct {
	Network                  string
	RPCUrl                   string
	AccountAddress           gethcommon.Address
	UserAddress              gethcommon.Address
	PermissionManagerAddress gethcommon.Address
	ChainID                  *big.Int
	Environment              string
}

type removeConfig struct {
	Network                  string
	RPCUrl                   string
	AccountAddress           gethcommon.Address
	UserAddress              gethcommon.Address
	Target                   gethcommon.Address
	SignerConfig             types.SignerConfig
	Selector                 [4]byte
	PermissionManagerAddress gethcommon.Address
	ChainID                  *big.Int
	Environment              string
}

type setConfig struct {
	Network                  string
	RPCUrl                   string
	AccountAddress           gethcommon.Address
	UserAddress              gethcommon.Address
	Target                   gethcommon.Address
	SignerConfig             types.SignerConfig
	Selector                 [4]byte
	PermissionManagerAddress gethcommon.Address
	ChainID                  *big.Int
	Environment              string
}
