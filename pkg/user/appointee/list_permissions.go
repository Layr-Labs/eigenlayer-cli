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
)

type PermissionsReader interface {
	ListUserPermissions(
		ctx context.Context,
		appointed gethcommon.Address,
		userAddress gethcommon.Address,
	) ([]gethcommon.Address, [][4]byte, error)
}

func ListPermissionsCmd(readerGenerator func(logging.Logger, *listUserPermissionsConfig) (PermissionsReader, error)) *cli.Command {
	cmd := &cli.Command{
		Name:      "list-permissions",
		Usage:     "user appointee list-permissions --account-address <AccountAddress> --appointee-address <AppointeeAddress>",
		UsageText: "List all permissions for a user.",
		Description: `
		List all permissions of a user.
		`,
		Action: func(c *cli.Context) error {
			return listPermissions(c, readerGenerator)
		},
		After: telemetry.AfterRunAction(),
		Flags: []cli.Flag{
			&flags.VerboseFlag,
			&AccountAddressFlag,
			&AppointeeAddressFlag,
			&flags.NetworkFlag,
			&flags.EnvironmentFlag,
			&flags.ETHRpcUrlFlag,
		},
	}

	return cmd
}

func listPermissions(cliCtx *cli.Context, generator func(logging.Logger, *listUserPermissionsConfig) (PermissionsReader, error)) error {
	ctx := cliCtx.Context
	logger := common.GetLogger(cliCtx)

	config, err := readAndValidateListUserPermissionsConfig(cliCtx, logger)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to read and validate list user permissions config", err)
	}

	reader, err := generator(logger, config)
	if err != nil {
		return err
	}

	users, permissions, err := reader.ListUserPermissions(ctx, config.AccountAddress, config.UserAddress)
	if err != nil {
		return err
	}
	printPermissions(config, users, permissions)
	return nil
}

func readAndValidateListUserPermissionsConfig(cliContext *cli.Context, logger logging.Logger) (*listUserPermissionsConfig, error) {
	accountAddress := gethcommon.HexToAddress(cliContext.String(AccountAddressFlag.Name))
	userAddress := gethcommon.HexToAddress(cliContext.String(AppointeeAddressFlag.Name))
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

	return &listUserPermissionsConfig{
		Network:                  network,
		RPCUrl:                   ethRpcUrl,
		AccountAddress:           accountAddress,
		UserAddress:              userAddress,
		PermissionManagerAddress: gethcommon.HexToAddress(permissionManagerAddress),
		ChainID:                  chainID,
		Environment:              environment,
	}, nil
}

func printPermissions(config *listUserPermissionsConfig, targets []gethcommon.Address, selectors [][4]byte) {
	fmt.Printf("User: %s\n", config.UserAddress)
	fmt.Printf("Appointed by: %s\n", config.AccountAddress)
	fmt.Println("====================================================================================")

	for _, target := range targets {
		for _, selector := range selectors {
			fmt.Printf("Target: %s, Selector: %x\n", target, selector)
		}

	}
}

func generateListUserPermissionsReader(
	logger logging.Logger,
	config *listUserPermissionsConfig,
) (PermissionsReader, error) {
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
