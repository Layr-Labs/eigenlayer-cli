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
	contractIAllocationManager "github.com/Layr-Labs/eigensdk-go/contracts/bindings/IAllocationManager"
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
	config.delegationManagerAddress = gethcommon.HexToAddress("0xFF30144A9A749144e88bEb4FAbF020Cc7F71d2dC")

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
		logger.Debugf(
			"Allocatable magnitude for strategy %v: %s",
			strategyAddress,
			common.FormatNumberWithUnderscores(common.Uint64ToString(allocatableMagnitude)),
		)
	}

	// for each strategy address, get the total magnitude
	totalMagnitudeMap := make(map[string]uint64)
	totalMagnitudes, err := elReader.GetTotalMagnitudes(
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

	opSets, slashableMagnitudes, err := elReader.GetCurrentSlashableMagnitudes(
		&bind.CallOpts{Context: ctx},
		config.operatorAddress,
		config.strategyAddresses,
	)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to get slashable magnitude", err)
	}

	// Get Pending allocations
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
			pendingAllocationMap[getUniqueKey(strategyAddress, opSet)] = AllocDetails{
				Magnitude: pendingAllocation,
				Timestamp: timestamp,
			}
		}
	}

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
			pendingdeAllocationMap[getUniqueKey(strategyAddress, opSet)] = AllocDetails{
				Magnitude: pendingAllocation.MagnitudeDiff,
				Timestamp: pendingAllocation.CompletableTimestamp,
			}
		}
	}

	// Get Operator Shares
	operatorScaledSharesMap := make(map[string]*big.Int)
	for _, strategyAddress := range config.strategyAddresses {
		shares, err := elReader.GetOperatorScaledShares(&bind.CallOpts{}, config.operatorAddress, strategyAddress)
		if err != nil {
			return err
		}
		operatorScaledSharesMap[strategyAddress.String()] = shares
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

			operatorScaledShares := operatorScaledSharesMap[strategyAddress.String()]

			currSlashableMagBigInt := big.NewInt(1)
			currSlashableMagBigInt = currSlashableMagBigInt.SetUint64(currSlashableMag)

			scaledOpShares := big.NewInt(1)
			scaledOpShares = scaledOpShares.Set(operatorScaledShares)
			scaledOpShares = scaledOpShares.Div(scaledOpShares, PrecisionFactor)
			shares := scaledOpShares.Mul(scaledOpShares, currSlashableMagBigInt)

			newShares := getSharesFromMagnitude(operatorScaledShares, newAllocation)

			percentageShares := big.NewInt(1)
			percentageShares = percentageShares.Mul(scaledOpShares, big.NewInt(100))
			percentageSharesFloat := new(
				big.Float,
			).Quo(new(big.Float).SetInt(percentageShares), new(big.Float).SetInt(operatorScaledShares))
			slashableMagnitudeHolders = append(slashableMagnitudeHolders, SlashableMagnitudesHolder{
				StrategyAddress:       strategyAddress,
				AVSAddress:            opSet.Avs,
				OperatorSetId:         opSet.OperatorSetId,
				SlashableMagnitude:    currSlashableMag,
				NewMagnitude:          newAllocation,
				NewMagnitudeTimestamp: newTimestamp,
				Shares:                shares,
				SharesPercentage:      percentageSharesFloat.String(),
				NewAllocationShares:   newShares,
			})
		}
	}

	// Get Operator Shares
	operatorSharesMap := make(map[string]*big.Int)
	for _, strategyAddress := range config.strategyAddresses {
		shares, err := elReader.GetOperatorShares(&bind.CallOpts{}, config.operatorAddress, strategyAddress)
		if err != nil {
			return err
		}
		operatorSharesMap[strategyAddress.String()] = shares
	}

	for key, val := range operatorSharesMap {
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

func getSharesFromMagnitude(totalScaledShare *big.Int, magnitude uint64) *big.Int {
	slashableMagBigInt := big.NewInt(1)
	slashableMagBigInt = slashableMagBigInt.SetUint64(magnitude)

	scaledOpShares := big.NewInt(1)
	scaledOpShares = scaledOpShares.Set(totalScaledShare)
	scaledOpShares = scaledOpShares.Div(scaledOpShares, PrecisionFactor)
	shares := scaledOpShares.Mul(scaledOpShares, slashableMagBigInt)
	return shares
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
