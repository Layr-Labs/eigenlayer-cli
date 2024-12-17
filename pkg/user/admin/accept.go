package admin

import (
	"context"
	"fmt"
	"sort"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	"github.com/Layr-Labs/eigensdk-go/logging"
	eigenSdkUtils "github.com/Layr-Labs/eigensdk-go/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethcommon "github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/urfave/cli/v2"
)

type AcceptAdminWriter interface {
	AcceptAdmin(
		ctx context.Context,
		request elcontracts.AcceptAdminRequest,
	) (*gethtypes.Receipt, error)
	NewAcceptAdminTx(
		txOpts *bind.TransactOpts,
		request elcontracts.AcceptAdminRequest,
	) (*gethtypes.Transaction, error)
}

func AcceptCmd(generator func(logging.Logger, *acceptAdminConfig) (AcceptAdminWriter, error)) *cli.Command {
	acceptCmd := &cli.Command{
		Name:  "accept-admin",
		Usage: "Accepts a user to become admin who is currently pending admin acceptance.",
		Action: func(c *cli.Context) error {
			return acceptAdmin(c, generator)
		},
		After: telemetry.AfterRunAction(),
		Flags: acceptFlags(),
	}

	return acceptCmd
}

func acceptAdmin(
	cliCtx *cli.Context,
	generator func(logging.Logger, *acceptAdminConfig) (AcceptAdminWriter, error),
) error {
	ctx := cliCtx.Context
	logger := common.GetLogger(cliCtx)

	config, err := readAndValidateAcceptAdminConfig(cliCtx, logger)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to read and validate user admin accept config", err)
	}
	cliCtx.App.Metadata["network"] = config.ChainID.String()
	elWriter, err := generator(logger, config)
	if err != nil {
		return err
	}

	acceptRequest := elcontracts.AcceptAdminRequest{
		AccountAddress: config.AccountAddress,
		WaitForReceipt: true,
	}

	if config.Broadcast {
		return broadcastAcceptAdminTx(ctx, elWriter, config, acceptRequest)
	}

	return printAcceptAdminTx(logger, elWriter, config, acceptRequest)
}

func broadcastAcceptAdminTx(
	ctx context.Context,
	elWriter AcceptAdminWriter,
	config *acceptAdminConfig,
	request elcontracts.AcceptAdminRequest,
) error {
	receipt, err := elWriter.AcceptAdmin(ctx, request)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to broadcast AcceptAdmin transaction", err)
	}
	common.PrintTransactionInfo(receipt.TxHash.String(), config.ChainID)
	return nil
}

func printAcceptAdminTx(
	logger logging.Logger,
	elWriter AcceptAdminWriter,
	config *acceptAdminConfig,
	request elcontracts.AcceptAdminRequest,
) error {
	ethClient, err := ethclient.Dial(config.RPCUrl)
	if err != nil {
		return err
	}

	noSendTxOpts := common.GetNoSendTxOpts(config.CallerAddress)
	if common.IsSmartContractAddress(config.CallerAddress, ethClient) {
		noSendTxOpts.GasLimit = 150_000
	}

	// Generate unsigned transaction
	unsignedTx, err := elWriter.NewAcceptAdminTx(noSendTxOpts, request)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to create unsigned tx", err)
	}

	if config.OutputType == string(common.OutputType_Calldata) {
		calldataHex := gethcommon.Bytes2Hex(unsignedTx.Data())
		if !common.IsEmptyString(config.OutputFile) {
			err = common.WriteToFile([]byte(calldataHex), config.OutputFile)
			if err != nil {
				return err
			}
			logger.Infof("Call data written to file: %s", config.OutputFile)
		} else {
			fmt.Println(calldataHex)
		}
	} else {
		if !common.IsEmptyString(config.OutputType) {
			fmt.Println("output file not supported for pretty output type")
			fmt.Println()
		}
		fmt.Printf(
			"Pending admin at address %s will accept admin role\n",
			config.CallerAddress,
		)
	}

	txFeeDetails := common.GetTxFeeDetails(unsignedTx)
	fmt.Println()
	txFeeDetails.Print()
	fmt.Println("To broadcast the transaction, use the --broadcast flag")
	return nil
}

