package allocations

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"os"
	"sort"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	contractIAllocationManager "github.com/Layr-Labs/eigensdk-go/contracts/bindings/IAllocationManager"
	"github.com/Layr-Labs/eigensdk-go/logging"
	eigenSdkUtils "github.com/Layr-Labs/eigensdk-go/utils"

	"github.com/gocarina/gocsv"
	"github.com/urfave/cli/v2"
)

type elChainReader interface {
	GetTotalMagnitudes(
		opts *bind.CallOpts,
		operatorAddress gethcommon.Address,
		strategyAddresses []gethcommon.Address,
	) ([]uint64, error)
	GetAllocatableMagnitude(
		opts *bind.CallOpts,
		operator gethcommon.Address,
		strategy gethcommon.Address,
	) (uint64, error)
}

func UpdateCmd(p utils.Prompter) *cli.Command {
	updateCmd := &cli.Command{
		Name:      "update",
		Usage:     "Update allocations",
		UsageText: "update",
		Description: `
		Command to update allocations
		`,
		Flags: getUpdateFlags(),
		After: telemetry.AfterRunAction(),
		Action: func(context *cli.Context) error {
			return updateAllocations(context, p)
		},
	}

	return updateCmd
}

func updateAllocations(cCtx *cli.Context, p utils.Prompter) error {
	ctx := cCtx.Context
	logger := common.GetLogger(cCtx)

	config, err := readAndValidateUpdateFlags(cCtx, logger)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to read and validate update flags", err)
	}
	cCtx.App.Metadata["network"] = config.chainID.String()

	ethClient, err := ethclient.Dial(config.rpcUrl)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to create new eth client", err)
	}

	// Temp to test modify Allocations
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

	allocationsToUpdate, err := generateAllocationsParams(ctx, elReader, config, logger)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to generate Allocations params", err)
	}

	if config.broadcast {
		if config.signerConfig == nil {
			return errors.New("signer is required for broadcasting")
		}
		logger.Info("Broadcasting magnitude allocation update...")
		eLWriter, err := common.GetELWriter(
			config.operatorAddress,
			config.signerConfig,
			ethClient,
			elcontracts.Config{
				DelegationManagerAddress: config.delegationManagerAddress,
			},
			p,
			config.chainID,
			logger,
		)
		if err != nil {
			return eigenSdkUtils.WrapError("failed to get EL writer", err)
		}

		receipt, err := eLWriter.ModifyAllocations(
			ctx,
			config.operatorAddress,
			allocationsToUpdate.Allocations,
			contractIAllocationManager.ISignatureUtilsSignatureWithSaltAndExpiry{
				Expiry: big.NewInt(0),
			},
			true,
		)
		if err != nil {
			return err
		}
		common.PrintTransactionInfo(receipt.TxHash.String(), config.chainID)
	} else {
		noSendTxOpts := common.GetNoSendTxOpts(config.operatorAddress)
		_, _, contractBindings, err := elcontracts.BuildClients(elcontracts.Config{
			DelegationManagerAddress: config.delegationManagerAddress,
		}, ethClient, nil, logger, nil)
		if err != nil {
			return err
		}
		// If operator is a smart contract, we can't estimate gas using geth
		// since balance of contract can be 0, as it can be called by an EOA
		// to claim. So we hardcode the gas limit to 150_000 so that we can
		// create unsigned tx without gas limit estimation from contract bindings
		if common.IsSmartContractAddress(config.operatorAddress, ethClient) {
			// address is a smart contract
			noSendTxOpts.GasLimit = 150_000
		}

		unsignedTx, err := contractBindings.AllocationManager.ModifyAllocations(
			noSendTxOpts,
			config.operatorAddress,
			allocationsToUpdate.Allocations,
			contractIAllocationManager.ISignatureUtilsSignatureWithSaltAndExpiry{
				Expiry: big.NewInt(0),
			},
		)
		if err != nil {
			return eigenSdkUtils.WrapError("failed to create unsigned tx", err)
		}

		if config.outputType == string(common.OutputType_Calldata) {
			calldataHex := gethcommon.Bytes2Hex(unsignedTx.Data())
			if !common.IsEmptyString(config.output) {
				err = common.WriteToFile([]byte(calldataHex), config.output)
				if err != nil {
					return err
				}
				logger.Infof("Call data written to file: %s", config.output)
			} else {
				fmt.Println(calldataHex)
			}
		} else {
			if !common.IsEmptyString(config.output) {
				fmt.Println("output file not supported for pretty output type")
				fmt.Println()
			}
			allocationsToUpdate.PrintPretty()
		}
		txFeeDetails := common.GetTxFeeDetails(unsignedTx)
		fmt.Println()
		txFeeDetails.Print()
		fmt.Println("To broadcast the transaction, use the --broadcast flag")
	}

	return nil
}

