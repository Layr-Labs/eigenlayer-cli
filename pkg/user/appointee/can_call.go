package appointee

import (
	"context"
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

type UserCanCallReader interface {
	UserCanCall(
		ctx context.Context,
		userAddress gethcommon.Address,
		callerAddress gethcommon.Address,
		target gethcommon.Address,
		selector [4]byte,
	) (bool, error)
}

func CanCallCmd() *cli.Command {
	canCallCmd := &cli.Command{
		Name:      "can-call",
		Usage:     "user appointee can-call <AccountsAddress> <CallerAddress> <TargetAddress> <Selector>",
		UsageText: "Checks if a user has a specific permission.",
		Description: `
		Checks if a user has a specific permission.
		`,
		Action: func(c *cli.Context) error {
			return canCall(c)
		},
		After: telemetry.AfterRunAction(),
		Flags: []cli.Flag{
			&AccountAddressFlag,
			&CallerAddressFlag,
			&TargetAddressFlag,
			&SelectorFlag,
			&PermissionControllerAddressFlag,
			&flags.NetworkFlag,
			&flags.EnvironmentFlag,
			&flags.ETHRpcUrlFlag,
		},
	}

	return canCallCmd
}

func canCall(cliCtx *cli.Context) error {
	ctx := cliCtx.Context
	logger := common.GetLogger(cliCtx)

	config, err := readAndValidateUserConfig(cliCtx, logger)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to read and validate user can call config", err)
	}

	elReader, err := getEigenLayerReader(cliCtx, logger, config)
	if err != nil {
		return err
	}

	_, err = elReader.UserCanCall(ctx, config.UserAddress, config.CallerAddress, config.Target, config.Selector)
	return err
}

func readAndValidateUserConfig(cliContext *cli.Context, logger logging.Logger) (*CanCallConfig, error) {
	userAddress := gethcommon.HexToAddress(cliContext.String(AccountAddressFlag.Name))
	callerAddress := gethcommon.HexToAddress(cliContext.String(CallerAddressFlag.Name))
	ethRpcUrl := cliContext.String(flags.ETHRpcUrlFlag.Name)
	network := cliContext.String(flags.NetworkFlag.Name)
	environment := cliContext.String(flags.EnvironmentFlag.Name)
	target := gethcommon.HexToAddress(cliContext.String(TargetAddressFlag.Name))
	selector := cliContext.String(SelectorFlag.Name)
	selectorBytes, err := common.ValidateAndConvertSelectorString(selector)
	if err != nil {
		return nil, err
	}

	if environment == "" {
		environment = common.GetEnvFromNetwork(network)
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

	return &CanCallConfig{
		Network:                  network,
		RPCUrl:                   ethRpcUrl,
		UserAddress:              userAddress,
		CallerAddress:            callerAddress,
		Target:                   target,
		Selector:                 selectorBytes,
		PermissionManagerAddress: gethcommon.HexToAddress(permissionManagerAddress),
		ChainID:                  chainID,
		Environment:              environment,
	}, nil
}

func getEigenLayerReader(
	cliContext *cli.Context,
	logger logging.Logger,
	config *CanCallConfig,
) (UserCanCallReader, error) {
	if reader, ok := cliContext.App.Metadata["elReader"].(UserCanCallReader); ok {
		return reader, nil
	}
	return createDefaultEigenLayerReader(cliContext, config, logger)
}

func createDefaultEigenLayerReader(
	cliContext *cli.Context,
	config *CanCallConfig,
	logger logging.Logger,
) (UserCanCallReader, error) {
	cliContext.App.Metadata["network"] = config.ChainID.String()

	ethClient, err := ethclient.Dial(config.RPCUrl)
	if err != nil {
		return nil, eigenSdkUtils.WrapError("failed to create new eth client", err)
	}
	logger.Infof("PERMISSION_ADDRESS: %s", config.PermissionManagerAddress)
	elReader, err := elcontracts.NewReaderFromConfig(
		elcontracts.Config{
			PermissionsControllerAddress: config.PermissionManagerAddress,
		},
		ethClient,
		logger,
	)
	return elReader, err
}