func readAndValidateAcceptAdminConfig(
	cliContext *cli.Context,
	logger logging.Logger,
) (*acceptAdminConfig, error) {
	accountAddress := gethcommon.HexToAddress(cliContext.String(AccountAddressFlag.Name))
	callerAddress := gethcommon.HexToAddress(cliContext.String(CallerAddressFlag.Name))
	ethRpcUrl := cliContext.String(flags.ETHRpcUrlFlag.Name)
	network := cliContext.String(flags.NetworkFlag.Name)
	environment := cliContext.String(flags.EnvironmentFlag.Name)
	outputType := cliContext.String(flags.OutputTypeFlag.Name)
	outputFile := cliContext.String(flags.OutputFileFlag.Name)
	broadcast := cliContext.Bool(flags.BroadcastFlag.Name)
	if environment == "" {
		environment = common.GetEnvFromNetwork(network)
	}
	signerConfig, err := common.GetSignerConfig(cliContext, logger)
	if err != nil {
		// We don't want to throw error since people can still use it to generate the claim
		// without broadcasting it
		logger.Debugf("Failed to get signer config: %s", err)
	}

	chainID := utils.NetworkNameToChainId(network)
	permissionManagerAddress := cliContext.String(PermissionControllerAddressFlag.Name)

	if common.IsEmptyString(permissionManagerAddress) {
		permissionManagerAddress, err = common.GetPermissionManagerAddress(utils.NetworkNameToChainId(network))
		if err != nil {
			return nil, err
		}
	}
	if common.IsEmptyString(callerAddress.String()) {
		logger.Infof(
			"Caller address not provided. Using account address (%s) as caller address",
			accountAddress,
		)
		callerAddress = accountAddress
	}

	logger.Debugf(
		"Env: %s, network: %s, chain ID: %s, PermissionManager address: %s",
		environment,
		network,
		chainID,
		permissionManagerAddress,
	)

	return &acceptAdminConfig{
		Network:                  network,
		RPCUrl:                   ethRpcUrl,
		AccountAddress:           accountAddress,
		CallerAddress:            callerAddress,
		PermissionManagerAddress: gethcommon.HexToAddress(permissionManagerAddress),
		SignerConfig:             *signerConfig,
		ChainID:                  chainID,
		Environment:              environment,
		OutputFile:               outputFile,
		OutputType:               outputType,
		Broadcast:                broadcast,
	}, nil
}

func acceptFlags() []cli.Flag {
	cmdFlags := []cli.Flag{
		&flags.VerboseFlag,
		&AccountAddressFlag,
		&CallerAddressFlag,
		&PermissionControllerAddressFlag,
		&flags.OutputTypeFlag,
		&flags.OutputFileFlag,
		&flags.BroadcastFlag,
		&flags.NetworkFlag,
		&flags.EnvironmentFlag,
		&flags.ETHRpcUrlFlag,
	}
	cmdFlags = append(cmdFlags, flags.GetSignerFlags()...)
	sort.Sort(cli.FlagsByName(cmdFlags))
	return cmdFlags
}

func generateAcceptAdminWriter(
	prompter utils.Prompter,
) func(logger logging.Logger, config *acceptAdminConfig) (AcceptAdminWriter, error) {
	return func(logger logging.Logger, config *acceptAdminConfig) (AcceptAdminWriter, error) {
		ethClient, err := ethclient.Dial(config.RPCUrl)
		if err != nil {
			return nil, eigenSdkUtils.WrapError("failed to create new eth client", err)
		}
		return common.GetELWriter(
			config.CallerAddress,
			&config.SignerConfig,
			ethClient,
			elcontracts.Config{
				PermissionsControllerAddress: config.PermissionManagerAddress,
			},
			prompter,
			config.ChainID,
			logger,
		)
	}
}
