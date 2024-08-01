package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChainIdToNetworkName(t *testing.T) {
	tests := []struct {
		name     string
		chainId  int64
		expected string
	}{
		{
			name:     "mainnet",
			chainId:  1,
			expected: "mainnet",
		},
		{
			name:     "holesky",
			chainId:  17000,
			expected: "holesky",
		},
		{
			name:     "anvil",
			chainId:  31337,
			expected: "anvil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			network := ChainIdToNetworkName(tt.chainId)
			assert.Equal(t, tt.expected, network)
		})
	}
}

func TestNetworkNameToChainId(t *testing.T) {
	tests := []struct {
		name     string
		network  string
		expected int64
	}{
		{
			name:     "mainnet",
			network:  "mainnet",
			expected: 1,
		},
		{
			name:     "holesky",
			network:  "holesky",
			expected: 17000,
		},
		{
			name:     "anvil",
			network:  "anvil",
			expected: 31337,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chainId := NetworkNameToChainId(tt.network)
			assert.Equal(t, tt.expected, chainId.Int64())
		})
	}
}
