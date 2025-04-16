package appointee

import (
	"math/big"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/types"
	gethcommon "github.com/ethereum/go-ethereum/common"
)

type canCallConfig struct {
	Network                     string
	RPCUrl                      string
	AccountAddress              gethcommon.Address
	AppointeeAddress            gethcommon.Address
	Target                      gethcommon.Address
	Selector                    [4]byte
	PermissionControllerAddress gethcommon.Address
	ChainID                     *big.Int
	Environment                 string
}

type listAppointeesConfig struct {
	Network                     string
	RPCUrl                      string
	AccountAddress              gethcommon.Address
	Target                      gethcommon.Address
	Selector                    [4]byte
	PermissionControllerAddress gethcommon.Address
	ChainID                     *big.Int
	Environment                 string
}

type listAppointeePermissionsConfig struct {
	Network                     string
	RPCUrl                      string
	AccountAddress              gethcommon.Address
	AppointeeAddress            gethcommon.Address
	PermissionControllerAddress gethcommon.Address
	ChainID                     *big.Int
	Environment                 string
}

type removeConfig struct {
	Network                     string
	RPCUrl                      string
	AccountAddress              gethcommon.Address
	AppointeeAddress            gethcommon.Address
	CallerAddress               gethcommon.Address
	Target                      gethcommon.Address
	SignerConfig                types.SignerConfig
	Selector                    [4]byte
	PermissionControllerAddress gethcommon.Address
	ChainID                     *big.Int
	Environment                 string
	OutputFile                  string
	OutputType                  string
	Broadcast                   bool
}

type setConfig struct {
	Network                     string
	RPCUrl                      string
	AccountAddress              gethcommon.Address
	AppointeeAddress            gethcommon.Address
	CallerAddress               gethcommon.Address
	Target                      gethcommon.Address
	SignerConfig                types.SignerConfig
	Selector                    [4]byte
	PermissionControllerAddress gethcommon.Address
	ChainID                     *big.Int
	Environment                 string
	OutputFile                  string
	OutputType                  string
	Broadcast                   bool
}
