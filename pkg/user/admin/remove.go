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

type RemoveAdminWriter interface {
	RemoveAdmin(
		ctx context.Context,
		request elcontracts.RemoveAdminRequest,
	) (*gethtypes.Receipt, error)
}

func RemoveCmd(generator func(logging.Logger, *removeAdminConfig) (RemoveAdminWriter, error)) *cli.Command {
	removeCmd := &cli.Command{
		Name:      "remove-admin",
		Usage:     "user admin remove-admin --account-address <AccountAddress> --admin-address <AdminAddress>",
		UsageText: "The remove command allows you to remove an admin user.",
		Description: `
		The remove command allows you to remove an admin user.
		`,
		Action: func(context *cli.Context) error {
			return removeAdmin(context, generator)
		},
		After: telemetry.AfterRunAction(),
		Flags: removeFlags(),
	}

	return removeCmd
}

func removeAdmin(
	cliCtx *cli.Context,
	generator func(logging.Logger, *removeAdminConfig) (RemoveAdminWriter, error),
) error {
	ctx := cliCtx.Context
	logger := common.GetLogger(cliCtx)

	config, err := readAndValidateRemoveAdminConfig(cliCtx, logger)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to read and validate user can call config", err)
	}
	cliCtx.App.Metadata["network"] = config.ChainID.String()
	elWriter, err := generator(logger, config)
	if err != nil {
		return err
	}

	receipt, err := elWriter.RemoveAdmin(
		ctx,
		elcontracts.RemoveAdminRequest{
			AccountAddress: config.AccountAddress,
			AdminAddress:   config.AdminAddress,
			WaitForReceipt: true,
		},
	)
	if err != nil {
		return err
	}
	common.PrintTransactionInfo(receipt.TxHash.String(), config.ChainID)
	return nil
}

func readAndValidateRemoveAdminConfig(
	cliContext *cli.Context,
	logger logging.Logger,
) (*removeAdminConfig, error) {
	accountAddress := gethcommon.HexToAddress(cliContext.String(AccountAddressFlag.Name))
	adminAddress := gethcommon.HexToAddress(cliContext.String(AdminAddressFlag.Name))
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

	logger.Debugf(
		"Env: %s, network: %s, chain ID: %s, PermissionManager address: %s",
		environment,
		network,
		chainID,
		permissionManagerAddress,
	)

	return &removeAdminConfig{
		Network:                  network,
		RPCUrl:                   ethRpcUrl,
		AccountAddress:           accountAddress,
		AdminAddress:             adminAddress,
		PermissionManagerAddress: gethcommon.HexToAddress(permissionManagerAddress),
		SignerConfig:             *signerConfig,
		ChainID:                  chainID,
		Environment:              environment,
	}, nil
}

func generateRemoveAdminWriter(
	prompter utils.Prompter,
) func(logger logging.Logger, config *removeAdminConfig) (RemoveAdminWriter, error) {
	return func(logger logging.Logger, config *removeAdminConfig) (RemoveAdminWriter, error) {
		ethClient, err := ethclient.Dial(config.RPCUrl)
		if err != nil {
			return nil, eigenSdkUtils.WrapError("failed to create new eth client", err)
		}
		elWriter, err := common.GetELWriter(
			config.AccountAddress,
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

func removeFlags() []cli.Flag {
	cmdFlags := []cli.Flag{
		&flags.VerboseFlag,
		&AccountAddressFlag,
		&AdminAddressFlag,
		&flags.OutputTypeFlag,
		&flags.OutputFileFlag,
		&flags.NetworkFlag,
		&flags.EnvironmentFlag,
		&flags.ETHRpcUrlFlag,
		&PermissionControllerAddressFlag,
	}
	sort.Sort(cli.FlagsByName(cmdFlags))
	return append(cmdFlags, flags.GetSignerFlags()...)
}
