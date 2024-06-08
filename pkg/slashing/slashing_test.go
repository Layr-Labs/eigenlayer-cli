package slashing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateNewState(t *testing.T) {

	var tests = []struct {
		name                string
		oldState            State
		stakeSource         StakeSource
		operatorSet         int
		slashableProportion float64
		newState            *State
	}{
		{
			name: "Simple case where stake is sourced from slashable stake",
			oldState: State{
				totalMagnitude:     10,
				operatorSets:       []int{1, 2},
				slashableMagnitude: []float64{1, 1},
			},
			stakeSource:         StakeSourceSlashable,
			operatorSet:         4,
			slashableProportion: 0.1,
			newState: &State{
				totalMagnitude:     20.0,
				operatorSets:       []int{4},
				slashableMagnitude: []float64{2.0},
			},
		},
		{
			name: "Simple case where stake is sourced from non slashable stake",
			oldState: State{
				totalMagnitude:     10,
				operatorSets:       []int{1, 2},
				slashableMagnitude: []float64{1, 1},
			},
			stakeSource:         StakeSourceNonSlashable,
			operatorSet:         4,
			slashableProportion: 0.1,
			newState: &State{
				totalMagnitude:     10.0,
				operatorSets:       []int{4},
				slashableMagnitude: []float64{1.0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newState := CalculateNewState(tt.oldState, tt.stakeSource, tt.operatorSet, tt.slashableProportion)
			assert.Equal(t, tt.newState, newState)
		})

	}

}
