package allocations

import (
	"fmt"
	"sort"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	contractIAllocationManager "github.com/Layr-Labs/eigensdk-go/contracts/bindings/IAllocationManager"
	"github.com/Layr-Labs/eigensdk-go/logging"
	eigenSdkUtils "github.com/Layr-Labs/eigensdk-go/utils"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/urfave/cli/v2"
)

func ShowCmd(p utils.Prompter) *cli.Command {
	showCmd := &cli.Command{
		Name:  "show",
		Usage: "Show allocations",
		After: telemetry.AfterRunAction(),
		Description: `
Command to show allocations
`,
		Flags: getShowFlags(),
		Action: func(cCtx *cli.Context) error {
			return showAction(cCtx, p)
		},
	}
	return showCmd
}

func showAction(cCtx *cli.Context, p utils.Prompter) error {
	ctx := cCtx.Context
	logger := common.GetLogger(cCtx)

	config, err := readAndValidateShowConfig(cCtx, &logger)
	if err != nil {
		return err
	}
	cCtx.App.Metadata["network"] = config.chainID.String()

	ethClient, err := ethclient.Dial(config.rpcUrl)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to create new eth client", err)
	}

	// Temp to test modify allocations
	config.delegationManagerAddress = gethcommon.HexToAddress("0x1a597729A7dCfeDDD1f6130fBb099892B7623FAd")

	elReader, err := elcontracts.NewReaderFromConfig(
		elcontracts.Config{
			DelegationManagerAddress: config.delegationManagerAddress,
		},
		ethClient,
		logger,
	)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to create new reader from config", err)
	}

	// for each strategy address, get the allocatable magnitude
	for _, strategyAddress := range config.strategyAddresses {
		allocatableMagnitude, err := elReader.GetAllocatableMagnitude(
			&bind.CallOpts{Context: ctx},
			config.operatorAddress,
			strategyAddress,
		)
		if err != nil {
			return eigenSdkUtils.WrapError("failed to get allocatable magnitude", err)
		}
		logger.Debugf("Allocatable magnitude for strategy %v: %s", strategyAddress, common.FormatNumberWithUnderscores(allocatableMagnitude))
	}

	opSets, slashableMagnitudes, err := elReader.GetCurrentSlashableMagnitudes(
		&bind.CallOpts{Context: ctx},
		config.operatorAddress,
		config.strategyAddresses,
	)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to get slashable magnitude", err)
	}

	// Get Pending allocations
	//pendingAllocationsDetails := make(AllocationDetailsHolder, 0)
	pendingAllocationMap := make(map[string]AllocDetails)
	for _, strategyAddress := range config.strategyAddresses {
		pendingAllocations, timestamps, err := elReader.GetPendingAllocations(
			&bind.CallOpts{Context: ctx},
			config.operatorAddress,
			strategyAddress,
			opSets,
		)
		if err != nil {
			return eigenSdkUtils.WrapError("failed to get pending allocations", err)
		}
		for i, opSet := range opSets {
			pendingAllocation := pendingAllocations[i]
			timestamp := timestamps[i]
			if pendingAllocation == 0 && timestamp == 0 {
				continue
			}
			//pendingAllocationsDetails = append(pendingAllocationsDetails, AllocationDetails{
			//	StrategyAddress: strategyAddress,
			//	AVSAddress:      opSet.Avs,
			//	OperatorSetId:   opSet.OperatorSetId,
			//	Allocation:      pendingAllocation,
			//	Timestamp:       timestamp,
			//})
			pendingAllocationMap[getUniqueKey(strategyAddress, opSet)] = AllocDetails{
				Magnitude: pendingAllocation,
				Timestamp: timestamp,
			}
		}
	}

	//pendingDeallocationsDetails := make(AllocationDetailsHolder, 0)
	pendingdeAllocationMap := make(map[string]AllocDetails)
	for _, strategyAddress := range config.strategyAddresses {
		pendingDeallocations, err := elReader.GetPendingDeallocations(
			&bind.CallOpts{Context: ctx},
			config.operatorAddress,
			strategyAddress,
			opSets,
		)
		if err != nil {
			return eigenSdkUtils.WrapError("failed to get pending deallocations", err)
		}
		for i, opSet := range opSets {
			pendingAllocation := pendingDeallocations[i]
			if pendingAllocation.MagnitudeDiff == 0 && pendingAllocation.CompletableTimestamp == 0 {
				continue
			}
			//pendingDeallocationsDetails = append(pendingDeallocationsDetails, AllocationDetails{
			//	StrategyAddress: strategyAddress,
			//	AVSAddress:      opSet.Avs,
			//	OperatorSetId:   opSet.OperatorSetId,
			//	Allocation:      pendingAllocation.MagnitudeDiff,
			//	Timestamp:       pendingAllocation.CompletableTimestamp,
			//})
			pendingdeAllocationMap[getUniqueKey(strategyAddress, opSet)] = AllocDetails{
				Magnitude: pendingAllocation.MagnitudeDiff,
				Timestamp: pendingAllocation.CompletableTimestamp,
			}
		}
	}

	slashableMagnitudeHolders := make(SlashableMagnitudeHolders, 0)
	for i, strategyAddress := range config.strategyAddresses {
		slashableMagnitude := slashableMagnitudes[i]
		for j, opSet := range opSets {
			newAllocation := uint64(0)
			newTimestamp := uint32(0)
			currSlashableMag := slashableMagnitude[j]
			someKey := getUniqueKey(strategyAddress, opSet)
			if _, ok := pendingAllocationMap[someKey]; ok {
				newAllocation = pendingAllocationMap[someKey].Magnitude
				newTimestamp = pendingAllocationMap[someKey].Timestamp
			}

			if _, ok := pendingdeAllocationMap[someKey]; ok {
				newAllocationDiff := pendingdeAllocationMap[someKey].Magnitude
				newTimestamp = pendingdeAllocationMap[someKey].Timestamp
				newAllocation = currSlashableMag
				currSlashableMag = currSlashableMag + newAllocationDiff
			}
			slashableMagnitudeHolders = append(slashableMagnitudeHolders, SlashableMagnitudesHolder{
				StrategyAddress:       strategyAddress,
				AVSAddress:            opSet.Avs,
				OperatorSetId:         opSet.OperatorSetId,
				SlashableMagnitude:    currSlashableMag,
				NewMagnitude:          newAllocation,
				NewMagnitudeTimestamp: newTimestamp,
			})
		}
	}

	fmt.Println()
	//fmt.Println("------------------Pending Allocations---------------------")
	//if config.outputType == string(common.OutputType_Json) {
	//	pendingAllocationsDetails.PrintJSON()
	//} else {
	//	pendingAllocationsDetails.PrintPretty()
	//}
	//fmt.Println()
	//
	//fmt.Println()
	//fmt.Println("------------------Pending Deallocations---------------------")
	//if config.outputType == string(common.OutputType_Json) {
	//	pendingDeallocationsDetails.PrintJSON()
	//} else {
	//	pendingDeallocationsDetails.PrintPretty()
	//}
	//fmt.Println()

	fmt.Println("------------------Current Slashable Magnitudes---------------------")
	if config.outputType == string(common.OutputType_Json) {
		slashableMagnitudeHolders.PrintJSON()
	} else {
		slashableMagnitudeHolders.PrintPretty()
	}

	return nil
}

