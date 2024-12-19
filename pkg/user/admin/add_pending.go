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

type AddPendingAdminWriter interface {
	AddPendingAdmin(
		ctx context.Context,
		request elcontracts.AddPendingAdminRequest,
	) (*gethtypes.Receipt, error)
	NewAddPendingAdminTx(
		txOpts *bind.TransactOpts,
		request elcontracts.AddPendingAdminRequest,
	) (*gethtypes.Transaction, error)
}

func AddPendingCmd(generator func(logging.Logger, *addPendingAdminConfig) (AddPendingAdminWriter, error)) *cli.Command {
	addPendingCmd := &cli.Command{
		Name:  "add-pending-admin",
		Usage: "Add an admin to be pending until accepted.",
		Action: func(context *cli.Context) error {
			return addPendingAdmin(context, generator)
		},
		After: telemetry.AfterRunAction(),
		Flags: addPendingFlags(),
	}

	return addPendingCmd
}

func addPendingAdmin(
	cliCtx *cli.Context,
	generator func(logging.Logger, *addPendingAdminConfig) (AddPendingAdminWriter, error),
) error {
	ctx := cliCtx.Context
	logger := common.GetLogger(cliCtx)

	config, err := readAndValidateAddPendingAdminConfig(cliCtx, logger)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to read and validate user admin add pending config", err)
	}
	cliCtx.App.Metadata["network"] = config.ChainID.String()
	elWriter, err := generator(logger, config)
	if err != nil {
		return err
	}

	addPendingRequest := elcontracts.AddPendingAdminRequest{
		AccountAddress: config.AccountAddress,
		AdminAddress:   config.AdminAddress,
		WaitForReceipt: true,
	}

	if config.Broadcast {
		return broadcastAddPendingAdminTx(ctx, elWriter, config, addPendingRequest)
	}
	return printAddPendingAdminTx(logger, elWriter, config, addPendingRequest)
}

func printAddPendingAdminTx(
	logger logging.Logger,
	elWriter AddPendingAdminWriter,
	config *addPendingAdminConfig,
	request elcontracts.AddPendingAdminRequest,
) error {
	ethClient, err := ethclient.Dial(config.RPCUrl)
	if err != nil {
		return err
	}
	noSendTxOpts := common.GetNoSendTxOpts(config.CallerAddress)
	if common.IsSmartContractAddress(config.CallerAddress, ethClient) {
		// address is a smart contract
		noSendTxOpts.GasLimit = 150_000
	}

	unsignedTx, err := elWriter.NewAddPendingAdminTx(noSendTxOpts, request)
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
			"Admin %s will be added as pending for account %s\n",
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

func broadcastAddPendingAdminTx(
	ctx context.Context,
	elWriter AddPendingAdminWriter,
	config *addPendingAdminConfig,
	request elcontracts.AddPendingAdminRequest,
) error {
	receipt, err := elWriter.AddPendingAdmin(
		ctx,
		request,
	)
	if err != nil {
		return err
	}
	common.PrintTransactionInfo(receipt.TxHash.String(), config.ChainID)
	return nil
}

func readAndValidateAddPendingAdminConfig(
	cliContext *cli.Context,
	logger logging.Logger,
) (*addPendingAdminConfig, error) {
	accountAddress := gethcommon.HexToAddress(cliContext.String(AccountAddressFlag.Name))
	adminAddress := gethcommon.HexToAddress(cliContext.String(AdminAddressFlag.Name))
	callerAddress := user.PopulateCallerAddress(cliContext, logger, accountAddress)
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

	logger.Debugf(
		"Env: %s, network: %s, chain ID: %s, PermissionManager address: %s",
		environment,
		network,
		chainID,
		permissionManagerAddress,
	)

	return &addPendingAdminConfig{
		Network:                  network,
		RPCUrl:                   ethRpcUrl,
		AccountAddress:           accountAddress,
		AdminAddress:             adminAddress,
		CallerAddress:            callerAddress,
		SignerConfig:             *signerConfig,
		PermissionManagerAddress: gethcommon.HexToAddress(permissionManagerAddress),
		ChainID:                  chainID,
		Environment:              environment,
		OutputFile:               outputFile,
		OutputType:               outputType,
		Broadcast:                broadcast,
	}, nil
}

func generateAddPendingAdminWriter(
	prompter utils.Prompter,
) func(logger logging.Logger, config *addPendingAdminConfig) (AddPendingAdminWriter, error) {
	return func(logger logging.Logger, config *addPendingAdminConfig) (AddPendingAdminWriter, error) {
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

func addPendingFlags() []cli.Flag {
	cmdFlags := []cli.Flag{
		&flags.VerboseFlag,
		&AccountAddressFlag,
		&AdminAddressFlag,
		&user.CallerAddressFlag,
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
