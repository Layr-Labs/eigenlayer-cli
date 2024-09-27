package allocations

import (
	"context"
	"errors"
	"math"
	"os"
	"testing"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/testutils"

	contractIAllocationManager "github.com/Layr-Labs/eigensdk-go/contracts/bindings/IAllocationManager"
	"github.com/Layr-Labs/eigensdk-go/logging"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/stretchr/testify/assert"
)

const (
	initialMagnitude = 1e18
)

type fakeElChainReader struct {
	// operator --> strategy --> magnitude
	allocatableMagnitudeMap map[gethcommon.Address]map[gethcommon.Address]uint64
	totalMagnitudeMap       map[gethcommon.Address]map[gethcommon.Address]uint64
}

func newFakeElChainReader(
	allocatableMagnitudeMap map[gethcommon.Address]map[gethcommon.Address]uint64,
	totalMagnitudeMap map[gethcommon.Address]map[gethcommon.Address]uint64,
) *fakeElChainReader {
	return &fakeElChainReader{
		allocatableMagnitudeMap: allocatableMagnitudeMap,
		totalMagnitudeMap:       totalMagnitudeMap,
	}
}

func (f *fakeElChainReader) GetTotalMagnitudes(
	opts *bind.CallOpts,
	operator gethcommon.Address,
	strategyAddresses []gethcommon.Address,
) ([]uint64, error) {
	stratMap, ok := f.totalMagnitudeMap[operator]
	if !ok {
		return []uint64{}, errors.New("operator not found")
	}

	// iterate over strategyAddresses and return the corresponding magnitudes
	magnitudes := make([]uint64, 0, len(strategyAddresses))
	for _, strategy := range strategyAddresses {
		magnitude, ok := stratMap[strategy]
		if !ok {
			magnitude = 0
		}
		magnitudes = append(magnitudes, magnitude)
	}
	return magnitudes, nil
}

func (f *fakeElChainReader) GetAllocatableMagnitude(
	opts *bind.CallOpts,
	operator gethcommon.Address,
	strategy gethcommon.Address,
) (uint64, error) {
	stratMap, ok := f.allocatableMagnitudeMap[operator]
	if !ok {
		return initialMagnitude, nil
	}

	magnitude, ok := stratMap[strategy]
	if !ok {
		return initialMagnitude, nil
	}
	return magnitude, nil
}

