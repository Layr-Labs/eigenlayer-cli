package allocations

import (
	"math/big"
	"testing"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	allocationmanager "github.com/Layr-Labs/eigensdk-go/contracts/bindings/AllocationManager"
	"github.com/Layr-Labs/eigensdk-go/testutils"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestPrepareAllocationsData(t *testing.T) {
	avsAddress := gethcommon.HexToAddress("0xa1")

	tests := []struct {
		name                    string
		allAllocations          map[string][]elcontracts.AllocationInfo
		registeredOperatorSets  map[string]allocationmanager.OperatorSet
		operatorDelegatedShares map[string]*big.Int
		totalMagnitude          map[string]uint64
		slashableShares         map[gethcommon.Address]map[string]*big.Int
		expectedSlashable       SlashableMagnitudeHolders
		expectedDeregistered    DeregsiteredOperatorSets
		expectedErr             error
	}{
		{
			name: "Happy path - registered operator set with pending changes",
			allAllocations: map[string][]elcontracts.AllocationInfo{
				"0x1": {
					{
						CurrentMagnitude: big.NewInt(100),
						PendingDiff:      big.NewInt(50),
						EffectBlock:      1000,
						OperatorSetId:    1,
						AvsAddress:       avsAddress,
					},
				},
			},
			registeredOperatorSets: map[string]allocationmanager.OperatorSet{
				getUniqueKey(avsAddress, 1): {
					Avs: gethcommon.HexToAddress("0xa1"),
					Id:  1,
				},
			},
			operatorDelegatedShares: map[string]*big.Int{
				"0x1": big.NewInt(1000),
			},
			totalMagnitude: map[string]uint64{
				"0x1": 200,
			},
			slashableShares: map[gethcommon.Address]map[string]*big.Int{
				gethcommon.HexToAddress("0x1"): {
					getUniqueKey(avsAddress, 1): big.NewInt(500),
				},
			},
			expectedSlashable: SlashableMagnitudeHolders{
				{
					StrategyAddress:          gethcommon.HexToAddress("0x1"),
					AVSAddress:               gethcommon.HexToAddress("0xa1"),
					OperatorSetId:            1,
					SlashableMagnitude:       100,
					NewMagnitude:             150,
					UpdateBlock:              1000,
					Shares:                   big.NewInt(500),
					SharesPercentage:         "50",
					NewAllocationShares:      big.NewInt(750),
					UpcomingSharesPercentage: "75",
				},
			},
			expectedDeregistered: DeregsiteredOperatorSets{},
			expectedErr:          nil,
		},
		{
			name: "Deregistered operator set",
			allAllocations: map[string][]elcontracts.AllocationInfo{
				"0x1": {
					{
						CurrentMagnitude: big.NewInt(100),
						PendingDiff:      big.NewInt(0),
						EffectBlock:      1000,
						OperatorSetId:    1,
						AvsAddress:       gethcommon.HexToAddress("0xa1"),
					},
				},
			},
			registeredOperatorSets: map[string]allocationmanager.OperatorSet{}, // Empty map means operator set is not registered
			operatorDelegatedShares: map[string]*big.Int{
				"0x1": big.NewInt(1000),
			},
			totalMagnitude: map[string]uint64{
				"0x1": 200,
			},
			slashableShares: map[gethcommon.Address]map[string]*big.Int{
				gethcommon.HexToAddress("0x1"): {
					getUniqueKey(avsAddress, 1): big.NewInt(500),
				},
			},
			expectedSlashable: SlashableMagnitudeHolders{},
			expectedDeregistered: DeregsiteredOperatorSets{
				{
					StrategyAddress:    gethcommon.HexToAddress("0x1"),
					AVSAddress:         gethcommon.HexToAddress("0xa1"),
					OperatorSetId:      1,
					SlashableMagnitude: 100,
					Shares:             big.NewInt(500),
					SharesPercentage:   "50",
				},
			},
			expectedErr: nil,
		},
		{
			name: "Zero total shares - should skip",
			allAllocations: map[string][]elcontracts.AllocationInfo{
				"0x1": {
					{
						CurrentMagnitude: big.NewInt(100),
						PendingDiff:      big.NewInt(50),
						EffectBlock:      1000,
						OperatorSetId:    1,
						AvsAddress:       gethcommon.HexToAddress("0xa1"),
					},
				},
			},
			registeredOperatorSets: map[string]allocationmanager.OperatorSet{
				getUniqueKey(avsAddress, 1): {
					Avs: gethcommon.HexToAddress("0xa1"),
					Id:  1,
				},
			},
			operatorDelegatedShares: map[string]*big.Int{
				"0x1": big.NewInt(0), // Zero total shares
			},
			totalMagnitude: map[string]uint64{
				"0x1": 200,
			},
			slashableShares: map[gethcommon.Address]map[string]*big.Int{
				gethcommon.HexToAddress("0x1"): {
					getUniqueKey(avsAddress, 1): big.NewInt(0),
				},
			},
			expectedSlashable:    SlashableMagnitudeHolders{},
			expectedDeregistered: DeregsiteredOperatorSets{},
			expectedErr:          nil,
		},
		{
			name:           "Empty allocations",
			allAllocations: map[string][]elcontracts.AllocationInfo{},
			registeredOperatorSets: map[string]allocationmanager.OperatorSet{
				getUniqueKey(avsAddress, 1): {
					Avs: gethcommon.HexToAddress("0xa1"),
					Id:  1,
				},
			},
			operatorDelegatedShares: map[string]*big.Int{
				"0x1": big.NewInt(1000),
			},
			totalMagnitude: map[string]uint64{
				"0x1": 200,
			},
			slashableShares: map[gethcommon.Address]map[string]*big.Int{
				gethcommon.HexToAddress("0x1"): {
					getUniqueKey(avsAddress, 1): big.NewInt(500),
				},
			},
			expectedSlashable:    SlashableMagnitudeHolders{},
			expectedDeregistered: DeregsiteredOperatorSets{},
			expectedErr:          nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slashable, deregistered, err := prepareAllocationsData(
				tt.allAllocations,
				tt.registeredOperatorSets,
				tt.operatorDelegatedShares,
				tt.totalMagnitude,
				tt.slashableShares,
				testutils.GetTestLogger(),
			)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSlashable, slashable)
				assert.Equal(t, tt.expectedDeregistered, deregistered)
			}
		})
	}
}
