package admin

import (
	"context"
	"fmt"
	"sort"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/user"
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

type RemovePendingAdminWriter interface {
	RemovePendingAdmin(
		ctx context.Context,
		request elcontracts.RemovePendingAdminRequest,
	) (*gethtypes.Receipt, error)
	NewRemovePendingAdminTx(
		txOpts *bind.TransactOpts,
		request elcontracts.RemovePendingAdminRequest,
	) (*gethtypes.Transaction, error)
}

func RemovePendingCmd(
	generator func(logging.Logger, *removePendingAdminConfig) (RemovePendingAdminWriter, error),
) *cli.Command {
	removeCmd := &cli.Command{
		Name:  "remove-pending-admin",
		Usage: "Remove a user who is pending admin acceptance.",
		Action: func(context *cli.Context) error {
			return removePendingAdmin(context, generator)
		},
		After: telemetry.AfterRunAction(),
		Flags: removePendingAdminFlags(),
	}

	return removeCmd
}

func removePendingAdmin(
	cliCtx *cli.Context,
	generator func(logging.Logger, *removePendingAdminConfig) (RemovePendingAdminWriter, error),
) error {
	ctx := cliCtx.Context
	logger := common.GetLogger(cliCtx)

	config, err := readAndValidateRemovePendingAdminConfig(cliCtx, logger)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to read and validate user admin remove pending config", err)
	}
	cliCtx.App.Metadata["network"] = config.ChainID.String()
	elWriter, err := generator(logger, config)
	if err != nil {
		return err
	}
	removeRequest := elcontracts.RemovePendingAdminRequest{
		AccountAddress: config.AccountAddress,
		AdminAddress:   config.AdminAddress,
		WaitForReceipt: true,
	}

	if config.Broadcast {
		return broadcastRemovePendingAdminTx(ctx, elWriter, config, removeRequest)
	}
	return printRemovePendingAdminTx(logger, elWriter, config, removeRequest)
}

func broadcastRemovePendingAdminTx(
	ctx context.Context,
	elWriter RemovePendingAdminWriter,
	config *removePendingAdminConfig,
	request elcontracts.RemovePendingAdminRequest,
) error {
	receipt, err := elWriter.RemovePendingAdmin(ctx, request)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to broadcast RemovePendingAdmin transaction", err)
	}
	common.PrintTransactionInfo(receipt.TxHash.String(), config.ChainID)
	return nil
}

func printRemovePendingAdminTx(
	logger logging.Logger,
	elWriter RemovePendingAdminWriter,
	config *removePendingAdminConfig,
	request elcontracts.RemovePendingAdminRequest,
) error {
	ethClient, err := ethclient.Dial(config.RPCUrl)
	if err != nil {
		return err
	}

	noSendTxOpts := common.GetNoSendTxOpts(config.CallerAddress)
	if common.IsSmartContractAddress(config.CallerAddress, ethClient) {
		noSendTxOpts.GasLimit = 150_000
	}
	unsignedTx, err := elWriter.NewRemovePendingAdminTx(noSendTxOpts, request)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to create unsigned transaction", err)
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
			fmt.Println("Output file not supported for pretty output type")
			fmt.Println()
		}
		fmt.Printf(
			"Pending admin %s will be removed for account %s\n",
			config.AdminAddress,
			config.AccountAddress,
		)
	}

	txFeeDetails := common.GetTxFeeDetails(unsignedTx)
	fmt.Println()
	txFeeDetails.Print()
	fmt.Println("To broadcast the transaction, use the --broadcast flag")
	return nil
}

func readAndValidateRemovePendingAdminConfig(
	cliContext *cli.Context,
	logger logging.Logger,
) (*removePendingAdminConfig, error) {
	accountAddress := gethcommon.HexToAddress(cliContext.String(AccountAddressFlag.Name))
	adminAddress := gethcommon.HexToAddress(cliContext.String(AdminAddressFlag.Name))
	callerAddress := common.PopulateCallerAddress(cliContext, logger, accountAddress, AccountAddressFlag.Name)
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
	permissionControllerAddress := cliContext.String(PermissionControllerAddressFlag.Name)

	if common.IsEmptyString(permissionControllerAddress) {
		permissionControllerAddress, err = common.GetPermissionControllerAddress(utils.NetworkNameToChainId(network))
		if err != nil {
			return nil, err
		}
	}

	logger.Debugf(
		"Env: %s, network: %s, chain ID: %s, PermissionController address: %s",
		environment,
		network,
		chainID,
		permissionControllerAddress,
	)

	return &removePendingAdminConfig{
		Network:                     network,
		RPCUrl:                      ethRpcUrl,
		AccountAddress:              accountAddress,
		AdminAddress:                adminAddress,
		CallerAddress:               callerAddress,
		PermissionControllerAddress: gethcommon.HexToAddress(permissionControllerAddress),
		SignerConfig:                *signerConfig,
		ChainID:                     chainID,
		Environment:                 environment,
		OutputFile:                  outputFile,
		OutputType:                  outputType,
		Broadcast:                   broadcast,
	}, nil
}

func generateRemovePendingAdminWriter(
	prompter utils.Prompter,
) func(logger logging.Logger, config *removePendingAdminConfig) (RemovePendingAdminWriter, error) {
	return func(logger logging.Logger, config *removePendingAdminConfig) (RemovePendingAdminWriter, error) {
		ethClient, err := ethclient.Dial(config.RPCUrl)
		if err != nil {
			return nil, eigenSdkUtils.WrapError("failed to create new eth client", err)
		}
		return common.GetELWriter(
			config.CallerAddress,
			&config.SignerConfig,
			ethClient,
			elcontracts.Config{
				PermissionControllerAddress: config.PermissionControllerAddress,
			},
			prompter,
			config.ChainID,
			logger,
		)
	}
}

func removePendingAdminFlags() []cli.Flag {
	cmdFlags := []cli.Flag{
		&flags.VerboseFlag,
		&AccountAddressFlag,
		&AdminAddressFlag,
		&user.CallerAddressFlag,
		&PermissionControllerAddressFlag,
		&flags.BroadcastFlag,
		&flags.OutputTypeFlag,
		&flags.OutputFileFlag,
		&flags.NetworkFlag,
		&flags.EnvironmentFlag,
		&flags.ETHRpcUrlFlag,
	}
	cmdFlags = append(cmdFlags, flags.GetSignerFlags()...)
	sort.Sort(cli.FlagsByName(cmdFlags))
	return cmdFlags
}