func TestGenerateAllocationsParams(t *testing.T) {
	avsAddress := testutils.GenerateRandomEthereumAddressString()
	strategyAddress := testutils.GenerateRandomEthereumAddressString()
	operatorAddress := testutils.GenerateRandomEthereumAddressString()
	tests := []struct {
		name                string
		config              *updateConfig
		expectError         bool
		expectedAllocations *BulkModifyAllocations
	}{
		{
			name: "simple single allocation without csv",
			config: &updateConfig{
				operatorAddress: gethcommon.HexToAddress("0x2222AAC0C980Cc029624b7ff55B88Bc6F63C538f"),
				avsAddress:      gethcommon.HexToAddress(avsAddress),
				strategyAddress: gethcommon.HexToAddress(strategyAddress),
				bipsToAllocate:  1000,
				operatorSetId:   1,
			},
			expectError: false,
			expectedAllocations: &BulkModifyAllocations{
				Allocations: []contractIAllocationManager.IAllocationManagerMagnitudeAllocation{
					{
						Strategy:               gethcommon.HexToAddress(strategyAddress),
						ExpectedTotalMagnitude: initialMagnitude,
						OperatorSets: []contractIAllocationManager.OperatorSet{
							{
								OperatorSetId: 1,
								Avs:           gethcommon.HexToAddress(avsAddress),
							},
						},
						Magnitudes: []uint64{1e17},
					},
				},
			},
		},
		{
			name: "csv file allocations1.csv",
			config: &updateConfig{
				csvFilePath:     "testdata/allocations1.csv",
				operatorAddress: gethcommon.HexToAddress(operatorAddress),
			},
			expectError: false,
			expectedAllocations: &BulkModifyAllocations{
				Allocations: []contractIAllocationManager.IAllocationManagerMagnitudeAllocation{
					{
						Strategy:               gethcommon.HexToAddress("0x49989b32351Eb9b8ab2d5623cF22E7F7C23e5630"),
						ExpectedTotalMagnitude: initialMagnitude,
						OperatorSets: []contractIAllocationManager.OperatorSet{
							{
								OperatorSetId: 1,
								Avs:           gethcommon.HexToAddress("0x2222AAC0C980Cc029624b7ff55B88Bc6F63C538f"),
							},
							{
								OperatorSetId: 3,
								Avs:           gethcommon.HexToAddress("0x2222AAC0C980Cc029624b7ff55B88Bc6F63C538f"),
							},
						},
						Magnitudes: []uint64{2e17, 1e17},
					},
					{
						Strategy:               gethcommon.HexToAddress("0x232326fE4F8C2f83E3eB2318F090557b7CD02222"),
						ExpectedTotalMagnitude: initialMagnitude,
						OperatorSets: []contractIAllocationManager.OperatorSet{
							{
								OperatorSetId: 4,
								Avs:           gethcommon.HexToAddress("0x111116fE4F8C2f83E3eB2318F090557b7CD0BF76"),
							},
						},
						Magnitudes: []uint64{3e17},
					},
					{
						Strategy:               gethcommon.HexToAddress("0x545456fE4F8C2f83E3eB2318F090557b7CD04567"),
						ExpectedTotalMagnitude: initialMagnitude,
						OperatorSets: []contractIAllocationManager.OperatorSet{
							{
								OperatorSetId: 5,
								Avs:           gethcommon.HexToAddress("0x111116fE4F8C2f83E3eB2318F090557b7CD0BF76"),
							},
						},
						Magnitudes: []uint64{4e17},
					},
				},
				AllocatableMagnitudes: map[gethcommon.Address]uint64{
					gethcommon.HexToAddress("0x49989b32351Eb9b8ab2d5623cF22E7F7C23e5630"): initialMagnitude,
					gethcommon.HexToAddress("0x232326fE4F8C2f83E3eB2318F090557b7CD02222"): initialMagnitude,
					gethcommon.HexToAddress("0x545456fE4F8C2f83E3eB2318F090557b7CD04567"): initialMagnitude,
				},
			},
		},
		{
			name: "csv file allocations_duplicate.csv",
			config: &updateConfig{
				csvFilePath:     "testdata/allocations_duplicate.csv",
				operatorAddress: gethcommon.HexToAddress(operatorAddress),
			},
			expectError: true,
		},
	}

	elReader := newFakeElChainReader(
		map[gethcommon.Address]map[gethcommon.Address]uint64{
			gethcommon.HexToAddress("0x2222AAC0C980Cc029624b7ff55B88Bc6F63C538f"): {
				gethcommon.HexToAddress(strategyAddress): initialMagnitude,
			},
			gethcommon.HexToAddress(operatorAddress): {
				gethcommon.HexToAddress("0x49989b32351Eb9b8ab2d5623cF22E7F7C23e5630"): initialMagnitude,
				gethcommon.HexToAddress("0x111116fE4F8C2f83E3eB2318F090557b7CD0BF76"): initialMagnitude,
				gethcommon.HexToAddress("0x545456fE4F8C2f83E3eB2318F090557b7CD04567"): initialMagnitude,
			},
		},
		map[gethcommon.Address]map[gethcommon.Address]uint64{
			gethcommon.HexToAddress("0x111116fE4F8C2f83E3eB2318F090557b7CD0BF76"): {
				gethcommon.HexToAddress(strategyAddress): initialMagnitude,
			},
			gethcommon.HexToAddress(operatorAddress): {
				gethcommon.HexToAddress("0x49989b32351Eb9b8ab2d5623cF22E7F7C23e5630"): initialMagnitude,
				gethcommon.HexToAddress("0x111116fE4F8C2f83E3eB2318F090557b7CD0BF76"): initialMagnitude,
				gethcommon.HexToAddress("0x545456fE4F8C2f83E3eB2318F090557b7CD04567"): initialMagnitude,
			},
		},
	)

	logger := logging.NewTextSLogger(os.Stdout, &logging.SLoggerOptions{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allocations, err := generateAllocationsParams(context.Background(), elReader, tt.config, logger)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedAllocations, allocations)
			}
		})
	}
}

func TestCalculateMagnitudeToUpdate(t *testing.T) {
	tests := []struct {
		name              string
		totalMagnitude    uint64
		bipsToAllocate    uint64
		expectedMagnitude uint64
	}{
		{
			name:              "Valid inputs",
			totalMagnitude:    1e18,
			bipsToAllocate:    1000,
			expectedMagnitude: 1e17,
		},
		{
			name:              "Zero total magnitude",
			totalMagnitude:    0,
			bipsToAllocate:    1000,
			expectedMagnitude: 0,
		},
		{
			name:              "Zero bips to allocate",
			totalMagnitude:    1e18,
			bipsToAllocate:    0,
			expectedMagnitude: 0,
		},
		{
			name:              "Max uint64 values",
			totalMagnitude:    math.MaxUint64,
			bipsToAllocate:    math.MaxUint64,
			expectedMagnitude: 0, // Result of overflow
		},
		{
			name:              "Valid inputs 1",
			totalMagnitude:    1e18,
			bipsToAllocate:    2555,
			expectedMagnitude: 2555e14,
		},
		{
			name:              "Valid inputs 2",
			totalMagnitude:    1e18,
			bipsToAllocate:    313,
			expectedMagnitude: 313e14,
		},
		{
			name:              "Valid inputs 3",
			totalMagnitude:    1e18,
			bipsToAllocate:    3,
			expectedMagnitude: 3e14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateMagnitudeToUpdate(tt.totalMagnitude, tt.bipsToAllocate)
			assert.Equal(t, tt.expectedMagnitude, result)
		})
	}
}
