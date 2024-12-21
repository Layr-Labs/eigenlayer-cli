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

type ListPendingAdminsReader interface {
	ListPendingAdmins(
		ctx context.Context,
		userAddress gethcommon.Address,
	) ([]gethcommon.Address, error)
}

func ListPendingCmd(
	readerGenerator func(logging.Logger, *listPendingAdminsConfig) (ListPendingAdminsReader, error),
) *cli.Command {
	listPendingCmd := &cli.Command{
		Name:  "list-pending-admins",
		Usage: "List all users who are pending admin acceptance.",
		Action: func(c *cli.Context) error {
			return listPendingAdmins(c, readerGenerator)
		},
		After: telemetry.AfterRunAction(),
		Flags: listPendingAdminsFlags(),
	}

	return listPendingCmd
}

func listPendingAdmins(
	cliCtx *cli.Context,
	generator func(logging.Logger, *listPendingAdminsConfig) (ListPendingAdminsReader, error),
) error {
	ctx := cliCtx.Context
	logger := common.GetLogger(cliCtx)

	config, err := readAndValidateListPendingAdminsConfig(cliCtx, logger)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to read and validate user admin list pending config", err)
	}
	cliCtx.App.Metadata["network"] = config.ChainID.String()
	elReader, err := generator(logger, config)
	if err != nil {
		return err
	}

	pendingAdmins, err := elReader.ListPendingAdmins(ctx, config.AccountAddress)
	if err != nil {
		return err
	}
	printPendingAdmins(config.AccountAddress, pendingAdmins)
	return nil
}

func printPendingAdmins(account gethcommon.Address, admins []gethcommon.Address) {
	fmt.Printf("Pending Admins for AccountAddress: %s \n", account)
	fmt.Println(strings.Repeat("=", 60))
	for _, admin := range admins {
		fmt.Printf("%s \n", admin.String())
	}
}

func readAndValidateListPendingAdminsConfig(
	cliContext *cli.Context,
	logger logging.Logger,
) (*listPendingAdminsConfig, error) {
	accountAddress := gethcommon.HexToAddress(cliContext.String(AccountAddressFlag.Name))
	ethRpcUrl := cliContext.String(flags.ETHRpcUrlFlag.Name)
	network := cliContext.String(flags.NetworkFlag.Name)
	environment := cliContext.String(flags.EnvironmentFlag.Name)
	if environment == "" {
		environment = common.GetEnvFromNetwork(network)
	}

	chainID := utils.NetworkNameToChainId(network)
	permissionControllerAddress := cliContext.String(PermissionControllerAddressFlag.Name)

	var err error
	if common.IsEmptyString(permissionControllerAddress) {
		permissionControllerAddress, err = common.GetPermissionControllerAddress(utils.NetworkNameToChainId(network))
		if err != nil {
			return nil, err
		}
	}

	logger.Debugf(
		"Env: %s, network: %s, chain ID: %s, PermissionController address: %s",
		environment,
		network,
		chainID,
		permissionControllerAddress,
	)

	return &listPendingAdminsConfig{
		Network:                     network,
		RPCUrl:                      ethRpcUrl,
		AccountAddress:              accountAddress,
		PermissionControllerAddress: gethcommon.HexToAddress(permissionControllerAddress),
		ChainID:                     chainID,
		Environment:                 environment,
	}, nil
}

func generateListPendingAdminsReader(
	logger logging.Logger,
	config *listPendingAdminsConfig,
) (ListPendingAdminsReader, error) {
	ethClient, err := ethclient.Dial(config.RPCUrl)
	if err != nil {
		return nil, eigenSdkUtils.WrapError("failed to create new eth client", err)
	}
	elReader, err := elcontracts.NewReaderFromConfig(
		elcontracts.Config{
			PermissionsControllerAddress: config.PermissionControllerAddress,
		},
		ethClient,
		logger,
	)
	return elReader, err
}

func listPendingAdminsFlags() []cli.Flag {
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
