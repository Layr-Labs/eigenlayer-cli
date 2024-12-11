package admin

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

type IsAdminReader interface {
	IsAdmin(
		ctx context.Context,
		accountAddress gethcommon.Address,
		pendingAdminAddress gethcommon.Address,
	) (bool, error)
}

func IsAdminCmd(readerGenerator func(logging.Logger, *isAdminConfig) (IsAdminReader, error)) *cli.Command {
	cmd := &cli.Command{
		Name:      "is-admin",
		Usage:     "user admin is-admin --account-address <AccountAddress> --caller-address <CallerAddress>",
		UsageText: "Checks if a user is an admin.",
		Description: `
		Checks if a user is an admin.
		`,
		Action: func(c *cli.Context) error {
			return isAdmin(c, readerGenerator)
		},
		After: telemetry.AfterRunAction(),
		Flags: IsAdminFlags(),
	}

	return cmd
}

func isAdmin(cliCtx *cli.Context, generator func(logging.Logger, *isAdminConfig) (IsAdminReader, error)) error {
	ctx := cliCtx.Context
	logger := common.GetLogger(cliCtx)

	config, err := readAndValidateIsAdminConfig(cliCtx, logger)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to read and validate user admin is admin config", err)
	}
	cliCtx.App.Metadata["network"] = config.ChainID.String()
	elReader, err := generator(logger, config)
	if err != nil {
		return err
	}

	result, err := elReader.IsAdmin(ctx, config.AccountAddress, config.AdminAddress)
	if err != nil {
		return err
	}
	printIsAdminResult(result)
	return nil
}

func readAndValidateIsAdminConfig(cliContext *cli.Context, logger logging.Logger) (*isAdminConfig, error) {
	accountAddress := gethcommon.HexToAddress(cliContext.String(AccountAddressFlag.Name))
	adminAddress := gethcommon.HexToAddress(cliContext.String(AdminAddressFlag.Name))
	ethRpcUrl := cliContext.String(flags.ETHRpcUrlFlag.Name)
	network := cliContext.String(flags.NetworkFlag.Name)
	environment := cliContext.String(flags.EnvironmentFlag.Name)
	if environment == "" {
		environment = common.GetEnvFromNetwork(network)
	}

	chainID := utils.NetworkNameToChainId(network)
	permissionManagerAddress := cliContext.String(PermissionControllerAddressFlag.Name)

	var err error
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

	return &isAdminConfig{
		Network:                  network,
		RPCUrl:                   ethRpcUrl,
		AccountAddress:           accountAddress,
		AdminAddress:             adminAddress,
		PermissionManagerAddress: gethcommon.HexToAddress(permissionManagerAddress),
		ChainID:                  chainID,
		Environment:              environment,
	}, nil
}

func printIsAdminResult(result bool) {
	if result {
		fmt.Printf("Address provided is an admin.")
	} else {
		fmt.Printf("Address provided is not an admin.")
	}
}

func generateIsAdminReader(logger logging.Logger, config *isAdminConfig) (IsAdminReader, error) {
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

func IsAdminFlags() []cli.Flag {
	cmdFlags := []cli.Flag{
		&flags.VerboseFlag,
		&AccountAddressFlag,
		&CallerAddressFlag,
		&PermissionControllerAddressFlag,
		&flags.NetworkFlag,
		&flags.EnvironmentFlag,
		&flags.ETHRpcUrlFlag,
	}
	sort.Sort(cli.FlagsByName(cmdFlags))
	return cmdFlags
}
