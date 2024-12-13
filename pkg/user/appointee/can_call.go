package appointee

import (
	"context"
	"fmt"
	"sort"

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

type CanCallReader interface {
	CanCall(
		ctx context.Context,
		accountAddress gethcommon.Address,
		appointeeAddress gethcommon.Address,
		target gethcommon.Address,
		selector [4]byte,
	) (bool, error)
}

func canCallCmd(readerGenerator func(logging.Logger, *canCallConfig) (CanCallReader, error)) *cli.Command {
	cmd := &cli.Command{
		Name:      "can-call",
		Usage:     "user appointee can-call --account-address <AccountsAddress> --appointee-address <AppointeeAddress> --target-address <TargetAddress> --selector <Selector>",
		UsageText: "Checks if an appointee has a specific permission.",
		Description: `
		Checks if an appointee has a specific permission.
		`,
		Action: func(c *cli.Context) error {
			return canCall(c, readerGenerator)
		},
		After: telemetry.AfterRunAction(),
		Flags: canCallFlags(),
	}

	return cmd
}

func canCall(cliCtx *cli.Context, generator func(logging.Logger, *canCallConfig) (CanCallReader, error)) error {
	ctx := cliCtx.Context
	logger := common.GetLogger(cliCtx)

	config, err := readAndValidateCanCallConfig(cliCtx, logger)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to read and validate user can call config", err)
	}
	cliCtx.App.Metadata["network"] = config.ChainID.String()
	elReader, err := generator(logger, config)
	if err != nil {
		return err
	}

	result, err := elReader.CanCall(ctx, config.AccountAddress, config.AppointeeAddress, config.Target, config.Selector)
	fmt.Printf("CanCall Result: %v\n", result)
	fmt.Printf(
		"Target, Selector and Appointee: %s, %x, %s\n",
		config.Target,
		string(config.Selector[:]),
		config.AppointeeAddress,
	)
	return err
}

func readAndValidateCanCallConfig(cliContext *cli.Context, logger logging.Logger) (*canCallConfig, error) {
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
		AccountAddress:           accountAddress,
		AppointeeAddress:         appointeeAddress,
		Target:                   target,
		Selector:                 selectorBytes,
		PermissionManagerAddress: gethcommon.HexToAddress(permissionManagerAddress),
		ChainID:                  chainID,
		Environment:              environment,
	}, nil
}

func generateCanCallReader(
	logger logging.Logger,
	config *canCallConfig,
) (CanCallReader, error) {
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
		&AppointeeAddressFlag,
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
