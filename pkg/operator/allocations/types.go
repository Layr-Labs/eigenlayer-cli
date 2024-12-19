package allocations

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"reflect"
	"strings"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/types"

	allocationmanager "github.com/Layr-Labs/eigensdk-go/contracts/bindings/AllocationManager"

	gethcommon "github.com/ethereum/go-ethereum/common"
)

type BulkModifyAllocations struct {
	Allocations           []allocationmanager.IAllocationManagerTypesAllocateParams
	AllocatableMagnitudes map[gethcommon.Address]uint64
}

func (b *BulkModifyAllocations) PrintPretty() {

	fmt.Println()
	fmt.Println("Allocations to be Updated")
	allocations := b.Allocations
	headers := []string{
		"Strategy",
		"Allocatable Magnitude",
		"Operator Set ID",
		"AVS",
		"Magnitude",
	}
	widths := []int{20, 25, 20, 20, 25}

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
		for i, strategy := range a.Strategies {
			fmt.Printf(
				"| %-*s| %-*s| %-*d| %-*s| %-*s|\n",
				widths[0],
				common.ShortEthAddress(strategy),
				widths[1],
				common.FormatNumberWithUnderscores(common.Uint64ToString(b.AllocatableMagnitudes[strategy])),
				widths[2],
				a.OperatorSet.Id,
				widths[3],
				common.ShortEthAddress(a.OperatorSet.Avs),
				widths[4],
				common.FormatNumberWithUnderscores(common.Uint64ToString(a.NewMagnitudes[i])),
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
	callerAddress            gethcommon.Address
	operatorSetId            uint32
	bipsToAllocate           uint64
	signerConfig             *types.SignerConfig
	csvFilePath              string
	isSilent                 bool
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
	callerAddress            gethcommon.Address
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
	StrategyAddress          gethcommon.Address `csv:"strategy_address"`
	AVSAddress               gethcommon.Address `csv:"avs_address"`
	OperatorSetId            uint32             `csv:"operator_set_id"`
	SlashableMagnitude       uint64             `csv:"-"`
	NewMagnitude             uint64             `csv:"-"`
	Shares                   *big.Int           `csv:"shares"`
	SharesPercentage         string             `csv:"shares_percentage"`
	NewAllocationShares      *big.Int           `csv:"new_allocation_shares"`
	UpcomingSharesPercentage string             `csv:"upcoming_shares_percentage"`
	UpdateBlock              uint32             `csv:"update_block"`
}

func (s SlashableMagnitudeHolders) WriteToCSV(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Get fields and their CSV names, excluding skipped fields
	var headers []string
	var fieldIndices []int
	val := reflect.ValueOf(SlashableMagnitudesHolder{})
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		csvTag := field.Tag.Get("csv")

		// Skip if tag is "-"
		if csvTag == "-" {
			continue
		}

		// Use tag value if present, otherwise use field name
		if csvTag != "" {
			headers = append(headers, csvTag)
		} else {
			headers = append(headers, field.Name)
		}
		fieldIndices = append(fieldIndices, i)
	}

	// Write headers
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %v", err)
	}

	// Write data rows
	for _, eachRow := range s {
		val := reflect.ValueOf(eachRow)
		row := make([]string, len(fieldIndices))
		// Only include non-skipped fields
		for i, fieldIndex := range fieldIndices {
			field := val.Field(fieldIndex)
			row[i] = fmt.Sprintf("%v", field.Interface())
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %v", err)
		}
	}

	return nil
}

func (s SlashableMagnitudeHolders) PrintPretty() {
	// Define column headers and widths
	headers := []string{
		"Strategy Address",
		"AVS Address",
		"OperatorSet ID",
		"Slashable Shares (Wei)",
		"Shares %",
		"Upcoming Shares (Wei)",
		"Upcoming Shares %",
		"Update Block",
	}
	widths := []int{len(headers[0]) + 1, len(headers[1]) + 3, 15, 30, 25, 30, 25, 25}

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

		upcomingSharesDisplay := common.FormatNumberWithUnderscores(holder.NewAllocationShares.String())

		fmt.Printf("| %-*s| %-*s| %-*d| %-*s| %-*s| %-*s| %-*s| %-*d|\n",
			widths[0], common.ShortEthAddress(holder.StrategyAddress),
			widths[1], common.ShortEthAddress(holder.AVSAddress),
			widths[2], holder.OperatorSetId,
			widths[3], common.FormatNumberWithUnderscores(holder.Shares.String()),
			widths[4], holder.SharesPercentage+" %",
			widths[5], upcomingSharesDisplay,
			widths[6], holder.UpcomingSharesPercentage+" %",
			widths[7], holder.UpdateBlock,
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

type DeregsiteredOperatorSets []DeregisteredOperatorSet
type DeregisteredOperatorSet struct {
	StrategyAddress    gethcommon.Address
	AVSAddress         gethcommon.Address
	OperatorSetId      uint32
	SlashableMagnitude uint64
	Shares             *big.Int
	SharesPercentage   string
}

func (s DeregsiteredOperatorSets) PrintPretty() {
	// Define column headers and widths
	headers := []string{
		"Strategy Address",
		"AVS Address",
		"OperatorSet ID",
		"Slashable Shares (Wei)",
		"Shares %",
	}
	widths := []int{len(headers[0]) + 1, len(headers[1]) + 3, 15, 30, 25}

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
		fmt.Printf("| %-*s| %-*s| %-*d| %-*s| %-*s|\n",
			widths[0], common.ShortEthAddress(holder.StrategyAddress),
			widths[1], common.ShortEthAddress(holder.AVSAddress),
			widths[2], holder.OperatorSetId,
			widths[3], common.FormatNumberWithUnderscores(holder.Shares.String()),
			widths[4], holder.SharesPercentage+" %",
		)
	}

	// print dashes
	for _, width := range widths {
		fmt.Print("+" + strings.Repeat("-", width+1))
	}
	fmt.Println("+")
}

func (s DeregsiteredOperatorSets) PrintJSON() {
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
