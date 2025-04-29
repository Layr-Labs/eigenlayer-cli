package allocations

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strings"

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
		5. Get the operator shares for all strategies
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
		6. Using all of the above, get Slashable Shares for the operator
	*/
	slashableSharesMap, err := getSlashableShares(
		ctx,
		config.operatorAddress,
		registeredOperatorSets,
		config.strategyAddresses,
		elReader,
	)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to get slashable shares", err)
	}

	/*
		7. Using all of the above, calculate SlashableMagnitudeHolders object
		for displaying the allocation state of the operator
	*/
	slashableMagnitudeHolders, dergisteredOpsets, err := prepareAllocationsData(
		allAllocations,
		registeredOperatorSetsMap,
		operatorDelegatedSharesMap,
		totalMagnitudeMap,
		slashableSharesMap,
		logger,
	)
	if err != nil {
		return err
	}

	for key, val := range operatorDelegatedSharesMap {
		fmt.Printf("Strategy Address: %s, Shares %s\n", key, common.FormatNumberWithUnderscores(val.String()))
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
	if config.outputType == utils.JsonOutputType {
		slashableMagnitudeHolders.PrintJSON()
	} else {
		if !common.IsEmptyString(config.output) {
			if !strings.HasSuffix(config.output, ".csv") {
				return errors.New("output file must be a .csv file")
			}
			err = slashableMagnitudeHolders.WriteToCSV(config.output)
			if err != nil {
				return err
			}
			logger.Infof("Allocation state written to file: %s", config.output)
		} else {
			slashableMagnitudeHolders.PrintPretty()
		}
	}

	if len(dergisteredOpsets) > 0 {
		fmt.Println()
		fmt.Printf(
			"NOTE: You have %d deregistered operator sets which have nonzero allocations as listed below. Please deallocate to use those funds.\n",
			len(dergisteredOpsets),
		)
		if config.outputType == utils.JsonOutputType {
			dergisteredOpsets.PrintJSON()
		} else {
			dergisteredOpsets.PrintPretty()
		}
	}

	return nil
}

func prepareAllocationsData(
	allAllocations map[string][]elcontracts.AllocationInfo,
	registeredOperatorSetsMap map[string]allocationmanager.OperatorSet,
	operatorDelegatedSharesMap map[string]*big.Int,
	totalMagnitudeMap map[string]uint64,
	slashableSharesMap map[gethcommon.Address]map[string]*big.Int,
	logger logging.Logger,
) (SlashableMagnitudeHolders, DeregsiteredOperatorSets, error) {
	slashableMagnitudeHolders := make(SlashableMagnitudeHolders, 0)
	dergisteredOpsets := make(DeregsiteredOperatorSets, 0)
	for strategy, allocations := range allAllocations {
		logger.Debugf("Strategy: %s, Allocations: %v", strategy, allocations)
		totalStrategyShares := operatorDelegatedSharesMap[strategy]
		totalMagnitude := totalMagnitudeMap[strategy]
		for _, alloc := range allocations {

			// Check if the operator set is not registered and add it to the unregistered list
			// Then skip the rest of the loop
			if _, ok := registeredOperatorSetsMap[getUniqueKey(alloc.AvsAddress, alloc.OperatorSetId)]; !ok {
				currentShares, currentSharesPercentage := getSharesFromMagnitude(
					totalStrategyShares,
					alloc.CurrentMagnitude.Uint64(),
					totalMagnitude,
				)

				// If the operator set is not registered and has no shares, skip it
				// This comes as valid scenario since we iterate first over
				// strategy addresses and then over allocations.
				// This can be fixed by first going over allocations and then over strategy addresses
				// We will fix this in a subsequent PR and improve (TODO: shrimalmadhur)
				if currentShares == nil || currentShares.Cmp(big.NewInt(0)) == 0 {
					continue
				}

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

			// If the total shares in that strategy are zero, skip the operator set
			if totalStrategyShares == nil || totalStrategyShares.Cmp(big.NewInt(0)) == 0 {
				continue
			}
			currentShares := slashableSharesMap[gethcommon.HexToAddress(strategy)][getUniqueKey(alloc.AvsAddress, alloc.OperatorSetId)]
			currentSharesPercentage := getSharePercentage(currentShares, totalStrategyShares)

			newMagnitudeBigInt := big.NewInt(0)
			if alloc.PendingDiff != nil && alloc.PendingDiff.Cmp(big.NewInt(0)) != 0 {
				newMagnitudeBigInt = big.NewInt(0).Add(alloc.CurrentMagnitude, alloc.PendingDiff)
			}

			newShares, newSharesPercentage := getSharesFromMagnitude(
				totalStrategyShares,
				newMagnitudeBigInt.Uint64(),
				totalMagnitude,
			)

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
	return slashableMagnitudeHolders, dergisteredOpsets, nil
}

func getSharePercentage(shares *big.Int, totalShares *big.Int) *big.Float {
	percentageShares := big.NewInt(1)
	percentageShares = percentageShares.Mul(shares, big.NewInt(100))
	percentageSharesFloat := new(
		big.Float,
	).Quo(new(big.Float).SetInt(percentageShares), new(big.Float).SetInt(totalShares))
	return percentageSharesFloat
}

func getSlashableShares(
	ctx context.Context,
	operatorAddress gethcommon.Address,
	opSets []allocationmanager.OperatorSet,
	strategyAddresses []gethcommon.Address,
	reader elChainReader,
) (map[gethcommon.Address]map[string]*big.Int, error) {
	result := make(map[gethcommon.Address]map[string]*big.Int)
	for _, opSet := range opSets {
		slashableSharesMap, err := reader.GetSlashableShares(ctx, operatorAddress, opSet, strategyAddresses)
		if err != nil {
			return nil, err
		}

		for strat, shares := range slashableSharesMap {
			if _, ok := result[strat]; !ok {
				result[strat] = make(map[string]*big.Int)
			}
			result[strat][getUniqueKey(opSet.Avs, opSet.Id)] = shares
		}
	}
	return result, nil
}

func getSharesFromMagnitude(totalShare *big.Int, magnitude uint64, totalMagnitude uint64) (*big.Int, *big.Float) {
	/*
	 * shares = totalShare * magnitude / totalMagnitude
	 * percentageShares = (shares / totalShare) * 100
	 */
	// Check for zero magnitude or totalScaledShare to avoid divide-by-zero errors
	if magnitude == 0 || totalShare.Cmp(big.NewInt(0)) == 0 {
		return big.NewInt(0), big.NewFloat(0)
	}

	opShares := big.NewInt(1)
	opShares = opShares.Set(totalShare)
	shares := opShares.Mul(opShares, big.NewInt(int64(magnitude)))
	shares = shares.Div(shares, big.NewInt(int64(totalMagnitude)))

	percentageShares := big.NewInt(1)
	percentageShares = percentageShares.Mul(opShares, big.NewInt(100))
	percentageSharesFloat := new(
		big.Float,
	).Quo(new(big.Float).SetInt(percentageShares), new(big.Float).SetInt(totalShare))

	return shares, percentageSharesFloat
}

func getUniqueKey(avsAddress gethcommon.Address, opSetId uint32) string {
	return fmt.Sprintf("%s-%d", avsAddress.String(), opSetId)
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
