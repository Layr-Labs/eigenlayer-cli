package common

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const (
	GweiToWei = 1_000_000_000
	EthToWei  = 1_000_000_000_000_000_000
	// This is hardcoded in eigensdk, we will make it configurable in the future
	// so that default exist here and can be overridden by the user
	gasMultiplier = 1.2
)

type TxFeeDetails struct {
	GasLimit      uint64
	CostInEth     float64
	GasTipCapGwei float64
	GasFeeCapGwei float64
}

func (t *TxFeeDetails) Print() {
	message := strings.Repeat("-", 30) + " Gas Fee Details " + strings.Repeat("-", 30)
	fmt.Println(message)
	fmt.Printf("Gas Tip Cap: %0.9f Gwei\n", t.GasTipCapGwei)
	fmt.Printf("Gas Fee Cap: %0.9f Gwei\n", t.GasFeeCapGwei)
	fmt.Printf("Gas Limit: %d (If claimer is a smart contract, this value is hardcoded)\n", t.GasLimit)
	fmt.Printf("Approximate Max Cost of transaction: %0.12f ETH\n", t.CostInEth)
	fmt.Println(strings.Repeat("-", len(message)))
}

func GetTxFeeDetails(tx *types.Transaction) *TxFeeDetails {
	gasTipCapGwei := float64(tx.GasTipCap().Uint64()) / GweiToWei
	gasFeeCapGwei := float64(tx.GasFeeCap().Uint64()) / GweiToWei
	gasLimit := uint64(float64(tx.Gas()) * gasMultiplier)
	cost := new(big.Int).Mul(tx.GasFeeCap(), new(big.Int).SetUint64(gasLimit))
	costInEth := float64(cost.Uint64()) / EthToWei
	return &TxFeeDetails{
		GasLimit:      gasLimit,
		CostInEth:     costInEth,
		GasTipCapGwei: gasTipCapGwei,
		GasFeeCapGwei: gasFeeCapGwei,
	}
}

func ConvertStringSliceToGethAddressSlice(addresses []string) []common.Address {
	gethAddresses := make([]common.Address, 0, len(addresses))
	for _, address := range addresses {
		parsed := common.HexToAddress(address)
		gethAddresses = append(gethAddresses, parsed)
	}
	return gethAddresses
}

func ShortEthAddress(address common.Address) string {
	return fmt.Sprintf("%s...%s", address.Hex()[:6], address.Hex()[len(address.Hex())-4:])
}

func Uint64ToString(num uint64) string {
	return strconv.FormatUint(num, 10)
}

func FormatNumberWithUnderscores(numStr string) string {

	// If the number is less than 1000, no formatting is needed
	if len(numStr) <= 3 {
		return numStr
	}

	// Calculate the number of groups of 3 digits
	groups := (len(numStr) - 1) / 3

	// Create a slice to hold the result
	result := make([]byte, len(numStr)+groups)

	// Fill the result slice from right to left
	resultIndex := len(result) - 1
	for i := len(numStr) - 1; i >= 0; i-- {
		if (len(numStr)-i-1)%3 == 0 && i != len(numStr)-1 {
			result[resultIndex] = '_'
			resultIndex--
		}
		result[resultIndex] = numStr[i]
		resultIndex--
	}

	return string(result)
}
