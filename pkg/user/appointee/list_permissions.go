package appointee

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

type PermissionsReader interface {
	ListAppointeePermissions(
		ctx context.Context,
		accountAddress gethcommon.Address,
		appointeeAddress gethcommon.Address,
	) ([]gethcommon.Address, [][4]byte, error)
}

func ListPermissionsCmd(
	readerGenerator func(logging.Logger, *listAppointeePermissionsConfig) (PermissionsReader, error),
) *cli.Command {
	cmd := &cli.Command{
		Name:  "list-permissions",
		Usage: "List all permissions of a user.",
		Action: func(c *cli.Context) error {
			return listPermissions(c, readerGenerator)
		},
		After: telemetry.AfterRunAction(),
		Flags: listPermissionFlags(),
	}

	return cmd
}

func listPermissions(
	cliCtx *cli.Context,
	generator func(logging.Logger, *listAppointeePermissionsConfig) (PermissionsReader, error),
) error {
	ctx := cliCtx.Context
	logger := common.GetLogger(cliCtx)

	config, err := readAndValidateListAppointeePermissionsConfig(cliCtx, logger)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to read and validate list user permissions config", err)
	}

	cliCtx.App.Metadata["network"] = config.ChainID.String()
	reader, err := generator(logger, config)
	if err != nil {
		return err
	}

	appointees, permissions, err := reader.ListAppointeePermissions(ctx, config.AccountAddress, config.AppointeeAddress)
	if err != nil {
		return err
	}
	printPermissions(config, appointees, permissions)
	return nil
}

func readAndValidateListAppointeePermissionsConfig(
	cliContext *cli.Context,
	logger logging.Logger,
) (*listAppointeePermissionsConfig, error) {
	accountAddress := gethcommon.HexToAddress(cliContext.String(AccountAddressFlag.Name))
	appointeeAddress := gethcommon.HexToAddress(cliContext.String(AppointeeAddressFlag.Name))
	ethRpcUrl := cliContext.String(flags.ETHRpcUrlFlag.Name)
	network := cliContext.String(flags.NetworkFlag.Name)
	environment := cliContext.String(flags.EnvironmentFlag.Name)

	if environment == "" {
		environment = common.GetEnvFromNetwork(network)
	}

	chainID := utils.NetworkNameToChainId(network)
	PermissionControllerAddress := cliContext.String(PermissionControllerAddressFlag.Name)

	var err error
	if common.IsEmptyString(PermissionControllerAddress) {
		PermissionControllerAddress, err = common.GetPermissionControllerAddress(utils.NetworkNameToChainId(network))
		if err != nil {
			return nil, err
		}
	}

	logger.Debugf(
		"Env: %s, network: %s, chain ID: %s, PermissionController address: %s",
		environment,
		network,
		chainID,
		PermissionControllerAddress,
	)

	return &listAppointeePermissionsConfig{
		Network:                     network,
		RPCUrl:                      ethRpcUrl,
		AccountAddress:              accountAddress,
		AppointeeAddress:            appointeeAddress,
		PermissionControllerAddress: gethcommon.HexToAddress(PermissionControllerAddress),
		ChainID:                     chainID,
		Environment:                 environment,
	}, nil
}

func printPermissions(config *listAppointeePermissionsConfig, targets []gethcommon.Address, selectors [][4]byte) {
	fmt.Printf("Appointee address: %s\n", config.AppointeeAddress)
	fmt.Printf("Appointed by: %s\n", config.AccountAddress)
	fmt.Println(strings.Repeat("=", 60))

	for index := range targets {
		fmt.Printf("Target: %s, Selector: %x\n", targets[index], selectors[index])
	}
}

func generateListAppointeePermissionsReader(
	logger logging.Logger,
	config *listAppointeePermissionsConfig,
) (PermissionsReader, error) {
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

func listPermissionFlags() []cli.Flag {
	cmdFlags := []cli.Flag{
		&flags.VerboseFlag,
		&AccountAddressFlag,
		&AppointeeAddressFlag,
		&PermissionControllerAddressFlag,
		&flags.NetworkFlag,
		&flags.EnvironmentFlag,
		&flags.ETHRpcUrlFlag,
	}
	sort.Sort(cli.FlagsByName(cmdFlags))
	return cmdFlags
}
