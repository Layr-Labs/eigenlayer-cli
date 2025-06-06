package operator

import (
	"fmt"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/command"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/keys"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	allocationmanager "github.com/Layr-Labs/eigensdk-go/contracts/bindings/AllocationManager"
	"github.com/Layr-Labs/eigensdk-go/logging"
	eigenSdkUtils "github.com/Layr-Labs/eigensdk-go/utils"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/urfave/cli/v2"
)

type RegisterOperatorSetCmd struct {
	prompter utils.Prompter
}

func NewRegisterOperatorSetsCmd(p utils.Prompter) *cli.Command {
	delegateCommand := &RegisterOperatorSetCmd{p}
	registerOperatorSetCmd := command.NewWriteableCallDataCommand(
		delegateCommand,
		"register-operator-sets",
		"register operator from specified operator sets",
		"register-operator-sets [flags]",
		"",
		getRegistrationFlags(),
	)

	return registerOperatorSetCmd
}

func (r RegisterOperatorSetCmd) Execute(cCtx *cli.Context) error {
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
			r.prompter,
			config.chainID,
			logger,
		)

		if err != nil {
			return eigenSdkUtils.WrapError("failed to get EL writer", err)
		}
		receipt, err := eLWriter.RegisterForOperatorSets(
			ctx,
			config.registryCoordinatorAddress,
			elcontracts.RegistrationRequest{
				OperatorAddress: config.operatorAddress,
				AVSAddress:      config.avsAddress,
				OperatorSetIds:  config.operatorSetIds,
				BlsKeyPair:      config.blsKeyPair,
				WaitForReceipt:  true,
			})
		if err != nil {
			return eigenSdkUtils.WrapError("failed to register for operator sets", err)
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
		// If caller is a smart contract, we can't estimate gas using geth
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

		if config.outputType == utils.CallDataOutputType {
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

	operatorAddressString := cCtx.String(flags.OperatorAddressFlag.Name)
	if common.IsEmptyString(operatorAddressString) {
		logger.Error("--operator-address flag must be set")
		return nil, fmt.Errorf("Empty operator address provided")
	}
	operatorAddress := gethcommon.HexToAddress(operatorAddressString)

	callerAddress := common.PopulateCallerAddress(cCtx, logger, operatorAddress, operatorAddressString)
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

	registryCoordinatorAddress := cCtx.String(flags.RegistryCoordinatorAddressFlag.Name)
	if common.IsEmptyString(registryCoordinatorAddress) {
		logger.Error("--registry-coordinator-address flag must be set")
		return nil, fmt.Errorf("Empty registry coordinator address provided")
	}

	operatorSetIdsString := cCtx.Uint64Slice(flags.OperatorSetIdsFlag.Name)
	operatorSetIds := make([]uint32, len(operatorSetIdsString))
	for i, id := range operatorSetIdsString {
		operatorSetIds[i] = uint32(id)
	}

	blsPrivateKey := cCtx.String(flags.BlsPrivateKeyFlag.Name)
	if common.IsEmptyString(blsPrivateKey) {
		logger.Error("--bls-private-key flag must be set")
		return nil, fmt.Errorf("Empty BLS private key provided")
	}

	blsKeyPair, err := keys.ParseBlsPrivateKey(blsPrivateKey)
	if err != nil {
		return nil, err
	}

	config := &RegisterConfig{
		avsAddress:                 avsAddress,
		operatorSetIds:             operatorSetIds,
		operatorAddress:            operatorAddress,
		callerAddress:              callerAddress,
		network:                    network,
		environment:                environment,
		broadcast:                  broadcast,
		rpcUrl:                     rpcUrl,
		chainID:                    chainId,
		signerConfig:               signerConfig,
		output:                     output,
		outputType:                 outputType,
		delegationManagerAddress:   gethcommon.HexToAddress(delegationManagerAddress),
		isSilent:                   isSilent,
		registryCoordinatorAddress: gethcommon.HexToAddress(registryCoordinatorAddress),
		blsKeyPair:                 blsKeyPair,
	}

	return config, nil
}

func getRegistrationFlags() []cli.Flag {
	return []cli.Flag{
		&flags.NetworkFlag,
		&flags.EnvironmentFlag,
		&flags.ETHRpcUrlFlag,
		&flags.VerboseFlag,
		&flags.AVSAddressFlag,
		&flags.OperatorAddressFlag,
		&flags.OperatorSetIdsFlag,
		&flags.DelegationManagerAddressFlag,
		&flags.SilentFlag,
		&flags.RegistryCoordinatorAddressFlag,
		&flags.BlsPrivateKeyFlag,
	}
}
