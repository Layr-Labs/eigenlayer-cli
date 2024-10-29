package rewards

import (
	"context"
	"errors"
	"math/big"
	"testing"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestCalculateUnclaimedRewards(t *testing.T) {
	tests := []struct {
		name           string
		allRewards     map[gethcommon.Address]*big.Int
		claimedRewards map[gethcommon.Address]*big.Int
		expected       map[gethcommon.Address]*big.Int
	}{
		{
			name: "Basic case",
			allRewards: map[gethcommon.Address]*big.Int{
				gethcommon.HexToAddress("0x1"): big.NewInt(100),
				gethcommon.HexToAddress("0x2"): big.NewInt(200),
			},
			claimedRewards: map[gethcommon.Address]*big.Int{
				gethcommon.HexToAddress("0x1"): big.NewInt(60),
				gethcommon.HexToAddress("0x2"): big.NewInt(50),
			},
			expected: map[gethcommon.Address]*big.Int{
				gethcommon.HexToAddress("0x1"): big.NewInt(40),
				gethcommon.HexToAddress("0x2"): big.NewInt(150),
			},
		},
		{
			name: "Address with no claimed rewards",
			allRewards: map[gethcommon.Address]*big.Int{
				gethcommon.HexToAddress("0x1"): big.NewInt(100),
				gethcommon.HexToAddress("0x2"): big.NewInt(200),
			},
			claimedRewards: map[gethcommon.Address]*big.Int{
				gethcommon.HexToAddress("0x1"): big.NewInt(60),
				gethcommon.HexToAddress("0x2"): big.NewInt(0),
			},
			expected: map[gethcommon.Address]*big.Int{
				gethcommon.HexToAddress("0x1"): big.NewInt(40),
				gethcommon.HexToAddress("0x2"): big.NewInt(200),
			},
		},
		{
			name: "All rewards claimed",
			allRewards: map[gethcommon.Address]*big.Int{
				gethcommon.HexToAddress("0x1"): big.NewInt(100),
				gethcommon.HexToAddress("0x2"): big.NewInt(200),
			},
			claimedRewards: map[gethcommon.Address]*big.Int{
				gethcommon.HexToAddress("0x1"): big.NewInt(100),
				gethcommon.HexToAddress("0x2"): big.NewInt(200),
			},
			expected: map[gethcommon.Address]*big.Int{
				gethcommon.HexToAddress("0x1"): big.NewInt(0),
				gethcommon.HexToAddress("0x2"): big.NewInt(0),
			},
		},
		{
			name:           "Empty maps",
			allRewards:     map[gethcommon.Address]*big.Int{},
			claimedRewards: map[gethcommon.Address]*big.Int{},
			expected:       map[gethcommon.Address]*big.Int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateUnclaimedRewards(tt.allRewards, tt.claimedRewards)
			for k, v := range result {
				assert.Equal(t, tt.expected[k].Cmp(v), 0)
			}
		})
	}
}

// FakeELReader is a mock implementation of the ELReader interface
type FakeELReader struct {
	claimedRewards map[gethcommon.Address]*big.Int
	shouldError    bool
}

func (f *FakeELReader) GetCumulativeClaimed(
	ctx context.Context,
	earnerAddress, tokenAddress gethcommon.Address,
) (*big.Int, error) {
	if f.shouldError {
		return nil, errors.New("mock error")
	}
	return f.claimedRewards[tokenAddress], nil
}

func TestGetClaimedRewards(t *testing.T) {
	tests := []struct {
		name           string
		earnerAddress  gethcommon.Address
		allRewards     map[gethcommon.Address]*big.Int
		claimedRewards map[gethcommon.Address]*big.Int
		shouldError    bool
		expected       map[gethcommon.Address]*big.Int
		expectedError  bool
	}{
		{
			name:          "Basic case",
			earnerAddress: gethcommon.HexToAddress("0x1"),
			allRewards: map[gethcommon.Address]*big.Int{
				gethcommon.HexToAddress("0x2"): big.NewInt(100),
				gethcommon.HexToAddress("0x3"): big.NewInt(200),
			},
			claimedRewards: map[gethcommon.Address]*big.Int{
				gethcommon.HexToAddress("0x2"): big.NewInt(50),
				gethcommon.HexToAddress("0x3"): big.NewInt(100),
			},
			shouldError: false,
			expected: map[gethcommon.Address]*big.Int{
				gethcommon.HexToAddress("0x2"): big.NewInt(50),
				gethcommon.HexToAddress("0x3"): big.NewInt(100),
			},
			expectedError: false,
		},
		{
			name:          "No claimed rewards",
			earnerAddress: gethcommon.HexToAddress("0x1"),
			allRewards: map[gethcommon.Address]*big.Int{
				gethcommon.HexToAddress("0x2"): big.NewInt(100),
				gethcommon.HexToAddress("0x3"): big.NewInt(200),
			},
			claimedRewards: map[gethcommon.Address]*big.Int{
				gethcommon.HexToAddress("0x2"): nil,
				gethcommon.HexToAddress("0x3"): nil,
			},
			shouldError: false,
			expected: map[gethcommon.Address]*big.Int{
				gethcommon.HexToAddress("0x2"): big.NewInt(0),
				gethcommon.HexToAddress("0x3"): big.NewInt(0),
			},
			expectedError: false,
		},
		{
			name:          "Error case",
			earnerAddress: gethcommon.HexToAddress("0x1"),
			allRewards: map[gethcommon.Address]*big.Int{
				gethcommon.HexToAddress("0x2"): big.NewInt(100),
			},
			claimedRewards: nil,
			shouldError:    true,
			expected:       nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeELReader := &FakeELReader{
				claimedRewards: tt.claimedRewards,
				shouldError:    tt.shouldError,
			}

			result, err := getClaimedRewards(context.Background(), fakeELReader, tt.earnerAddress, tt.allRewards)

			if tt.expectedError {
				assert.Error(t, err, "Expected an error, but got none")
			} else {
				assert.NoError(t, err, "Expected no error, but got one")
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
