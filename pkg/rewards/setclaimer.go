package rewards

import (
	"context"
	"fmt"
	"sort"

	eigenSdkUtils "github.com/Layr-Labs/eigensdk-go/utils"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	"github.com/Layr-Labs/eigensdk-go/logging"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/urfave/cli/v2"
)

func SetClaimerCmd(p utils.Prompter) *cli.Command {
	setClaimerCmd := &cli.Command{
		Name:      "set-claimer",
		Usage:     "Set the claimer address for the earner",
		UsageText: "set-claimer",
		Description: `
Set the rewards claimer address for the earner.
		`,
		After: telemetry.AfterRunAction(),
		Flags: getSetClaimerFlags(),
		Action: func(cCtx *cli.Context) error {
			return SetClaimer(cCtx, p)
		},
	}

	return setClaimerCmd
}

func getSetClaimerFlags() []cli.Flag {
	baseFlags := []cli.Flag{
		&flags.NetworkFlag,
		&flags.ETHRpcUrlFlag,
		&flags.OutputFileFlag,
		&flags.OutputTypeFlag,
		&flags.BroadcastFlag,
		&EarnerAddressFlag,
		&RewardsCoordinatorAddressFlag,
		&ClaimerAddressFlag,
		&flags.VerboseFlag,
	}

	allFlags := append(baseFlags, flags.GetSignerFlags()...)
	sort.Sort(cli.FlagsByName(allFlags))
	return allFlags
}

func SetClaimer(cCtx *cli.Context, p utils.Prompter) error {
	logger := common.GetLogger(cCtx)
	config, err := readAndValidateSetClaimerConfig(cCtx, logger)
	if err != nil {
		return err
	}

	cCtx.App.Metadata["network"] = config.ChainID.String()

	ethClient, err := ethclient.Dial(config.RPCUrl)
	if err != nil {
		return err
	}

	if !config.Broadcast {
		_, _, contractBindings, err := elcontracts.BuildClients(elcontracts.Config{
			RewardsCoordinatorAddress: config.RewardsCoordinatorAddress,
		}, ethClient, nil, logger, nil)
		if err != nil {
			return err
		}

		noSendTxOpts := common.GetNoSendTxOpts(config.EarnerAddress)
		unsignedTx, err := contractBindings.RewardsCoordinator.SetClaimerFor(noSendTxOpts, config.ClaimerAddress)

		if config.OutputType == string(common.OutputType_Calldata) {
			if err != nil {
				return err
			}
			calldataHex := gethcommon.Bytes2Hex(unsignedTx.Data())
			if !common.IsEmptyString(config.Output) {
				err := common.WriteToFile([]byte(calldataHex), config.Output)
				if err != nil {
					return err
				}
			} else {
				fmt.Println(calldataHex)
			}
		} else if config.OutputType == string(common.OutputType_Pretty) {
			if !common.IsEmptyString(config.Output) {
				fmt.Println("output file not supported for pretty output type")
				fmt.Println()
			}
			fmt.Printf(
				"Claimer address %s will be set for earner %s\n",
				config.ClaimerAddress.String(),
				config.EarnerAddress.String(),
			)
		} else {
			return fmt.Errorf("unsupported output type for this command %s", config.Output)
		}
		txFeeDetails := common.GetTxFeeDetails(unsignedTx)
		fmt.Println()
		txFeeDetails.Print()
		fmt.Println("To broadcast the claim, use the --broadcast flag")

		return nil
	}

	elWriter, err := common.GetELWriter(
		config.EarnerAddress,
		config.SignerConfig,
		ethClient,
		elcontracts.Config{
			RewardsCoordinatorAddress: config.RewardsCoordinatorAddress,
		},
		p,
		config.ChainID,
		logger,
	)

	if err != nil {
		return eigenSdkUtils.WrapError("failed to get EL writer", err)
	}

	receipt, err := elWriter.SetClaimerFor(context.Background(), config.ClaimerAddress, true)
	if err != nil {
		return err
	}

	logger.Infof(
		"%s Claimer address %s set successfully for operator %s\n",
		utils.EmojiCheckMark,
		config.ClaimerAddress,
		config.EarnerAddress.String(),
	)

	common.PrintTransactionInfo(
		receipt.TxHash.String(),
		config.ChainID,
	)

	return nil
}

func readAndValidateSetClaimerConfig(cCtx *cli.Context, logger logging.Logger) (*SetClaimerConfig, error) {
	network := cCtx.String(flags.NetworkFlag.Name)
	environment := cCtx.String(EnvironmentFlag.Name)
	rpcUrl := cCtx.String(flags.ETHRpcUrlFlag.Name)
	output := cCtx.String(flags.OutputFileFlag.Name)
	outputType := cCtx.String(flags.OutputTypeFlag.Name)
	earnerAddress := gethcommon.HexToAddress(cCtx.String(EarnerAddressFlag.Name))
	broadcast := cCtx.Bool(flags.BroadcastFlag.Name)
	claimerAddress := cCtx.String(ClaimerAddressFlag.Name)
	if common.IsEmptyString(claimerAddress) {
		return nil, fmt.Errorf("claimer address is required")
	}

	rewardsCoordinatorAddress := cCtx.String(RewardsCoordinatorAddressFlag.Name)
	var err error
	if common.IsEmptyString(rewardsCoordinatorAddress) {
		rewardsCoordinatorAddress, err = utils.GetRewardCoordinatorAddress(utils.NetworkNameToChainId(network))
		if err != nil {
			return nil, err
		}
	}
	logger.Debugf("Using Rewards Coordinator address: %s", rewardsCoordinatorAddress)

	chainID := utils.NetworkNameToChainId(network)
	logger.Debugf("Using chain ID: %s", chainID.String())

	if common.IsEmptyString(environment) {
		environment = getEnvFromNetwork(network)
	}
	logger.Debugf("Using network %s and environment: %s", network, environment)

	// Get SignerConfig
	signerConfig, err := common.GetSignerConfig(cCtx, logger)
	if err != nil {
		// We don't want to throw error since people can still use it to generate the
		// set claimer calldata/output without broadcasting it
		logger.Debugf("Failed to get signer config: %s", err)
	}

	return &SetClaimerConfig{
		ClaimerAddress:            gethcommon.HexToAddress(claimerAddress),
		Network:                   network,
		RPCUrl:                    rpcUrl,
		Broadcast:                 broadcast,
		RewardsCoordinatorAddress: gethcommon.HexToAddress(rewardsCoordinatorAddress),
		ChainID:                   chainID,
		SignerConfig:              signerConfig,
		EarnerAddress:             earnerAddress,
		Output:                    output,
		OutputType:                outputType,
	}, nil
}
