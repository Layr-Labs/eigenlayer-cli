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

type IsPendingAdminReader interface {
	IsPendingAdmin(
		ctx context.Context,
		accountAddress gethcommon.Address,
		pendingAdminAddress gethcommon.Address,
	) (bool, error)
}

func IsPendingCmd(
	readerGenerator func(logging.Logger, *isPendingAdminConfig) (IsPendingAdminReader, error),
) *cli.Command {
	isPendingCmd := &cli.Command{
		Name:      "is-pending-admin",
		Usage:     "user admin is-pending-admin --account-address <AccountAddress> --pending-admin-address <PendingAdminAddress>",
		UsageText: "Checks if a user is pending acceptance to admin.",
		Description: `
		Checks if a user is pending acceptance to admin.
		`,
		Action: func(c *cli.Context) error {
			return isPendingAdmin(c, readerGenerator)
		},
		After: telemetry.AfterRunAction(),
		Flags: isPendingAdminFlags(),
	}

	return isPendingCmd
}

func isPendingAdmin(
	cliCtx *cli.Context,
	generator func(logging.Logger, *isPendingAdminConfig) (IsPendingAdminReader, error),
) error {
	ctx := cliCtx.Context
	logger := common.GetLogger(cliCtx)

	config, err := readAndValidateIsPendingAdminConfig(cliCtx, logger)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to read and validate user admin is pending config", err)
	}
	cliCtx.App.Metadata["network"] = config.ChainID.String()
	elReader, err := generator(logger, config)
	if err != nil {
		return err
	}

	result, err := elReader.IsPendingAdmin(ctx, config.AccountAddress, config.PendingAdminAddress)
	if err != nil {
		return err
	}
	printIsPendingAdminResult(result)
	return nil
}

func readAndValidateIsPendingAdminConfig(
	cliContext *cli.Context,
	logger logging.Logger,
) (*isPendingAdminConfig, error) {
	accountAddress := gethcommon.HexToAddress(cliContext.String(AccountAddressFlag.Name))
	pendingAdminAddress := gethcommon.HexToAddress(cliContext.String(PendingAdminAddressFlag.Name))
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

	return &isPendingAdminConfig{
		Network:                  network,
		RPCUrl:                   ethRpcUrl,
		AccountAddress:           accountAddress,
		PendingAdminAddress:      pendingAdminAddress,
		PermissionManagerAddress: gethcommon.HexToAddress(permissionManagerAddress),
		ChainID:                  chainID,
		Environment:              environment,
	}, nil
}

func printIsPendingAdminResult(result bool) {
	if result {
		fmt.Printf("Address provided is a pending admin.")
	} else {
		fmt.Printf("Address provided is not a pending admin.")
	}
}

func generateIsPendingAdminReader(logger logging.Logger, config *isPendingAdminConfig) (IsPendingAdminReader, error) {
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

func isPendingAdminFlags() []cli.Flag {
	cmdFlags := []cli.Flag{
		&flags.VerboseFlag,
		&AccountAddressFlag,
		&PendingAdminAddressFlag,
		&flags.OutputTypeFlag,
		&flags.OutputFileFlag,
		&PermissionControllerAddressFlag,
		&flags.NetworkFlag,
		&flags.EnvironmentFlag,
		&flags.ETHRpcUrlFlag,
	}
	sort.Sort(cli.FlagsByName(cmdFlags))
	return cmdFlags
}
