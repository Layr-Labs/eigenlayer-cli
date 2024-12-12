package operator

import (
	"fmt"
	"math"
	"strings"

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

func DeregisterCommand(p utils.Prompter) *cli.Command {
	getDeregisterCmd := &cli.Command{
		Name:      "deregister-operator-sets",
		Usage:     "Deregister operator from specified operator sets",
		UsageText: "deregister-operator-sets [flags]",
		Description: `
Deregister operator from operator sets. 
This command doesn't automatically deallocate your slashable stake from that operator set so you will have to use the 'operator allocations update' command to deallocate your stake from the operator set.

To find what operator set you are part of, use the 'eigenlayer operator allocations show' command.

`,
		Flags: getDeregistrationFlags(),
		After: telemetry.AfterRunAction(),
		Action: func(context *cli.Context) error {
			return deregisterAction(context, p)
		},
	}
	return getDeregisterCmd
}

func deregisterAction(cCtx *cli.Context, p utils.Prompter) error {
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

	operatorAddress := gethcommon.HexToAddress(cCtx.String(flags.OperatorAddressFlag.Name))
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
	}

	allFlags := append(baseFlags, flags.GetSignerFlags()...)
	sort.Sort(cli.FlagsByName(allFlags))
	return allFlags
}
