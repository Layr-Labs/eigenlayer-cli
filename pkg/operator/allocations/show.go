package allocations

import (
	"fmt"
	"math/big"
	"sort"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	allocationmanager "github.com/Layr-Labs/eigensdk-go/contracts/bindings/AllocationManager"
	"github.com/Layr-Labs/eigensdk-go/logging"
	eigenSdkUtils "github.com/Layr-Labs/eigensdk-go/utils"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/urfave/cli/v2"
)

var (
	// PrecisionFactor comes from the allocation manager contract
	PrecisionFactor = big.NewInt(1e18)
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
	config.delegationManagerAddress = gethcommon.HexToAddress("0x3391eBafDD4b2e84Eeecf1711Ff9FC06EF9Ed182")

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

	/*
		1. Get the allocatable magnitude for all strategies
	*/
	for _, strategyAddress := range config.strategyAddresses {
		allocatableMagnitude, err := elReader.GetAllocatableMagnitude(
			&bind.CallOpts{Context: ctx},
			config.operatorAddress,
			strategyAddress,
		)
		if err != nil {
			return eigenSdkUtils.WrapError("failed to get allocatable magnitude", err)
		}
		logger.Debugf(
			"Allocatable magnitude for strategy %v: %s",
			strategyAddress,
			common.FormatNumberWithUnderscores(common.Uint64ToString(allocatableMagnitude)),
		)
	}

	/*
		2. Get the total magnitude for all strategies
	*/
	totalMagnitudeMap := make(map[string]uint64)
	totalMagnitudes, err := elReader.GetMaxMagnitudes(
		&bind.CallOpts{Context: ctx},
		config.operatorAddress,
		config.strategyAddresses,
	)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to get allocatable magnitude", err)
	}
	for i, strategyAddress := range config.strategyAddresses {
		totalMagnitudeMap[strategyAddress.String()] = totalMagnitudes[i]
	}

	/*
		3. Get allocation info for the operator
	*/

	allAllocations := make(map[string][]elcontracts.AllocationInfo, len(config.strategyAddresses))
	for _, strategyAddress := range config.strategyAddresses {
		allocations, err := elReader.GetAllocationInfo(
			&bind.CallOpts{Context: ctx},
			config.operatorAddress,
			strategyAddress,
		)
		if err != nil {
			return eigenSdkUtils.WrapError("failed to get allocations", err)
		}
		allAllocations[strategyAddress.String()] = allocations
	}

	/*
		6. Get the operator scaled shares for all strategies
	*/
	operatorDelegatedSharesMap := make(map[string]*big.Int)
	shares, err := elReader.GetOperatorShares(&bind.CallOpts{}, config.operatorAddress, config.strategyAddresses)
	for i, strategyAddress := range config.strategyAddresses {
		operatorDelegatedSharesMap[strategyAddress.String()] = shares[i]
	}

	/*
		7. Using all of the above, calculate SlashableMagnitudeHolders object
		   for displaying the allocation state of the operator
	*/
	slashableMagnitudeHolders := make(SlashableMagnitudeHolders, 0)
	for strategy, allocations := range allAllocations {
		strategyShares := operatorDelegatedSharesMap[strategy]
		for _, alloc := range allocations {
			currentShares, currentSharesPercentage := getSharesFromMagnitude(strategyShares, alloc.CurrentMagnitude.Uint64())
			newMagnitudeBigInt := big.NewInt(0)
			if alloc.PendingDiff.Cmp(big.NewInt(0)) != 0 {
				newMagnitudeBigInt = big.NewInt(0).Add(alloc.CurrentMagnitude, alloc.PendingDiff)
			}
			newShares, newSharesPercentage := getSharesFromMagnitude(strategyShares, newMagnitudeBigInt.Uint64())
			slashableMagnitudeHolders = append(slashableMagnitudeHolders, SlashableMagnitudesHolder{
				StrategyAddress:          gethcommon.HexToAddress(strategy),
				AVSAddress:               alloc.AvsAddress,
				OperatorSetId:            alloc.OperatorSetId,
				SlashableMagnitude:       alloc.CurrentMagnitude.Uint64(),
				Shares:                   currentShares,
				SharesPercentage:         currentSharesPercentage.String(),
				NewMagnitude:             newMagnitudeBigInt.Uint64(),
				NewMagnitudeTimestamp:    alloc.CompletableTimestamp,
				NewAllocationShares:      newShares,
				UpcomingSharesPercentage: newSharesPercentage.String(),
			})
		}
	}
	//for i, strategyAddress := range config.strategyAddresses {
	//	slashableMagnitude := slashableMagnitudes[i]
	//	for j, opSet := range opSets {
	//		newAllocation := uint64(0)
	//		newTimestamp := uint32(0)
	//		currSlashableMag := slashableMagnitude[j]
	//		someKey := getUniqueKey(strategyAddress, opSet)
	//
	//		/*
	//			1. Check if there's a pending allocation for this opera
	//		*/
	//		if _, ok := pendingAllocationMap[someKey]; ok {
	//			newAllocation = pendingAllocationMap[someKey].Magnitude
	//			newTimestamp = pendingAllocationMap[someKey].Timestamp
	//		}
	//
	//		/*
	//			2. Check if there's a pending deallocation for this operator
	//		*/
	//		if _, ok := pendingdeAllocationMap[someKey]; ok {
	//			// pendingdeAllocationMap has the magnitude diff for deallocations so we have to
	//			// do some extra math to get the new magnitude
	//			newAllocationDiff := pendingdeAllocationMap[someKey].Magnitude
	//			newTimestamp = pendingdeAllocationMap[someKey].Timestamp
	//			newAllocation = currSlashableMag
	//			currSlashableMag = currSlashableMag + newAllocationDiff
	//		}
	//
	//		operatorScaledShares := operatorDelegatedSharesMap[strategyAddress.String()]
	//
	//		/*
	//			3. Calculate the current shares and percentage shares for the operator
	//		*/
	//		shares, percentShares := getSharesFromMagnitude(operatorScaledShares, currSlashableMag)
	//
	//		/*
	//			4. Calculate the new shares and percentage shares for the operator if any
	//		*/
	//		newShares, newSharesPercentage := getSharesFromMagnitude(operatorScaledShares, newAllocation)
	//
	//		/*
	//			5. Append the SlashableMagnitudeHolder object to the list
	//		*/
	//		slashableMagnitudeHolders = append(slashableMagnitudeHolders, SlashableMagnitudesHolder{
	//			StrategyAddress:          strategyAddress,
	//			AVSAddress:               opSet.Avs,
	//			OperatorSetId:            opSet.OperatorSetId,
	//			SlashableMagnitude:       currSlashableMag,
	//			NewMagnitude:             newAllocation,
	//			NewMagnitudeTimestamp:    newTimestamp,
	//			Shares:                   shares,
	//			SharesPercentage:         percentShares.String(),
	//			NewAllocationShares:      newShares,
	//			UpcomingSharesPercentage: newSharesPercentage.String(),
	//		})
	//	}
	//}

	// Get Operator Shares
	//operatorSharesMap := make(map[string]*big.Int)
	//for _, strategyAddress := range config.strategyAddresses {
	//	shares, err := elReader.GetOperatorShares(&bind.CallOpts{}, config.operatorAddress, strategyAddress)
	//	if err != nil {
	//		return err
	//	}
	//	operatorSharesMap[strategyAddress.String()] = shares
	//}

	for key, val := range operatorDelegatedSharesMap {
		fmt.Printf("Strategy Address: %s, Shares %s\n", key, val.String())
	}

	fmt.Println()
	fmt.Printf("------------------ Allocation State for %s ---------------------\n", config.operatorAddress.String())
	if config.outputType == string(common.OutputType_Json) {
		slashableMagnitudeHolders.PrintJSON()
	} else {
		slashableMagnitudeHolders.PrintPretty()
	}

	return nil
}