func getUpdateFlags() []cli.Flag {
	baseFlags := []cli.Flag{
		&flags.NetworkFlag,
		&flags.EnvironmentFlag,
		&flags.ETHRpcUrlFlag,
		&flags.OutputFileFlag,
		&flags.OutputTypeFlag,
		&flags.BroadcastFlag,
		&flags.VerboseFlag,
		&flags.AVSAddressFlag,
		&flags.StrategyAddressFlag,
		&flags.OperatorAddressFlag,
		&flags.OperatorSetIdFlag,
		&flags.CSVFileFlag,
		&BipsToAllocateFlag,
	}
	allFlags := append(baseFlags, flags.GetSignerFlags()...)
	sort.Sort(cli.FlagsByName(allFlags))
	return allFlags
}

func generateAllocationsParams(
	ctx context.Context,
	elReader elChainReader,
	config *updateConfig,
	logger logging.Logger,
) (*BulkModifyAllocations, error) {
	allocations := make([]contractIAllocationManager.IAllocationManagerMagnitudeAllocation, 0)
	var allocatableMagnitudes map[gethcommon.Address]uint64

	var err error
	if len(config.csvFilePath) == 0 {
		magnitude, err := elReader.GetTotalMagnitudes(
			&bind.CallOpts{Context: ctx},
			config.operatorAddress,
			[]gethcommon.Address{config.strategyAddress},
		)
		if err != nil {
			return nil, eigenSdkUtils.WrapError("failed to get latest total magnitude", err)
		}
		allocatableMagnitude, err := elReader.GetAllocatableMagnitude(
			&bind.CallOpts{Context: ctx},
			config.operatorAddress,
			config.strategyAddress,
		)
		if err != nil {
			return nil, eigenSdkUtils.WrapError("failed to get allocatable magnitude", err)
		}
		logger.Debugf("Total Magnitude: %d", magnitude)
		logger.Debugf("Allocatable Magnitude: %d", allocatableMagnitude)
		logger.Debugf("Bips to allocate: %d", config.bipsToAllocate)
		magnitudeToUpdate := calculateMagnitudeToUpdate(magnitude[0], config.bipsToAllocate)
		logger.Debugf("Magnitude to update: %d", magnitudeToUpdate)
		malloc := contractIAllocationManager.IAllocationManagerMagnitudeAllocation{
			Strategy:               config.strategyAddress,
			ExpectedTotalMagnitude: magnitude[0],
			OperatorSets: []contractIAllocationManager.OperatorSet{
				{
					Avs:           config.avsAddress,
					OperatorSetId: config.operatorSetId,
				},
			},
			Magnitudes: []uint64{magnitudeToUpdate},
		}
		allocations = append(allocations, malloc)
	} else {
		allocations, allocatableMagnitudes, err = computeAllocations(config.csvFilePath, config.operatorAddress, elReader)
		if err != nil {
			return nil, eigenSdkUtils.WrapError("failed to compute allocations", err)
		}
	}

	return &BulkModifyAllocations{
		Allocations:           allocations,
		AllocatableMagnitudes: allocatableMagnitudes,
	}, nil
}

func computeAllocations(
	filePath string,
	operatorAddress gethcommon.Address,
	elReader elChainReader,
) ([]contractIAllocationManager.IAllocationManagerMagnitudeAllocation, map[gethcommon.Address]uint64, error) {
	allocations, err := parseAllocationsCSV(filePath)
	if err != nil {
		return nil, nil, eigenSdkUtils.WrapError("failed to parse allocations csv", err)
	}

	err = validateDataFromCSV(allocations)
	if err != nil {
		return nil, nil, eigenSdkUtils.WrapError("failed to validate data from csv", err)
	}

	strategies := getUniqueStrategies(allocations)
	strategyTotalMagnitudes, err := getMagnitudes(strategies, operatorAddress, elReader)
	if err != nil {
		return nil, nil, eigenSdkUtils.WrapError("failed to get total magnitudes", err)
	}

	allocatableMagnitudePerStrategy, err := parallelGetAllocatableMagnitudes(strategies, operatorAddress, elReader)
	if err != nil {
		return nil, nil, eigenSdkUtils.WrapError("failed to get allocatable magnitudes", err)
	}

	magnitudeAllocations := convertAllocationsToMagnitudeAllocations(allocations, strategyTotalMagnitudes)
	return magnitudeAllocations, allocatableMagnitudePerStrategy, nil
}

