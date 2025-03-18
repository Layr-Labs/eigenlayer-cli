package admin

import (
	"math/big"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/types"
	gethcommon "github.com/ethereum/go-ethereum/common"
)

type listPendingAdminsConfig struct {
	Network                     string
	RPCUrl                      string
	AccountAddress              gethcommon.Address
	PermissionControllerAddress gethcommon.Address
	ChainID                     *big.Int
	Environment                 string
}

type listAdminsConfig struct {
	Network                     string
	RPCUrl                      string
	AccountAddress              gethcommon.Address
	PermissionControllerAddress gethcommon.Address
	ChainID                     *big.Int
	Environment                 string
}

type isPendingAdminConfig struct {
	Network                     string
	RPCUrl                      string
	AccountAddress              gethcommon.Address
	PendingAdminAddress         gethcommon.Address
	PermissionControllerAddress gethcommon.Address
	ChainID                     *big.Int
	Environment                 string
}

type isAdminConfig struct {
	Network                     string
	RPCUrl                      string
	AccountAddress              gethcommon.Address
	AdminAddress                gethcommon.Address
	PermissionControllerAddress gethcommon.Address
	ChainID                     *big.Int
	Environment                 string
}

type acceptAdminConfig struct {
	Network                     string
	RPCUrl                      string
	AccountAddress              gethcommon.Address
	AcceptorAddress             gethcommon.Address
	PermissionControllerAddress gethcommon.Address
	SignerConfig                types.SignerConfig
	ChainID                     *big.Int
	Environment                 string
	OutputFile                  string
	OutputType                  string
	Broadcast                   bool
}

type addPendingAdminConfig struct {
	Network                     string
	RPCUrl                      string
	AccountAddress              gethcommon.Address
	AdminAddress                gethcommon.Address
	CallerAddress               gethcommon.Address
	PermissionControllerAddress gethcommon.Address
	SignerConfig                types.SignerConfig
	ChainID                     *big.Int
	Environment                 string
	OutputFile                  string
	OutputType                  string
	Broadcast                   bool
}

type removeAdminConfig struct {
	Network                     string
	RPCUrl                      string
	AccountAddress              gethcommon.Address
	AdminAddress                gethcommon.Address
	CallerAddress               gethcommon.Address
	PermissionControllerAddress gethcommon.Address
	SignerConfig                types.SignerConfig
	ChainID                     *big.Int
	Environment                 string
	OutputFile                  string
	OutputType                  string
	Broadcast                   bool
}

type removePendingAdminConfig struct {
	Network                     string
	RPCUrl                      string
	AccountAddress              gethcommon.Address
	AdminAddress                gethcommon.Address
	CallerAddress               gethcommon.Address
	PermissionControllerAddress gethcommon.Address
	SignerConfig                types.SignerConfig
	ChainID                     *big.Int
	Environment                 string
	OutputFile                  string
	OutputType                  string
	Broadcast                   bool
}
