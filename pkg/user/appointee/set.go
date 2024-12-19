package appointee

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/user"
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

type SetAppointeePermissionWriter interface {
	SetPermission(
		ctx context.Context,
		request elcontracts.SetPermissionRequest,
	) (*gethtypes.Receipt, error)
	NewSetPermissionTx(
		txOpts *bind.TransactOpts,
		request elcontracts.SetPermissionRequest,
	) (*gethtypes.Transaction, error)
}

func SetCmd(generator func(logging.Logger, *setConfig) (SetAppointeePermissionWriter, error)) *cli.Command {
	setCmd := &cli.Command{
		Name:  "set",
		Usage: "Grant an appointee a permission.",
		Action: func(c *cli.Context) error {
			return setAppointeePermission(c, generator)
		},
		After: telemetry.AfterRunAction(),
		Flags: setCommandFlags(),
	}

	return setCmd
}

func setAppointeePermission(
	cliCtx *cli.Context,
	generator func(logging.Logger, *setConfig) (SetAppointeePermissionWriter, error),
) error {
	ctx := cliCtx.Context
	logger := common.GetLogger(cliCtx)

	config, err := readAndValidateSetConfig(cliCtx, logger)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to read and validate user appointee set config", err)
	}
	cliCtx.App.Metadata["network"] = config.ChainID.String()
	permissionWriter, err := generator(logger, config)
	if err != nil {
		return err
	}
	return broadcastOrPrint(ctx, logger, permissionWriter, config)
}

func broadcastOrPrint(
	ctx context.Context,
	logger logging.Logger,
	permissionWriter SetAppointeePermissionWriter,
	config *setConfig,
) error {
	permissionRequest := elcontracts.SetPermissionRequest{
		AccountAddress:   config.AccountAddress,
		AppointeeAddress: config.AppointeeAddress,
		Target:           config.Target,
		Selector:         config.Selector,
		WaitForReceipt:   true,
	}
	if config.Broadcast {
		return broadcastSetAppointeeTx(ctx, permissionWriter, config, permissionRequest)
	}
	return printSetAppointeeResults(logger, permissionWriter, config, permissionRequest)
}

func broadcastSetAppointeeTx(
	ctx context.Context,
	permissionWriter SetAppointeePermissionWriter,
	config *setConfig,
	request elcontracts.SetPermissionRequest,
) error {
	receipt, err := permissionWriter.SetPermission(ctx, request)
	if err != nil {
		return err
	}
	common.PrintTransactionInfo(receipt.TxHash.String(), config.ChainID)
	return nil
}

func printSetAppointeeResults(
	logger logging.Logger,
	permissionWriter SetAppointeePermissionWriter,
	config *setConfig,
	request elcontracts.SetPermissionRequest,
) error {
	ethClient, err := ethclient.Dial(config.RPCUrl)
	if err != nil {
		return err
	}
	noSendTxOpts := common.GetNoSendTxOpts(config.CallerAddress)
	if common.IsSmartContractAddress(config.CallerAddress, ethClient) {
		noSendTxOpts.GasLimit = 150_000
	}

	tx, err := permissionWriter.NewSetPermissionTx(noSendTxOpts, request)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to create unsigned tx", err)
	}
	if config.OutputType == string(common.OutputType_Calldata) {
		calldataHex := gethcommon.Bytes2Hex(tx.Data())
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
			"Appointee %s will be given permission to target %s selector %s by account %s\n",
			config.AppointeeAddress,
			config.Target,
			config.Selector,
			config.AccountAddress,
		)
	}
	txFeeDetails := common.GetTxFeeDetails(tx)
	fmt.Println()
	txFeeDetails.Print()
	fmt.Println("To broadcast the transaction, use the --broadcast flag")
	return nil
}

func generateSetAppointeePermissionWriter(
	prompter utils.Prompter,
) func(logger logging.Logger, config *setConfig) (SetAppointeePermissionWriter, error) {
	return func(logger logging.Logger, config *setConfig) (SetAppointeePermissionWriter, error) {
		ethClient, err := ethclient.Dial(config.RPCUrl)
		if err != nil {
			return nil, eigenSdkUtils.WrapError("failed to create new eth client", err)
		}
		elWriter, err := common.GetELWriter(
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
		return elWriter, err
	}
}

func readAndValidateSetConfig(cliContext *cli.Context, logger logging.Logger) (*setConfig, error) {
	accountAddress := gethcommon.HexToAddress(cliContext.String(AccountAddressFlag.Name))
	appointeeAddress := gethcommon.HexToAddress(cliContext.String(AppointeeAddressFlag.Name))
	callerAddress := user.PopulateCallerAddress(cliContext, logger, accountAddress)
	ethRpcUrl := cliContext.String(flags.ETHRpcUrlFlag.Name)
	network := cliContext.String(flags.NetworkFlag.Name)
	environment := cliContext.String(flags.EnvironmentFlag.Name)
	outputType := cliContext.String(flags.OutputTypeFlag.Name)
	outputFile := cliContext.String(flags.OutputFileFlag.Name)
	broadcast := cliContext.Bool(flags.BroadcastFlag.Name)
	target := gethcommon.HexToAddress(cliContext.String(TargetAddressFlag.Name))
	selector := cliContext.String(SelectorFlag.Name)
	selectorBytes, err := common.ValidateAndConvertSelectorString(selector)
	if err != nil {
		return nil, err
	}
	signerConfig, err := common.GetSignerConfig(cliContext, logger)
	if err != nil {
		// We don't want to throw error since people can still use it to generate the claim
		// without broadcasting it
		logger.Debugf("Failed to get signer config: %s", err)
	}

	if environment == "" {
		environment = common.GetEnvFromNetwork(network)
	}

	chainID := utils.NetworkNameToChainId(network)
	cliContext.App.Metadata["network"] = chainID.String()
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

	return &setConfig{
		Network:                  network,
		RPCUrl:                   ethRpcUrl,
		AccountAddress:           accountAddress,
		AppointeeAddress:         appointeeAddress,
		CallerAddress:            callerAddress,
		Target:                   target,
		Selector:                 selectorBytes,
		SignerConfig:             *signerConfig,
		PermissionManagerAddress: gethcommon.HexToAddress(permissionManagerAddress),
		ChainID:                  chainID,
		Environment:              environment,
		OutputFile:               outputFile,
		OutputType:               outputType,
		Broadcast:                broadcast,
	}, nil
}

func setCommandFlags() []cli.Flag {
	cmdFlags := []cli.Flag{
		&flags.VerboseFlag,
		&AccountAddressFlag,
		&AppointeeAddressFlag,
		&user.CallerAddressFlag,
		&TargetAddressFlag,
		&SelectorFlag,
		&PermissionControllerAddressFlag,
		&flags.NetworkFlag,
		&flags.EnvironmentFlag,
		&flags.ETHRpcUrlFlag,
		&flags.BroadcastFlag,
		&flags.OutputTypeFlag,
		&flags.OutputFileFlag,
	}
	cmdFlags = append(cmdFlags, flags.GetSignerFlags()...)
	sort.Sort(cli.FlagsByName(cmdFlags))
	return cmdFlags
}