func validateDataFromCSV(allocations []allocation) error {
	// check for duplicated (avs_address,operator_set_id,strategy_address)
	tuples := make(map[string]struct{})

	for _, alloc := range allocations {
		tuple := fmt.Sprintf("%s_%d_%s", alloc.AvsAddress.Hex(), alloc.OperatorSetId, alloc.StrategyAddress.Hex())
		if _, exists := tuples[tuple]; exists {
			return fmt.Errorf(
				"duplicate combination found: avs_address=%s, operator_set_id=%d, strategy_address=%s",
				alloc.AvsAddress.Hex(),
				alloc.OperatorSetId,
				alloc.StrategyAddress.Hex(),
			)
		}
		tuples[tuple] = struct{}{}
	}

	return nil
}

func parallelGetAllocatableMagnitudes(
	strategies []gethcommon.Address,
	operatorAddress gethcommon.Address,
	elReader elChainReader,
) (map[gethcommon.Address]uint64, error) {
	strategyAllocatableMagnitudes := make(map[gethcommon.Address]uint64, len(strategies))
	var wg sync.WaitGroup
	errChan := make(chan error, len(strategies))

	for _, s := range strategies {
		wg.Add(1)
		go func(strategy gethcommon.Address) {
			defer wg.Done()
			magnitude, err := elReader.GetAllocatableMagnitude(&bind.CallOpts{}, operatorAddress, strategy)
			if err != nil {
				errChan <- err
				return
			}
			strategyAllocatableMagnitudes[strategy] = magnitude
		}(s)
	}

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		return nil, <-errChan // Return the first error encountered
	}

	return strategyAllocatableMagnitudes, nil
}

func getMagnitudes(
	strategies []gethcommon.Address,
	operatorAddress gethcommon.Address,
	reader elChainReader,
) (map[gethcommon.Address]uint64, error) {
	strategyTotalMagnitudes := make(map[gethcommon.Address]uint64, len(strategies))
	totalMagnitudes, err := reader.GetTotalMagnitudes(
		&bind.CallOpts{Context: context.Background()},
		operatorAddress,
		strategies,
	)
	if err != nil {
		return nil, err
	}
	i := 0
	for _, strategy := range strategies {
		strategyTotalMagnitudes[strategy] = totalMagnitudes[i]
		i++
	}

	return strategyTotalMagnitudes, nil
}

func getUniqueStrategies(allocations []allocation) []gethcommon.Address {
	uniqueStrategies := make(map[gethcommon.Address]struct{})
	for _, a := range allocations {
		uniqueStrategies[a.StrategyAddress] = struct{}{}
	}
	strategies := make([]gethcommon.Address, 0, len(uniqueStrategies))
	for s := range uniqueStrategies {
		strategies = append(strategies, s)
	}
	return strategies
}

func parseAllocationsCSV(filePath string) ([]allocation, error) {
	var allocations []allocation
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	if err := gocsv.UnmarshalFile(file, &allocations); err != nil {
		return nil, err
	}

	return allocations, nil
}

