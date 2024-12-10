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
	"sort"
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

func canCallCmd(readerGenerator func(logging.Logger, *canCallConfig) (UserCanCallReader, error)) *cli.Command {
	cmd := &cli.Command{
		Name:      "can-call",
		Usage:     "user appointee can-call --account-address <AccountsAddress> --caller-address <CallerAddress> --taget-address <TargetAddress> --selector <Selector>",
		UsageText: "Checks if a user has a specific permission.",
		Description: `
		Checks if a user has a specific permission.
		`,
		Action: func(c *cli.Context) error {
			return canCall(c, readerGenerator)
		},
		After: telemetry.AfterRunAction(),
		Flags: canCallFlags(),
	}

	return cmd
}

func canCall(cliCtx *cli.Context, generator func(logging.Logger, *canCallConfig) (UserCanCallReader, error)) error {
	ctx := cliCtx.Context
	logger := common.GetLogger(cliCtx)

	config, err := readAndValidateUserConfig(cliCtx, logger)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to read and validate user can call config", err)
	}
	cliCtx.App.Metadata["network"] = config.ChainID.String()
	elReader, err := generator(logger, config)
	if err != nil {
		return err
	}

	result, err := elReader.UserCanCall(ctx, config.UserAddress, config.CallerAddress, config.Target, config.Selector)
	fmt.Printf("CanCall Result: %v", result)
	fmt.Println()
	fmt.Printf("Selector, Target and User: %s, %x, %s", config.Target, string(config.Selector[:]), config.UserAddress)
	return err
}

func readAndValidateUserConfig(cliContext *cli.Context, logger logging.Logger) (*canCallConfig, error) {
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

	return &canCallConfig{
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

func generateUserCanCallReader(
	logger logging.Logger,
	config *canCallConfig,
) (UserCanCallReader, error) {
	ethClient, err := ethclient.Dial(config.RPCUrl)
	if err != nil {
		return nil, eigenSdkUtils.WrapError("failed to create new eth client", err)
	}
	elReader, err := elcontracts.NewReaderFromConfig(
		elcontracts.Config{
			PermissionsControllerAddress: config.PermissionManagerAddress,
		},
		ethClient,
		logger,
	)
	return elReader, err
}

func canCallFlags() []cli.Flag {
	cmdFlags := []cli.Flag{
		&AccountAddressFlag,
		&CallerAddressFlag,
		&TargetAddressFlag,
		&SelectorFlag,
		&PermissionControllerAddressFlag,
		&flags.NetworkFlag,
		&flags.EnvironmentFlag,
		&flags.ETHRpcUrlFlag,
	}
	sort.Sort(cli.FlagsByName(cmdFlags))
	return cmdFlags
}
