package appointee

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

type RemoveAppointeePermissionWriter interface {
	RemovePermission(
		ctx context.Context,
		request elcontracts.RemovePermissionRequest,
	) (*gethtypes.Receipt, error)
}

func RemoveCmd(generator func(logging.Logger, *removeConfig) (RemoveAppointeePermissionWriter, error)) *cli.Command {
	removeCmd := &cli.Command{
		Name:      "remove",
		Usage:     "user appointee remove --account-address <AccountAddress> --appointee-address <AppointeeAddress> --target-address <TargetAddress> --selector <Selector>",
		UsageText: "Remove an appointee's permission",
		Description: `
		Remove an appointee's permission'.
		`,
		After: telemetry.AfterRunAction(),
		Action: func(c *cli.Context) error {
			return removeAppointeePermission(c, generator)
		},
		Flags: removeCommandFlags(),
	}

	return removeCmd
}

func removeAppointeePermission(
	cliCtx *cli.Context,
	generator func(logging.Logger, *removeConfig) (RemoveAppointeePermissionWriter, error),
) error {
	ctx := cliCtx.Context
	logger := common.GetLogger(cliCtx)

	config, err := readAndValidateRemoveConfig(cliCtx, logger)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to read and validate user appointee remove config", err)
	}
	cliCtx.App.Metadata["network"] = config.ChainID.String()
	permissionWriter, err := generator(logger, config)
	if err != nil {
		return err
	}
	receipt, err := permissionWriter.RemovePermission(
		ctx,
		elcontracts.RemovePermissionRequest{
			AccountAddress:   config.AccountAddress,
			AppointeeAddress: config.AppointeeAddress,
			Target:           config.Target,
			Selector:         config.Selector,
			WaitForReceipt:   true,
		},
	)
	if err != nil {
		return err
	}
	common.PrintTransactionInfo(receipt.TxHash.String(), config.ChainID)
	return nil
}

func generateRemoveAppointeePermissionWriter(
	prompter utils.Prompter,
) func(
	logger logging.Logger,
	config *removeConfig,
) (RemoveAppointeePermissionWriter, error) {
	return func(logger logging.Logger, config *removeConfig) (RemoveAppointeePermissionWriter, error) {
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

func readAndValidateRemoveConfig(cliContext *cli.Context, logger logging.Logger) (*removeConfig, error) {
	accountAddress := gethcommon.HexToAddress(cliContext.String(AccountAddressFlag.Name))
	appointeeAddress := gethcommon.HexToAddress(cliContext.String(AppointeeAddressFlag.Name))
	ethRpcUrl := cliContext.String(flags.ETHRpcUrlFlag.Name)
	network := cliContext.String(flags.NetworkFlag.Name)
	environment := cliContext.String(flags.EnvironmentFlag.Name)
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

	return &removeConfig{
		Network:                  network,
		RPCUrl:                   ethRpcUrl,
		AccountAddress:           accountAddress,
		AppointeeAddress:         appointeeAddress,
		Target:                   target,
		Selector:                 selectorBytes,
		SignerConfig:             *signerConfig,
		PermissionManagerAddress: gethcommon.HexToAddress(permissionManagerAddress),
		ChainID:                  chainID,
		Environment:              environment,
	}, nil
}

func removeCommandFlags() []cli.Flag {
	cmdFlags := []cli.Flag{
		&flags.VerboseFlag,
		&AccountAddressFlag,
		&AppointeeAddressFlag,
		&TargetAddressFlag,
		&SelectorFlag,
		&PermissionControllerAddressFlag,
		&flags.NetworkFlag,
		&flags.EnvironmentFlag,
		&flags.ETHRpcUrlFlag,
		&flags.BroadcastFlag,
	}
	cmdFlags = append(cmdFlags, flags.GetSignerFlags()...)
	sort.Sort(cli.FlagsByName(cmdFlags))
	return cmdFlags
}
