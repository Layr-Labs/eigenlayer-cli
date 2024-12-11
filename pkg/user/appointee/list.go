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

type ListUsersReader interface {
	ListUsers(
		ctx context.Context,
		userAddress gethcommon.Address,
		target gethcommon.Address,
		selector [4]byte,
	) ([]gethcommon.Address, error)
}

func ListCmd(readerGenerator func(logging.Logger, *listUsersConfig) (ListUsersReader, error)) *cli.Command {
	listCmd := &cli.Command{
		Name:      "list",
		Usage:     "user appointee list --account-address <AccountAddress> --target-address <TargetAddress> --selector <Selector>",
		UsageText: "Lists all appointed users for an account with the provided permissions.",
		Description: `
		Lists all appointed users for an account with the provided permissions.
		`,
		Action: func(c *cli.Context) error {
			return listAppointees(c, readerGenerator)
		},
		After: telemetry.AfterRunAction(),
		Flags: listFlags(),
	}

	return listCmd
}

func listAppointees(
	cliCtx *cli.Context,
	generator func(logging.Logger, *listUsersConfig) (ListUsersReader, error),
) error {
	ctx := cliCtx.Context
	logger := common.GetLogger(cliCtx)

	config, err := readAndValidateListUsersConfig(cliCtx, logger)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to read and validate user can call config", err)
	}

	elReader, err := generator(logger, config)
	if err != nil {
		return err
	}

	users, err := elReader.ListUsers(ctx, config.UserAddress, config.Target, config.Selector)
	if err != nil {
		return err
	}
	printResults(config, users)
	return nil
}

func printResults(config *listUsersConfig, users []gethcommon.Address) {
	fmt.Printf(
		"Selector, Target and Appointer: %s, %x, %s",
		config.Target,
		string(config.Selector[:]),
		config.UserAddress,
	)
	fmt.Println(strings.Repeat("=", 80))

	for _, user := range users {
		fmt.Printf("User Id: 0x%b\n", user)
	}
}

func readAndValidateListUsersConfig(cliContext *cli.Context, logger logging.Logger) (*listUsersConfig, error) {
	userAddress := gethcommon.HexToAddress(cliContext.String(AccountAddressFlag.Name))
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
	cliContext.App.Metadata["network"] = chainID.String()
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

	return &listUsersConfig{
		Network:                  network,
		RPCUrl:                   ethRpcUrl,
		UserAddress:              userAddress,
		Target:                   target,
		Selector:                 selectorBytes,
		PermissionManagerAddress: gethcommon.HexToAddress(permissionManagerAddress),
		ChainID:                  chainID,
		Environment:              environment,
	}, nil
}

func generateListUsersReader(logger logging.Logger, config *listUsersConfig) (ListUsersReader, error) {
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

func listFlags() []cli.Flag {
	cmdFlags := []cli.Flag{
		&flags.VerboseFlag,
		&AccountAddressFlag,
		&TargetAddressFlag,
		&SelectorFlag,
		&flags.NetworkFlag,
		&flags.EnvironmentFlag,
		&flags.ETHRpcUrlFlag,
	}
	sort.Sort(cli.FlagsByName(cmdFlags))
	return cmdFlags
}
