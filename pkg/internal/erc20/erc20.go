package erc20

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	UnknownTokenName = "Unknown"
)

// ABI is a simplified ABI for the ERC20 token standard
var ABI = `[{"constant":true,"inputs":[],"name":"name","outputs":[{"name":"","type":"string"}],"type":"function"}]`

// ERC20 is the Go binding of the ERC20 contract
type ERC20 struct {
	Caller // Read-only binding to the contract
}

// Caller is an auto generated read-only Go binding around an Ethereum contract.
type Caller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
func (c *Caller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := c.contract.Call(opts, &out, "name")
	if err != nil {
		return "", err
	}
	return out[0].(string), nil
}

// NewERC20 creates a new instance of ERC20, bound to a specific deployed contract.
func NewERC20(address common.Address, backend bind.ContractBackend) (*ERC20, error) {
	contract, err := bindERC20(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ERC20{Caller: Caller{contract: contract}}, nil
}

// bindERC20 binds a generic wrapper to an already deployed contract.
func bindERC20(
	address common.Address,
	caller bind.ContractCaller,
	transactor bind.ContractTransactor,
	filterer bind.ContractFilterer,
) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

func GetTokenName(tokenAddress common.Address, client *ethclient.Client) string {
	erc20Client, err := NewERC20(tokenAddress, client)
	if err != nil {
		return UnknownTokenName
	}

	name, err := erc20Client.Name(&bind.CallOpts{})
	if err != nil {
		return UnknownTokenName
	}

	return name
}
