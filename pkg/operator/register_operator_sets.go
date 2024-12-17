package operator

import (
	"fmt"
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

func RegisterOperatorSetsCommand(p utils.Prompter) *cli.Command {
	registerOperatorSetsCmd := &cli.Command{
		Name:      "register-operator-sets",
		Usage:     "register operator from specified operator sets",
		UsageText: "register-operator-sets [flags]",
		Description: `
register operator sets for operator.

To find what operator set you are registered for, use the 'eigenlayer operator allocations show' command.

`,
		Flags: getRegistrationFlags(),
		After: telemetry.AfterRunAction(),
		Action: func(context *cli.Context) error {
			return registerOperatorSetsAction(context, p)
		},
	}
	return registerOperatorSetsCmd
}

func registerOperatorSetsAction(cCtx *cli.Context, p utils.Prompter) error {
	ctx := cCtx.Context
	logger := common.GetLogger(cCtx)

	config, err := readAndValidateRegisterOperatorSetsConfig(cCtx, logger)
	if err != nil {
		return eigenSdkUtils.WrapError(err, "failed to read and validate register config")
	}

	cCtx.App.Metadata["network"] = config.chainID.String()

	ethClient, err := ethclient.Dial(config.rpcUrl)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to create new eth client", err)
	}

	if config.broadcast {
		if config.signerConfig == nil {
			return fmt.Errorf("signer config is required to broadcast the transaction")
		}
		logger.Info("Signing and broadcasting registration transaction")
		eLWriter, err := common.GetELWriter(
			config.callerAddress,
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
		receipt, err := eLWriter.RegisterForOperatorSets(
			ctx,
			elcontracts.RegistrationRequest{
				AVSAddress:     config.avsAddress,
				OperatorSetIds: config.operatorSetIds,
				WaitForReceipt: true,
			})
		if err != nil {
			return eigenSdkUtils.WrapError("failed to deregister from operator sets", err)
		}
		common.PrintTransactionInfo(receipt.TxHash.String(), config.chainID)
	} else {
		noSendTxOpts := common.GetNoSendTxOpts(config.callerAddress)
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
		if common.IsSmartContractAddress(config.callerAddress, ethClient) {
			// address is a smart contract
			noSendTxOpts.GasLimit = 150_000
		}
		unsignedTx, err := contractBindings.AllocationManager.RegisterForOperatorSets(
			noSendTxOpts,
			config.operatorAddress,
			allocationmanager.IAllocationManagerTypesRegisterParams{
				Avs:            config.avsAddress,
				OperatorSetIds: config.operatorSetIds,
			},
		)
		if err != nil {
			return eigenSdkUtils.WrapError("failed to create unsigned transaction", err)
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
			fmt.Println()
			fmt.Println("Registering from operator sets: ", config.operatorSetIds)
		}
		if !config.isSilent {
			txFeeDetails := common.GetTxFeeDetails(unsignedTx)
			fmt.Println()
			txFeeDetails.Print()
			fmt.Println("To broadcast the transaction, use the --broadcast flag")
		}

	}
	return nil
}

func readAndValidateRegisterOperatorSetsConfig(cCtx *cli.Context, logger logging.Logger) (*RegisterConfig, error) {
	network := cCtx.String(flags.NetworkFlag.Name)
	environment := cCtx.String(flags.EnvironmentFlag.Name)
	logger.Debugf("Using network %s and environment: %s", network, environment)

	rpcUrl := cCtx.String(flags.ETHRpcUrlFlag.Name)
	output := cCtx.String(flags.OutputFileFlag.Name)
	outputType := cCtx.String(flags.OutputTypeFlag.Name)
	broadcast := cCtx.Bool(flags.BroadcastFlag.Name)
	isSilent := cCtx.Bool(flags.SilentFlag.Name)

	operatorAddress := cCtx.String(flags.OperatorAddressFlag.Name)
	callerAddress := cCtx.String(flags.CallerAddressFlag.Name)
	if common.IsEmptyString(callerAddress) {
		logger.Infof("Caller address not provided. Using operator address (%s) as caller address", operatorAddress)
		callerAddress = operatorAddress
	}
	avsAddress := gethcommon.HexToAddress(cCtx.String(flags.AVSAddressFlag.Name))

	// Get signerConfig
	signerConfig, err := common.GetSignerConfig(cCtx, logger)
	if err != nil {
		// We don't want to throw error since people can still use it to generate the claim
		// without broadcasting it
		logger.Debugf("Failed to get signer config: %s", err)
	}

	chainId := utils.NetworkNameToChainId(network)

	delegationManagerAddress := cCtx.String(flags.DelegationManagerAddressFlag.Name)
	if delegationManagerAddress == "" {
		delegationManagerAddress, err = common.GetDelegationManagerAddress(chainId)
		if err != nil {
			return nil, err
		}
	}

	operatorSetIdsString := cCtx.Uint64Slice(flags.OperatorSetIdsFlag.Name)
	operatorSetIds := make([]uint32, len(operatorSetIdsString))
	for i, id := range operatorSetIdsString {
		operatorSetIds[i] = uint32(id)
	}

	config := &RegisterConfig{
		avsAddress:               avsAddress,
		operatorSetIds:           operatorSetIds,
		operatorAddress:          gethcommon.HexToAddress(operatorAddress),
		callerAddress:            gethcommon.HexToAddress(callerAddress),
		network:                  network,
		environment:              environment,
		broadcast:                broadcast,
		rpcUrl:                   rpcUrl,
		chainID:                  chainId,
		signerConfig:             signerConfig,
		output:                   output,
		outputType:               outputType,
		delegationManagerAddress: gethcommon.HexToAddress(delegationManagerAddress),
		isSilent:                 isSilent,
	}

	return config, nil
}

func getRegistrationFlags() []cli.Flag {
	baseFlags := []cli.Flag{
		&flags.NetworkFlag,
		&flags.EnvironmentFlag,
		&flags.ETHRpcUrlFlag,
		&flags.OutputFileFlag,
		&flags.OutputTypeFlag,
		&flags.BroadcastFlag,
		&flags.VerboseFlag,
		&flags.AVSAddressFlag,
		&flags.OperatorAddressFlag,
		&flags.OperatorSetIdsFlag,
		&flags.DelegationManagerAddressFlag,
		&flags.SilentFlag,
		&flags.CallerAddressFlag,
	}

	allFlags := append(baseFlags, flags.GetSignerFlags()...)
	sort.Sort(cli.FlagsByName(allFlags))
	return allFlags
}
