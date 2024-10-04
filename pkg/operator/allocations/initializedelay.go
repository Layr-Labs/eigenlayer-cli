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

func InitializeDelayCmd(p utils.Prompter) *cli.Command {
	initializeDelayCmd := &cli.Command{
		Name:        "initialize-delay",
		UsageText:   "initialize-delay [flags] <delay>",
		Usage:       "Initialize the allocation delay for operator",
		Description: "Initializes the allocation delay for operator. This is a one time command. You can not change the allocation delay once",
		Flags:       getInitializeAllocationDelayFlags(),
		After:       telemetry.AfterRunAction(),
		Action: func(c *cli.Context) error {
			return initializeDelayAction(c, p)
		},
	}

	return initializeDelayCmd
}

func initializeDelayAction(cCtx *cli.Context, p utils.Prompter) error {
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

	// Temp to test modify Allocations
	config.delegationManagerAddress = gethcommon.HexToAddress("0xa5960a80e91D200794ec699b6aBE920908C0e5C5")

	if config.broadcast {
		confirm, err := p.Confirm(
			"This will initialize the allocation delay for operator. You won't be able to set or change it again. Do you want to continue?",
		)
		if err != nil {
			return err
		}
		if !confirm {
			logger.Info("Operation cancelled")
			return nil
		}
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

		receipt, err := eLWriter.InitializeAllocationDelay(ctx, config.operatorAddress, config.allocationDelay, true)
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

func getInitializeAllocationDelayFlags() []cli.Flag {
	baseFlags := []cli.Flag{
		&flags.NetworkFlag,
		&flags.EnvironmentFlag,
		&flags.ETHRpcUrlFlag,
		&flags.OutputFileFlag,
		&flags.OutputTypeFlag,
		&flags.BroadcastFlag,
		&flags.VerboseFlag,
		&flags.OperatorAddressFlag,
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
	allocationDelayInt, err := strconv.Atoi(allocationDelayString)
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

	delegationManagerAddress, err := common.GetDelegationManagerAddress(chainID)
	if err != nil {
		return nil, err
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
		allocationDelay:          uint32(allocationDelayInt),
	}, nil
}
