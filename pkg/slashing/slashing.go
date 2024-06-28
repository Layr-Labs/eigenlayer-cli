package slashing

import (
	"fmt"
	"math"
)

type StakeSource string

const (
	StakeSourceSlashable    StakeSource = "slashable"
	StakeSourceNonSlashable StakeSource = "non-slashable"
	StakeSourceBoth         StakeSource = "both"
)

type State struct {
	totalMagnitude     float64
	operatorSets       []int
	slashableMagnitude []float64
}

func CalculateNewState(
	oldState State,
	stakeSource StakeSource,
	operatorSet int,
	slashableProportion float64,
) *State {

	if stakeSource == StakeSourceSlashable {
		// non slashable proportion
		nonSlashableProportion := (oldState.totalMagnitude - sumFloatArray(oldState.slashableMagnitude)) / oldState.totalMagnitude
		fmt.Println("nonSlashableProportion: ", nonSlashableProportion)

		allocatedMagnitude := sumFloatArray(oldState.slashableMagnitude)
		fmt.Println("allocatedMagnitude: ", allocatedMagnitude)

		totalMagnitudeNew := allocatedMagnitude / (1 - slashableProportion - nonSlashableProportion)
		fmt.Println("totalMagnitudeNew: ", totalMagnitudeNew)

		slashableMagnitude := slashableProportion * totalMagnitudeNew
		fmt.Println("slashableMagnitude: ", slashableMagnitude)

		nonSlashableMagnitude := nonSlashableProportion * totalMagnitudeNew
		fmt.Println("nonSlashableMagnitude: ", nonSlashableMagnitude)

		opSetToUpdate := []int{operatorSet}
		slashableMagnitudeSet := []float64{Round(slashableMagnitude, 10)}

		return &State{
			totalMagnitude:     Round(totalMagnitudeNew, 10),
			operatorSets:       opSetToUpdate,
			slashableMagnitude: slashableMagnitudeSet,
		}
	} else if stakeSource == StakeSourceNonSlashable {
		// TODO: need to first verify if the operator set is already in the state
		// and also if there is enough non slashable stake to allocate

		return &State{
			totalMagnitude:     oldState.totalMagnitude,
			operatorSets:       []int{operatorSet},
			slashableMagnitude: []float64{Round(oldState.totalMagnitude*slashableProportion, 10)},
		}
	} else {
		// Stake is sourced from both slashable and non-slashable proportionally
	}

	// non slashable proportion
	nonSlashableProportion := (oldState.totalMagnitude - sumFloatArray(oldState.slashableMagnitude)) / oldState.totalMagnitude
	fmt.Println("nonSlashableProportion: ", nonSlashableProportion)

	allocatedMagnitude := sumFloatArray(oldState.slashableMagnitude)
	fmt.Println("allocatedMagnitude: ", allocatedMagnitude)

	/*
		i = new operator set
		slashablePercentage = slashable magnitude (i) / total magnitude new

	*/
	totalMagnitudeNew := allocatedMagnitude / (1 - slashableProportion - nonSlashableProportion)
	fmt.Println("totalMagnitudeNew: ", totalMagnitudeNew)

	slashableMagnitude := slashableProportion * totalMagnitudeNew
	fmt.Println("slashableMagnitude: ", slashableMagnitude)

	nonSlashableMagnitude := nonSlashableProportion * totalMagnitudeNew
	fmt.Println("nonSlashableMagnitude: ", nonSlashableMagnitude)

	opSetToUpdate := []int{operatorSet}
	slashableMagnitudeSet := []float64{Round(slashableMagnitude, 10)}

	return &State{
		totalMagnitude:     Round(totalMagnitudeNew, 10),
		operatorSets:       opSetToUpdate,
		slashableMagnitude: slashableMagnitudeSet,
	}
}

func sumFloatArray(arr []float64) float64 {
	sum := 0.0
	for _, i := range arr {
		sum += i
	}
	return sum
}

// Round rounds the floating point number to the specified number of decimal places.
func Round(val float64, places int) float64 {
	if places < 0 {
		return val
	}
	factor := math.Pow(10, float64(places))
	return math.Round(val*factor) / factor
}
