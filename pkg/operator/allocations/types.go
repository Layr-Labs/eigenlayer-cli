package allocations

import (
	"encoding/json"
	"fmt"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"math/big"
	"strings"
	"time"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/types"

	contractIAllocationManager "github.com/Layr-Labs/eigensdk-go/contracts/bindings/IAllocationManager"

	gethcommon "github.com/ethereum/go-ethereum/common"
)

type BulkModifyAllocations struct {
	Allocations           []contractIAllocationManager.IAllocationManagerMagnitudeAllocation
	AllocatableMagnitudes map[gethcommon.Address]uint64
}

func (b *BulkModifyAllocations) Print() {
	for _, a := range b.Allocations {
		fmt.Printf(
			"Strategy: %s, Expected Total Magnitude: %d, Allocatable Magnitude %d\n",
			a.Strategy.Hex(),
			a.ExpectedTotalMagnitude,
			b.AllocatableMagnitudes[a.Strategy],
		)
		for i, opSet := range a.OperatorSets {
			fmt.Printf(
				"Operator Set: %d, AVS: %s, Magnitude: %d\n",
				opSet.OperatorSetId,
				opSet.Avs.Hex(),
				a.Magnitudes[i],
			)
		}
		fmt.Println()
	}
}

type updateConfig struct {
	network                  string
	rpcUrl                   string
	environment              string
	chainID                  *big.Int
	output                   string
	outputType               string
	broadcast                bool
	operatorAddress          gethcommon.Address
	avsAddress               gethcommon.Address
	strategyAddress          gethcommon.Address
	delegationManagerAddress gethcommon.Address
	operatorSetId            uint32
	bipsToAllocate           uint64
	signerConfig             *types.SignerConfig
	csvFilePath              string
}

type allocation struct {
	AvsAddress      gethcommon.Address `csv:"avs_address"`
	OperatorSetId   uint32             `csv:"operator_set_id"`
	StrategyAddress gethcommon.Address `csv:"strategy_address"`
	Bips            uint64             `csv:"bips"`
}

type allocationDelayConfig struct {
	network                  string
	rpcUrl                   string
	environment              string
	chainID                  *big.Int
	output                   string
	outputType               string
	broadcast                bool
	operatorAddress          gethcommon.Address
	signerConfig             *types.SignerConfig
	allocationDelay          uint32
	delegationManagerAddress gethcommon.Address
}

type showConfig struct {
	network                  string
	rpcUrl                   string
	environment              string
	chainID                  *big.Int
	output                   string
	outputType               string
	operatorAddress          gethcommon.Address
	delegationManagerAddress gethcommon.Address
	avsAddresses             []gethcommon.Address
	strategyAddresses        []gethcommon.Address
}

type SlashableMagnitudeHolders []SlashableMagnitudesHolder

type SlashableMagnitudesHolder struct {
	StrategyAddress    gethcommon.Address
	AVSAddress         gethcommon.Address
	OperatorSetId      uint32
	SlashableMagnitude uint64
}

func (s SlashableMagnitudeHolders) PrintPretty() {
	// Define column headers and widths
	headers := []string{"Strategy Address", "AVS Address", "Operator Set ID", "Slashable Magnitude"}
	widths := []int{20, 20, 16, 20}

	// print dashes
	for _, width := range widths {
		fmt.Print("+" + strings.Repeat("-", width+1))
	}
	fmt.Println("+")

	// Print header
	for i, header := range headers {
		fmt.Printf("| %-*s", widths[i], header)
	}
	fmt.Println("|")

	// Print separator
	for _, width := range widths {
		fmt.Print("|", strings.Repeat("-", width+1))
	}
	fmt.Println("|")

	// Print data rows
	for _, holder := range s {
		fmt.Printf("| %-*s| %-*s| %-*d| %-*d|\n",
			widths[0], common.ShortEthAddress(holder.StrategyAddress),
			widths[1], common.ShortEthAddress(holder.AVSAddress),
			widths[2], holder.OperatorSetId,
			widths[3], holder.SlashableMagnitude)
	}

	// print dashes
	for _, width := range widths {
		fmt.Print("+" + strings.Repeat("-", width+1))
	}
	fmt.Println("+")
}

func (s SlashableMagnitudeHolders) PrintJSON() {
	obj, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}
	fmt.Println(string(obj))
}

type AllocationDetailsHolder []AllocationDetails

type AllocationDetails struct {
	StrategyAddress gethcommon.Address
	AVSAddress      gethcommon.Address
	OperatorSetId   uint32
	Allocation      uint64
	Timestamp       uint32
}

func (a AllocationDetailsHolder) PrintPretty() {
	// Define column headers and widths
	headers := []string{"Strategy Address", "AVS Address", "Operator Set ID", "Allocation", "Application Timestamp"}
	widths := []int{20, 20, 16, 20, 25}

	// print dashes
	for _, width := range widths {
		fmt.Print("+" + strings.Repeat("-", width+1))
	}
	fmt.Println("+")

	// Print header
	for i, header := range headers {
		fmt.Printf("| %-*s", widths[i], header)
	}
	fmt.Println("|")

	// Print separator
	for _, width := range widths {
		fmt.Print("|", strings.Repeat("-", width+1))
	}
	fmt.Println("|")

	// Print data rows
	for _, holder := range a {
		// Example timestamp (Unix timestamp in seconds)
		timestamp := int64(holder.Timestamp)

		// Convert timestamp to time.Time
		t := time.Unix(timestamp, 0)

		// Format the time as a string
		formattedTime := t.Format("2006-01-02 15:04:05")
		fmt.Printf("| %-*s| %-*s| %-*d| %-*d| %-*s|\n",
			widths[0], common.ShortEthAddress(holder.StrategyAddress),
			widths[1], common.ShortEthAddress(holder.AVSAddress),
			widths[2], holder.OperatorSetId,
			widths[3], holder.Allocation,
			widths[4], formattedTime)
	}

	// print dashes
	for _, width := range widths {
		fmt.Print("+" + strings.Repeat("-", width+1))
	}
	fmt.Println("+")
}

func (a AllocationDetailsHolder) PrintJSON() {
	obj, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}
	fmt.Println(string(obj))
}
