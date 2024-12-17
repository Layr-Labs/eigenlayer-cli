package admin

import (
	"context"
	"sort"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	"github.com/Layr-Labs/eigensdk-go/logging"
	eigenSdkUtils "github.com/Layr-Labs/eigensdk-go/utils"
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

	receipt, err := elWriter.AcceptAdmin(
		ctx,
		elcontracts.AcceptAdminRequest{AccountAddress: config.AccountAddress, WaitForReceipt: true},
	)
	if err != nil {
		return err
	}
	common.PrintTransactionInfo(receipt.TxHash.String(), config.ChainID)
	return nil
}

func readAndValidateAcceptAdminConfig(
	cliContext *cli.Context,
	logger logging.Logger,
) (*acceptAdminConfig, error) {
	accountAddress := gethcommon.HexToAddress(cliContext.String(AccountAddressFlag.Name))
	callerAddress := gethcommon.HexToAddress(cliContext.String(CallerAddress.Name))
	ethRpcUrl := cliContext.String(flags.ETHRpcUrlFlag.Name)
	network := cliContext.String(flags.NetworkFlag.Name)
	environment := cliContext.String(flags.EnvironmentFlag.Name)
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
	}, nil
}

func acceptFlags() []cli.Flag {
	cmdFlags := []cli.Flag{
		&flags.VerboseFlag,
		&AccountAddressFlag,
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
