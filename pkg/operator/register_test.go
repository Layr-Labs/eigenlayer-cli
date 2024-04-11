package operator

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/types"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	eigensdkTypes "github.com/Layr-Labs/eigensdk-go/types"

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

func TestValidateMainnetMetadata(t *testing.T) {
	var tests = []struct {
		name        string
		operatorCfg *types.OperatorConfigNew
		expectErr   bool
	}{
		{
			name: "Valid metadata",
			operatorCfg: &types.OperatorConfigNew{
				Operator: eigensdkTypes.Operator{
					MetadataUrl: "https://raw.githubusercontent.com/Layr-Labs/eigendata/master/operators/eigenlabs/metadata.json",
				},
				ChainId: *big.NewInt(utils.MainnetChainId),
			},
		},
		{
			name: "Invalid metadata - invalid logo url",
			operatorCfg: &types.OperatorConfigNew{
				Operator: eigensdkTypes.Operator{
					MetadataUrl: "https://raw.githubusercontent.com/shrimalmadhur/metadata/main/metadata1.json",
				},
				ChainId: *big.NewInt(utils.MainnetChainId),
			},
			expectErr: true,
		},
		{
			name: "Invalid metadata - Invalid metadata url",
			operatorCfg: &types.OperatorConfigNew{
				Operator: eigensdkTypes.Operator{
					MetadataUrl: "https://goerli-operator-metadata.s3.amazonaws.com/metadata.json",
				},
				ChainId: *big.NewInt(utils.MainnetChainId),
			},
			expectErr: true,
		},
		{
			name: "Valid metadata for holesky",
			operatorCfg: &types.OperatorConfigNew{
				Operator: eigensdkTypes.Operator{
					MetadataUrl: "https://goerli-operator-metadata.s3.amazonaws.com/metadata.json",
				},
				ChainId: *big.NewInt(utils.HoleskyChainId),
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMetadata(tt.operatorCfg)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