func getUniqueKey(strategyAddress gethcommon.Address, opSet contractIAllocationManager.OperatorSet) string {
	return fmt.Sprintf("%s-%s-%d", strategyAddress.String(), opSet.Avs.String(), opSet.OperatorSetId)
}

func readAndValidateShowConfig(cCtx *cli.Context, logger *logging.Logger) (*showConfig, error) {
	network := cCtx.String(flags.NetworkFlag.Name)
	rpcUrl := cCtx.String(flags.ETHRpcUrlFlag.Name)
	environment := cCtx.String(flags.EnvironmentFlag.Name)
	operatorAddress := gethcommon.HexToAddress(cCtx.String(flags.OperatorAddressFlag.Name))
	avsAddresses := common.ConvertStringSliceToGethAddressSlice(cCtx.StringSlice(flags.AVSAddressesFlag.Name))
	strategyAddresses := common.ConvertStringSliceToGethAddressSlice(cCtx.StringSlice(flags.StrategyAddressesFlag.Name))
	outputFile := cCtx.String(flags.OutputFileFlag.Name)
	outputType := cCtx.String(flags.OutputTypeFlag.Name)

	chainId := utils.NetworkNameToChainId(network)
	delegationManagerAddress, err := common.GetDelegationManagerAddress(chainId)
	if err != nil {
		return nil, err
	}

	return &showConfig{
		network:                  network,
		rpcUrl:                   rpcUrl,
		environment:              environment,
		operatorAddress:          operatorAddress,
		avsAddresses:             avsAddresses,
		strategyAddresses:        strategyAddresses,
		output:                   outputFile,
		outputType:               outputType,
		delegationManagerAddress: gethcommon.HexToAddress(delegationManagerAddress),
	}, nil
}

func getShowFlags() []cli.Flag {
	baseFlags := []cli.Flag{
		&flags.OperatorAddressFlag,
		&flags.AVSAddressesFlag,
		&flags.StrategyAddressesFlag,
		&flags.NetworkFlag,
		&flags.EnvironmentFlag,
		&flags.ETHRpcUrlFlag,
		&flags.VerboseFlag,
		&flags.OutputFileFlag,
		&flags.OutputTypeFlag,
	}

	sort.Sort(cli.FlagsByName(baseFlags))
	return baseFlags
}