func convertAllocationsToMagnitudeAllocations(
	allocations []allocation,
	strategyTotalMagnitudes map[gethcommon.Address]uint64,
) []contractIAllocationManager.IAllocationManagerMagnitudeAllocation {
	magnitudeAllocations := make([]contractIAllocationManager.IAllocationManagerMagnitudeAllocation, 0)
	operatorSetsPerStragyMap := make(map[gethcommon.Address][]contractIAllocationManager.OperatorSet)
	magnitudeAllocationsPerStrategyMap := make(map[gethcommon.Address][]uint64)
	for _, a := range allocations {
		totalMag := strategyTotalMagnitudes[a.StrategyAddress]
		magnitudeToUpdate := calculateMagnitudeToUpdate(totalMag, a.Bips)

		operatorSets, ok := operatorSetsPerStragyMap[a.StrategyAddress]
		if !ok {
			operatorSets = make([]contractIAllocationManager.OperatorSet, 0)
		}
		operatorSets = append(operatorSets, contractIAllocationManager.OperatorSet{
			Avs:           a.AvsAddress,
			OperatorSetId: a.OperatorSetId,
		})
		operatorSetsPerStragyMap[a.StrategyAddress] = operatorSets

		magnitudes := magnitudeAllocationsPerStrategyMap[a.StrategyAddress]
		magnitudes = append(magnitudes, magnitudeToUpdate)
		magnitudeAllocationsPerStrategyMap[a.StrategyAddress] = magnitudes
	}

	for strategy, operatorSets := range operatorSetsPerStragyMap {
		magnitudeAllocations = append(magnitudeAllocations, contractIAllocationManager.IAllocationManagerMagnitudeAllocation{
			Strategy:               strategy,
			ExpectedTotalMagnitude: strategyTotalMagnitudes[strategy],
			OperatorSets:           operatorSets,
			Magnitudes:             magnitudeAllocationsPerStrategyMap[strategy],
		})
	}
	return magnitudeAllocations
}

func calculateMagnitudeToUpdate(totalMagnitude uint64, bipsToAllocate uint64) uint64 {
	bigMagnitude := big.NewInt(int64(totalMagnitude))
	bigBipsToAllocate := big.NewInt(int64(bipsToAllocate))
	bigBipsMultiplier := big.NewInt(10_000)
	bigMagnitudeToUpdate := bigMagnitude.Mul(bigMagnitude, bigBipsToAllocate).Div(bigMagnitude, bigBipsMultiplier)
	return bigMagnitudeToUpdate.Uint64()
}

func readAndValidateUpdateFlags(cCtx *cli.Context, logger logging.Logger) (*updateConfig, error) {
	network := cCtx.String(flags.NetworkFlag.Name)
	environment := cCtx.String(flags.EnvironmentFlag.Name)
	logger.Debugf("Using network %s and environment: %s", network, environment)

	rpcUrl := cCtx.String(flags.ETHRpcUrlFlag.Name)
	output := cCtx.String(flags.OutputFileFlag.Name)
	outputType := cCtx.String(flags.OutputTypeFlag.Name)
	broadcast := cCtx.Bool(flags.BroadcastFlag.Name)

	operatorAddress := gethcommon.HexToAddress(cCtx.String(flags.OperatorAddressFlag.Name))
	avsAddress := gethcommon.HexToAddress(cCtx.String(flags.AVSAddressFlag.Name))
	strategyAddress := gethcommon.HexToAddress(cCtx.String(flags.StrategyAddressFlag.Name))
	operatorSetId := uint32(cCtx.Uint64(flags.OperatorSetIdFlag.Name))
	bipsToAllocate := cCtx.Uint64(BipsToAllocateFlag.Name)
	logger.Debugf(
		"Operator address: %s, AVS address: %s, Strategy address: %s, Bips to allocate: %d",
		operatorAddress.Hex(),
		avsAddress.Hex(),
		strategyAddress.Hex(),
		bipsToAllocate,
	)

	// Get signerConfig
	signerConfig, err := common.GetSignerConfig(cCtx, logger)
	if err != nil {
		// We don't want to throw error since people can still use it to generate the claim
		// without broadcasting it
		logger.Debugf("Failed to get signer config: %s", err)
	}

	csvFilePath := cCtx.String(flags.CSVFileFlag.Name)
	chainId := utils.NetworkNameToChainId(network)

	delegationManagerAddress, err := common.GetDelegationManagerAddress(chainId)
	if err != nil {
		return nil, err
	}

	return &updateConfig{
		network:                  network,
		rpcUrl:                   rpcUrl,
		environment:              environment,
		output:                   output,
		outputType:               outputType,
		broadcast:                broadcast,
		operatorAddress:          operatorAddress,
		avsAddress:               avsAddress,
		strategyAddress:          strategyAddress,
		bipsToAllocate:           bipsToAllocate,
		signerConfig:             signerConfig,
		csvFilePath:              csvFilePath,
		operatorSetId:            operatorSetId,
		chainID:                  chainId,
		delegationManagerAddress: gethcommon.HexToAddress(delegationManagerAddress),
	}, nil
}