func getSharesFromMagnitude(totalScaledShare *big.Int, magnitude uint64) (*big.Int, *big.Float) {

	/*
	 * shares = totalScaledShare * magnitude / PrecisionFactor
	 * percentageShares = (shares / totalScaledShare) * 100
	 */
	// Check for zero magnitude or totalScaledShare to avoid divide-by-zero errors
	if magnitude == 0 || totalScaledShare.Cmp(big.NewInt(0)) == 0 {
		return big.NewInt(0), big.NewFloat(0)
	}

	slashableMagBigInt := big.NewInt(1)
	slashableMagBigInt = slashableMagBigInt.SetUint64(magnitude)

	scaledOpShares := big.NewInt(1)
	scaledOpShares = scaledOpShares.Set(totalScaledShare)
	scaledOpShares = scaledOpShares.Div(scaledOpShares, PrecisionFactor)
	shares := scaledOpShares.Mul(scaledOpShares, slashableMagBigInt)

	percentageShares := big.NewInt(1)
	percentageShares = percentageShares.Mul(scaledOpShares, big.NewInt(100))
	percentageSharesFloat := new(
		big.Float,
	).Quo(new(big.Float).SetInt(percentageShares), new(big.Float).SetInt(totalScaledShare))

	return shares, percentageSharesFloat
}

func getUniqueKey(strategyAddress gethcommon.Address, opSet allocationmanager.OperatorSet) string {
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
