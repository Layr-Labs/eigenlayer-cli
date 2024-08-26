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
			ctx,
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
		ctx,
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
			ctx,
			config.operatorAddress,
			strategyAddress,
		)
		if err != nil {
			return eigenSdkUtils.WrapError("failed to get allocations", err)
		}
		allAllocations[strategyAddress.String()] = allocations
	}

	/*
		4. Get the operator's registered operator sets
	*/
	registeredOperatorSets, err := elReader.GetRegisteredSets(ctx, config.operatorAddress)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to get registered operator sets", err)
	}
	registeredOperatorSetsMap := make(map[string]allocationmanager.OperatorSet)
	for _, opSet := range registeredOperatorSets {
		registeredOperatorSetsMap[getUniqueKey(opSet.Avs, opSet.Id)] = opSet
	}

	/*
		5. Get the operator scaled shares for all strategies
	*/
	operatorDelegatedSharesMap := make(map[string]*big.Int)
	shares, err := elReader.GetOperatorShares(ctx, config.operatorAddress, config.strategyAddresses)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to get operator shares", err)
	}
	for i, strategyAddress := range config.strategyAddresses {
		operatorDelegatedSharesMap[strategyAddress.String()] = shares[i]
	}

	/*
		6. Using all of the above, calculate SlashableMagnitudeHolders object
		   for displaying the allocation state of the operator
	*/
	slashableMagnitudeHolders := make(SlashableMagnitudeHolders, 0)
	dergisteredOpsets := make(DeregsiteredOperatorSets, 0)
	for strategy, allocations := range allAllocations {
		strategyShares := operatorDelegatedSharesMap[strategy]
		for _, alloc := range allocations {
			currentShares, currentSharesPercentage := getSharesFromMagnitude(
				strategyShares,
				alloc.CurrentMagnitude.Uint64(),
			)
			newMagnitudeBigInt := big.NewInt(0)
			if alloc.PendingDiff.Cmp(big.NewInt(0)) != 0 {
				newMagnitudeBigInt = big.NewInt(0).Add(alloc.CurrentMagnitude, alloc.PendingDiff)
			}
			newShares, newSharesPercentage := getSharesFromMagnitude(strategyShares, newMagnitudeBigInt.Uint64())

			// Check if the operator set is not registered and add it to the unregistered list
			// Then skip the rest of the loop
			if _, ok := registeredOperatorSetsMap[getUniqueKey(alloc.AvsAddress, alloc.OperatorSetId)]; !ok {
				dergisteredOpsets = append(dergisteredOpsets, DeregisteredOperatorSet{
					StrategyAddress:    gethcommon.HexToAddress(strategy),
					AVSAddress:         alloc.AvsAddress,
					OperatorSetId:      alloc.OperatorSetId,
					SlashableMagnitude: alloc.CurrentMagnitude.Uint64(),
					Shares:             currentShares,
					SharesPercentage:   currentSharesPercentage.String(),
				})
				continue
			}

			// Add the operator set to the registered list
			slashableMagnitudeHolders = append(slashableMagnitudeHolders, SlashableMagnitudesHolder{
				StrategyAddress:          gethcommon.HexToAddress(strategy),
				AVSAddress:               alloc.AvsAddress,
				OperatorSetId:            alloc.OperatorSetId,
				SlashableMagnitude:       alloc.CurrentMagnitude.Uint64(),
				Shares:                   currentShares,
				SharesPercentage:         currentSharesPercentage.String(),
				NewMagnitude:             newMagnitudeBigInt.Uint64(),
				UpdateBlock:              alloc.EffectBlock,
				NewAllocationShares:      newShares,
				UpcomingSharesPercentage: newSharesPercentage.String(),
			})
		}
	}

	for key, val := range operatorDelegatedSharesMap {
		fmt.Printf("Strategy Address: %s, Shares %s\n", key, val.String())
	}

	currBlockNumber, err := ethClient.BlockNumber(ctx)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to get current block number", err)
	}
	delay, err := elReader.GetAllocationDelay(ctx, config.operatorAddress)
	if err != nil {
		return err
	}
	fmt.Println()
	fmt.Printf("Current allocation delay: %d blocks\n", delay)
	fmt.Println()
	fmt.Printf(
		"------------------ Allocation State for %s (Block: %d) ---------------------\n",
		config.operatorAddress.String(),
		currBlockNumber,
	)
	if config.outputType == string(common.OutputType_Json) {
		slashableMagnitudeHolders.PrintJSON()
	} else {
		slashableMagnitudeHolders.PrintPretty()
	}

	if len(dergisteredOpsets) > 0 {
		fmt.Println()
		fmt.Printf(
			"NOTE: You have %d deregistered operator sets which have nonzero allocations as listed below. Please deallocate to use those funds.\n",
			len(dergisteredOpsets),
		)
		if config.outputType == string(common.OutputType_Json) {
			dergisteredOpsets.PrintJSON()
		} else {
			dergisteredOpsets.PrintPretty()
		}
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

func getUniqueKey(strategyAddress gethcommon.Address, opSetId uint32) string {
	return fmt.Sprintf("%s-%d", strategyAddress.String(), opSetId)
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
	delegationManagerAddress := cCtx.String(flags.DelegationManagerAddressFlag.Name)
	var err error
	if delegationManagerAddress == "" {
		delegationManagerAddress, err = common.GetDelegationManagerAddress(chainId)
		if err != nil {
			return nil, err
		}
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
		&flags.DelegationManagerAddressFlag,
	}

	sort.Sort(cli.FlagsByName(baseFlags))
	return baseFlags
}
