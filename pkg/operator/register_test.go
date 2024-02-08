package operator

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTransactionLink(t *testing.T) {
	var tests = []struct {
		name           string
		chainID        *big.Int
		txHash         string
		expectedTxLink string
	}{
		{
			name:           "Valid goerli tx hash",
			chainID:        big.NewInt(5),
			txHash:         "0x123",
			expectedTxLink: fmt.Sprintf("%s/%s", "https://goerli.etherscan.io/tx", "0x123"),
		},
		{
			name:           "valid mainnet tx hash",
			chainID:        big.NewInt(1),
			txHash:         "0x123",
			expectedTxLink: fmt.Sprintf("%s/%s", "https://etherscan.io/tx", "0x123"),
		},
		{
			name:           "valid holesky tx hash",
			chainID:        big.NewInt(17000),
			txHash:         "0x123",
			expectedTxLink: fmt.Sprintf("%s/%s", "https://holesky.etherscan.io/tx", "0x123"),
		},
		{
			name:           "valid custom chain tx hash",
			chainID:        big.NewInt(100),
			txHash:         "0x123",
			expectedTxLink: "0x123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txLink := getTransactionLink(tt.txHash, tt.chainID)
			assert.Equal(t, tt.expectedTxLink, txLink)
		})
	}
}
