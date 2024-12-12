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

type RemovePendingAdminWriter interface {
	RemovePendingAdmin(
		ctx context.Context,
		request elcontracts.RemovePendingAdminRequest,
	) (*gethtypes.Receipt, error)
}

func RemovePendingCmd(
	generator func(logging.Logger, *removePendingAdminConfig) (RemovePendingAdminWriter, error),
) *cli.Command {
	removeCmd := &cli.Command{
		Name:      "remove-pending-admin",
		Usage:     "user admin remove-pending-admin --account-address <AccountAddress> --admin-address <AdminAddress>",
		UsageText: "Remove a user who is pending admin acceptance.",
		Description: `
		Remove a user who is pending admin acceptance.
		`,
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

	receipt, err := elWriter.RemovePendingAdmin(
		ctx,
		elcontracts.RemovePendingAdminRequest{
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

func readAndValidateRemovePendingAdminConfig(
	cliContext *cli.Context,
	logger logging.Logger,
) (*removePendingAdminConfig, error) {
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

	return &removePendingAdminConfig{
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

func generateRemovePendingAdminWriter(
	prompter utils.Prompter,
) func(logger logging.Logger, config *removePendingAdminConfig) (RemovePendingAdminWriter, error) {
	return func(logger logging.Logger, config *removePendingAdminConfig) (RemovePendingAdminWriter, error) {
		ethClient, err := ethclient.Dial(config.RPCUrl)
		if err != nil {
			return nil, eigenSdkUtils.WrapError("failed to create new eth client", err)
		}
		return common.GetELWriter(
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
	}
}

func removePendingAdminFlags() []cli.Flag {
	cmdFlags := []cli.Flag{
		&flags.VerboseFlag,
		&AccountAddressFlag,
		&AdminAddressFlag,
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
