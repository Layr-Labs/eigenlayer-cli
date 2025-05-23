package operator

import (
	"fmt"
	"math"
	"strings"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/command"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	allocationmanager "github.com/Layr-Labs/eigensdk-go/contracts/bindings/AllocationManager"
	"github.com/Layr-Labs/eigensdk-go/logging"
	eigenSdkUtils "github.com/Layr-Labs/eigensdk-go/utils"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/urfave/cli/v2"
)

type DeregisterOperatorSetsCmd struct {
	prompter utils.Prompter
}

func NewDeregisterOperatorSetsCmd(p utils.Prompter) *cli.Command {
	delegateCommand := &DeregisterOperatorSetsCmd{prompter: p}
	deregisterCmd := command.NewWriteableCallDataCommand(
		delegateCommand,
		"deregister-operator-sets",
		"Deregister operator from specified operator sets",
		"deregister-operator-sets [flags]",
		`
		Deregister operator from operator sets. 
		This command doesn't automatically deallocate your slashable stake from that operator set so you will have to use the 'operator allocations update' command to deallocate your stake from the operator set.

		To find what operator set you are part of, use the 'eigenlayer operator allocations show' command.
		`,
		getDeregistrationFlags(),
	)
	return deregisterCmd
}

func (d DeregisterOperatorSetsCmd) Execute(cCtx *cli.Context) error {
	ctx := cCtx.Context
	logger := common.GetLogger(cCtx)

	config, err := readAndValidateDeregisterConfig(cCtx, logger)
	if err != nil {
		return eigenSdkUtils.WrapError(err, "failed to read and validate deregister config")
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
		logger.Info("Signing and broadcasting deregistration transaction")
		eLWriter, err := common.GetELWriter(
			config.callerAddress,
			config.signerConfig,
			ethClient,
			elcontracts.Config{
				DelegationManagerAddress: config.delegationManagerAddress,
			},
			d.prompter,
			config.chainID,
			logger,
		)
		if err != nil {
			return eigenSdkUtils.WrapError("failed to get EL writer", err)
		}
		receipt, err := eLWriter.DeregisterFromOperatorSets(
			ctx,
			config.operatorAddress,
			elcontracts.DeregistrationRequest{
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
		unsignedTx, err := contractBindings.AllocationManager.DeregisterFromOperatorSets(
			noSendTxOpts,
			allocationmanager.IAllocationManagerTypesDeregisterParams{
				Operator:       config.operatorAddress,
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
			fmt.Println("Deregitering from operator sets: ", config.operatorSetIds)
		}
		if !config.isSilent {
			txFeeDetails := common.GetTxFeeDetails(unsignedTx)
			fmt.Println()
			txFeeDetails.Print()
			fmt.Println("To broadcast the transaction, use the --broadcast flag")
			fmt.Println()

			msg1 := "| NOTE: This command doesn't automatically deallocate your slashable stake from that operator set."
			msg2 := "| You will have to use the 'eigenlayer operator allocations update' command to deallocate your stake from the operator set."
			width := int(math.Max(float64(len(msg1)), float64(len(msg2))) + 1)
			fmt.Println("+" + strings.Repeat("-", width) + "+")
			fmt.Println(msg1 + strings.Repeat(" ", width-len(msg1)) + " |")
			fmt.Println(msg2 + strings.Repeat(" ", width-len(msg2)) + " |")
			fmt.Println("+" + strings.Repeat("-", width) + "+")
		}

	}
	return nil
}

func readAndValidateDeregisterConfig(cCtx *cli.Context, logger logging.Logger) (*DeregisterConfig, error) {
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

	operatorSetIdsString := cCtx.Uint64Slice(flags.OperatorSetIdsFlag.Name)
	operatorSetIds := make([]uint32, len(operatorSetIdsString))
	for i, id := range operatorSetIdsString {
		operatorSetIds[i] = uint32(id)
	}

	config := &DeregisterConfig{
		avsAddress:               avsAddress,
		operatorSetIds:           operatorSetIds,
		operatorAddress:          operatorAddress,
		callerAddress:            callerAddress,
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

func getDeregistrationFlags() []cli.Flag {
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
	}
}
