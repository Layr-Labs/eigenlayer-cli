package allocations

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/types"

	contractIAllocationManager "github.com/Layr-Labs/eigensdk-go/contracts/bindings/IAllocationManager"

	gethcommon "github.com/ethereum/go-ethereum/common"
)

type BulkModifyAllocations struct {
	Allocations           []contractIAllocationManager.IAllocationManagerMagnitudeAllocation
	AllocatableMagnitudes map[gethcommon.Address]uint64
}

func (b *BulkModifyAllocations) PrintPretty() {

	fmt.Println()
	fmt.Println("Allocations to be Updated")
	allocations := b.Allocations
	headers := []string{
		"Strategy",
		"Expected Total Magnitude",
		"Allocatable Magnitude",
		"Operator Set ID",
		"AVS",
		"Magnitude",
	}
	widths := []int{20, 25, 25, 20, 20, 25}

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
	for _, a := range allocations {
		for i, opSet := range a.OperatorSets {
			fmt.Printf(
				"| %-*s| %-*s| %-*s| %-*d| %-*s| %-*s|\n",
				widths[0],
				common.ShortEthAddress(a.Strategy),
				widths[1],
				common.FormatNumberWithUnderscores(common.Uint64ToString(a.ExpectedTotalMagnitude)),
				widths[2],
				common.FormatNumberWithUnderscores(common.Uint64ToString(b.AllocatableMagnitudes[a.Strategy])),
				widths[3],
				opSet.OperatorSetId,
				widths[4],
				common.ShortEthAddress(opSet.Avs),
				widths[5],
				common.FormatNumberWithUnderscores(common.Uint64ToString(a.Magnitudes[i])),
			)
		}
	}

	// print dashes
	for _, width := range widths {
		fmt.Print("+" + strings.Repeat("-", width+1))
	}

	fmt.Println("+")
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
	StrategyAddress       gethcommon.Address
	AVSAddress            gethcommon.Address
	OperatorSetId         uint32
	SlashableMagnitude    uint64
	NewMagnitude          uint64
	NewAllocationShares   *big.Int
	NewMagnitudeTimestamp uint32
	Shares                *big.Int
	SharesPercentage      string
}

func (s SlashableMagnitudeHolders) PrintPretty() {
	// Define column headers and widths
	headers := []string{
		"Strategy Address",
		"AVS Address",
		"OperatorSet ID",
		"Slashable Shares (Wei)",
		"Shares Percentage",
		"Upcoming Shares (Wei)",
		"Update Time",
	}
	widths := []int{len(headers[0]) + 1, len(headers[1]) + 3, 15, 30, 25, 30, 25}

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
		// Example timestamp (Unix timestamp in seconds)
		timestamp := int64(holder.NewMagnitudeTimestamp)

		// Convert timestamp to time.Time
		t := time.Unix(timestamp, 0)

		upcomingSharesDisplay := common.FormatNumberWithUnderscores(holder.NewAllocationShares.String())

		// Format the time as a string
		formattedTime := t.Format("2006-01-02 15:04:05")
		if holder.NewMagnitudeTimestamp == 0 && holder.NewMagnitude == 0 {
			formattedTime = "N/A"
			upcomingSharesDisplay = "N/A"
		}
		fmt.Printf("| %-*s| %-*s| %-*d| %-*s| %-*s| %-*s| %-*s|\n",
			widths[0], common.ShortEthAddress(holder.StrategyAddress),
			widths[1], common.ShortEthAddress(holder.AVSAddress),
			widths[2], holder.OperatorSetId,
			widths[3], common.FormatNumberWithUnderscores(holder.Shares.String()),
			widths[4], holder.SharesPercentage+" %",
			widths[5], upcomingSharesDisplay,
			widths[6], formattedTime,
		)
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

type AllocationDetails struct {
	StrategyAddress gethcommon.Address
	AVSAddress      gethcommon.Address
	OperatorSetId   uint32
	Allocation      uint64
	Timestamp       uint32
}

type AllocDetails struct {
	Magnitude uint64
	Timestamp uint32
}
