package allocations

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	"github.com/Layr-Labs/eigensdk-go/logging"
	eigenSdkUtils "github.com/Layr-Labs/eigensdk-go/utils"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/urfave/cli/v2"
)

func SetDelayCmd(p utils.Prompter) *cli.Command {
	setDelayCmd := &cli.Command{
		Name:        "set-delay",
		UsageText:   "set-delay [flags] <delay>",
		Usage:       "Set the allocation delay for operator in blocks",
		Description: "Set the allocation delay for operator. It will take effect after the delay period",
		Flags:       getSetAllocationDelayFlags(),
		After:       telemetry.AfterRunAction(),
		Action: func(c *cli.Context) error {
			return setDelayAction(c, p)
		},
	}

	return setDelayCmd
}

func setDelayAction(cCtx *cli.Context, p utils.Prompter) error {
	ctx := cCtx.Context
	logger := common.GetLogger(cCtx)

	config, err := readAndValidateAllocationDelayConfig(cCtx, logger)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to read and validate claim config", err)
	}
	cCtx.App.Metadata["network"] = config.chainID.String()
	ethClient, err := ethclient.Dial(config.rpcUrl)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to create new eth client", err)
	}

	if config.broadcast {
		confirm, err := p.Confirm(
			"This will set the allocation delay for operator. Do you want to continue?",
		)
		if err != nil {
			return err
		}
		if !confirm {
			logger.Info("Operation cancelled")
			return nil
		}
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

		receipt, err := eLWriter.SetAllocationDelay(ctx, config.operatorAddress, config.allocationDelay, true)
		if err != nil {
			return err
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

		unsignedTx, err := contractBindings.AllocationManager.SetAllocationDelay(noSendTxOpts, config.operatorAddress, config.allocationDelay)
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
			fmt.Printf("Allocation delay %d will be set for operator %s\n", config.allocationDelay, config.operatorAddress.String())
		}
		txFeeDetails := common.GetTxFeeDetails(unsignedTx)
		fmt.Println()
		txFeeDetails.Print()
		fmt.Println("To broadcast the transaction, use the --broadcast flag")
	}

	return nil
}

func getSetAllocationDelayFlags() []cli.Flag {
	baseFlags := []cli.Flag{
		&flags.NetworkFlag,
		&flags.EnvironmentFlag,
		&flags.ETHRpcUrlFlag,
		&flags.OutputFileFlag,
		&flags.OutputTypeFlag,
		&flags.BroadcastFlag,
		&flags.VerboseFlag,
		&flags.OperatorAddressFlag,
		&flags.DelegationManagerAddressFlag,
		&flags.CallerAddressFlag,
	}
	allFlags := append(baseFlags, flags.GetSignerFlags()...)
	sort.Sort(cli.FlagsByName(allFlags))
	return allFlags
}

func readAndValidateAllocationDelayConfig(c *cli.Context, logger logging.Logger) (*allocationDelayConfig, error) {
	args := c.Args()
	if args.Len() != 1 {
		return nil, fmt.Errorf("accepts 1 arg, received %d", args.Len())
	}

	allocationDelayString := c.Args().First()
	allocationDelayUint, err := strconv.ParseUint(allocationDelayString, 10, 32)
	if err != nil {
		return nil, eigenSdkUtils.WrapError("failed to convert allocation delay to int", err)
	}

	network := c.String(flags.NetworkFlag.Name)
	environment := c.String(EnvironmentFlag.Name)
	rpcUrl := c.String(flags.ETHRpcUrlFlag.Name)
	output := c.String(flags.OutputFileFlag.Name)
	outputType := c.String(flags.OutputTypeFlag.Name)
	broadcast := c.Bool(flags.BroadcastFlag.Name)
	operatorAddress := c.String(flags.OperatorAddressFlag.Name)

	callerAddress := c.String(flags.CallerAddressFlag.Name)
	if common.IsEmptyString(callerAddress) {
		callerAddress = operatorAddress
	}

	chainID := utils.NetworkNameToChainId(network)
	logger.Debugf("Using chain ID: %s", chainID.String())

	if common.IsEmptyString(environment) {
		environment = common.GetEnvFromNetwork(network)
	}
	logger.Debugf("Using network %s and environment: %s", network, environment)

	// Get signerConfig
	signerConfig, err := common.GetSignerConfig(c, logger)
	if err != nil {
		// We don't want to throw error since people can still use it to generate the claim
		// without broadcasting it
		logger.Debugf("Failed to get signer config: %s", err)
	}

	delegationManagerAddress := c.String(flags.DelegationManagerAddressFlag.Name)
	if delegationManagerAddress == "" {
		delegationManagerAddress, err = common.GetDelegationManagerAddress(chainID)
		if err != nil {
			return nil, err
		}
	}

	return &allocationDelayConfig{
		network:                  network,
		rpcUrl:                   rpcUrl,
		environment:              environment,
		chainID:                  chainID,
		output:                   output,
		outputType:               outputType,
		broadcast:                broadcast,
		operatorAddress:          gethcommon.HexToAddress(operatorAddress),
		signerConfig:             signerConfig,
		delegationManagerAddress: gethcommon.HexToAddress(delegationManagerAddress),
		allocationDelay:          uint32(allocationDelayUint),
		callerAddress:            gethcommon.HexToAddress(callerAddress),
	}, nil
}
