package admin

import (
	"context"
	"fmt"
	"sort"
	"strings"

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

type ListAdminsReader interface {
	ListAdmins(
		ctx context.Context,
		userAddress gethcommon.Address,
	) ([]gethcommon.Address, error)
}

func ListCmd(readerGenerator func(logging.Logger, *listAdminsConfig) (ListAdminsReader, error)) *cli.Command {
	listCmd := &cli.Command{
		Name:      "list-admins",
		Usage:     "user admin list-admins --account-address <AccountAddress>",
		UsageText: "List all users who are admins.",
		Description: `
		List all users who are admins.
		`,
		Action: func(c *cli.Context) error {
			return listAdmins(c, readerGenerator)
		},
		After: telemetry.AfterRunAction(),
		Flags: listAdminFlags(),
	}

	return listCmd
}

func listAdmins(
	cliCtx *cli.Context,
	generator func(logging.Logger, *listAdminsConfig) (ListAdminsReader, error),
) error {
	ctx := cliCtx.Context
	logger := common.GetLogger(cliCtx)

	config, err := readAndValidateListAdminsConfig(cliCtx, logger)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to read and validate user admin list config", err)
	}
	cliCtx.App.Metadata["network"] = config.ChainID.String()
	elReader, err := generator(logger, config)
	if err != nil {
		return err
	}

	pendingAdmins, err := elReader.ListAdmins(ctx, config.AccountAddress)
	if err != nil {
		return err
	}
	printAdmins(config.AccountAddress, pendingAdmins)
	return nil
}

func printAdmins(account gethcommon.Address, admins []gethcommon.Address) {
	fmt.Printf("Admins for AccountAddress: %s \n", account)
	fmt.Println(strings.Repeat("=", 60))
	for _, admin := range admins {
		fmt.Printf("%s \n", admin.String())
	}
}

func readAndValidateListAdminsConfig(cliContext *cli.Context, logger logging.Logger) (*listAdminsConfig, error) {
	accountAddress := gethcommon.HexToAddress(cliContext.String(AccountAddressFlag.Name))
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

	return &listAdminsConfig{
		Network:                  network,
		RPCUrl:                   ethRpcUrl,
		AccountAddress:           accountAddress,
		PermissionManagerAddress: gethcommon.HexToAddress(permissionManagerAddress),
		ChainID:                  chainID,
		Environment:              environment,
	}, nil
}

func generateListAdminsReader(logger logging.Logger, config *listAdminsConfig) (ListAdminsReader, error) {
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

func listAdminFlags() []cli.Flag {
	cmdFlags := []cli.Flag{
		&flags.VerboseFlag,
		&AccountAddressFlag,
		&PermissionControllerAddressFlag,
		&flags.NetworkFlag,
		&flags.EnvironmentFlag,
		&flags.ETHRpcUrlFlag,
	}
	sort.Sort(cli.FlagsByName(cmdFlags))
	return cmdFlags
}
