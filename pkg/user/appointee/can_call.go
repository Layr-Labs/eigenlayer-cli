package appointee

import (
	"context"
	"fmt"
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
	"unicode/utf8"
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
		UsageText: "Check if a user can call a contract function.",
		Description: `
		The can-call command allows you to check if a user has a specific permission.
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
			&flags.NetworkFlag,
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
	env := cliContext.String(flags.EnvironmentFlag.Name)
	target := gethcommon.HexToAddress(cliContext.String(TargetAddressFlag.Name))
	selector := cliContext.String(SelectorFlag.Name)
	selectorBytes, err := validateAndConvertSelectorString(selector)

	if err != nil {
		return nil, err
	}

	if env == "" {
		env = getEnvFromNetwork(network)
	}

	logger.Debugf("Network: %s, Env: %s", network, env)
	permissionManagerAddress := cliContext.String(PermissionManagerAddressFlag.Name)

	if common.IsEmptyString(permissionManagerAddress) {
		permissionManagerAddress, err = common.GetPermissionManagerAddress(utils.NetworkNameToChainId(network))
		if err != nil {
			return nil, err
		}
	}
	logger.Debugf("Using PermissionsManager address: %s", permissionManagerAddress)

	chainID := utils.NetworkNameToChainId(network)
	logger.Debugf("Using chain ID: %s", chainID.String())
	logger.Debugf("Using network: %s", network)

	return &CanCallConfig{
		Network:                  network,
		RPCUrl:                   ethRpcUrl,
		UserAddress:              userAddress,
		CallerAddress:            callerAddress,
		Target:                   target,
		Selector:                 selectorBytes,
		PermissionManagerAddress: gethcommon.HexToAddress(permissionManagerAddress),
		ChainID:                  chainID,
		Environment:              env,
	}, nil
}

func validateAndConvertSelectorString(selector string) ([4]byte, error) {
	if utf8.RuneCountInString(selector) != 4 {
		return [4]byte{}, fmt.Errorf("selector must be 4 characters long")
	}
	var selectorBytes [4]byte
	copy(selectorBytes[:], selector)
	return selectorBytes, nil
}

func getEigenLayerReader(cliContext *cli.Context, logger logging.Logger, config *CanCallConfig) (UserCanCallReader, error) {
	if reader, ok := cliContext.App.Metadata["elReader"].(UserCanCallReader); ok {
		return reader, nil
	}
	return createDefaultEigenLayerReader(cliContext, config, logger)
}

func createDefaultEigenLayerReader(cliContext *cli.Context, config *CanCallConfig, logger logging.Logger) (UserCanCallReader, error) {
	cliContext.App.Metadata["network"] = config.ChainID.String()

	ethClient, err := ethclient.Dial(config.RPCUrl)
	if err != nil {
		return nil, eigenSdkUtils.WrapError("failed to create new eth client", err)
	}

	elReader, err := elcontracts.NewReaderFromConfig(
		elcontracts.Config{
			PermissionManagerAddress: config.PermissionManagerAddress,
		},
		ethClient,
		logger,
	)
	return elReader, err
}

func getEnvFromNetwork(network string) string {
	switch network {
	case utils.HoleskyNetworkName:
		return "testnet"
	case utils.MainnetNetworkName:
		return "mainnet"
	default:
		return "local"
	}
}
