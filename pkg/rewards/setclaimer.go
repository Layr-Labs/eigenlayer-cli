package rewards

import (
	"context"
	"fmt"
	"sort"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/command"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	"github.com/Layr-Labs/eigensdk-go/logging"
	eigenSdkUtils "github.com/Layr-Labs/eigensdk-go/utils"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/urfave/cli/v2"
)

type SetClaimerCmd struct {
	prompter utils.Prompter
}

func NewSetClaimerCmd(p utils.Prompter) *cli.Command {
	delegateCommand := &SetClaimerCmd{prompter: p}
	setClaimerCmd := command.NewWriteableCallDataCommand(
		delegateCommand,
		"set-claimer",
		"Set the claimer address for the earner",
		"set-claimer",
		`Set the rewards claimer address for the earner.`,
		getSetClaimerFlags(),
	)

	return setClaimerCmd
}

func getSetClaimerFlags() []cli.Flag {
	baseFlags := []cli.Flag{
		&flags.NetworkFlag,
		&flags.ETHRpcUrlFlag,
		&EarnerAddressFlag,
		&RewardsCoordinatorAddressFlag,
		&ClaimerAddressFlag,
		&flags.VerboseFlag,
	}
	sort.Sort(cli.FlagsByName(baseFlags))
	return baseFlags
}

func (s SetClaimerCmd) Execute(cCtx *cli.Context) error {
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

		noSendTxOpts := common.GetNoSendTxOpts(config.CallerAddress)
		// If caller is a smart contract, we can't estimate gas using geth
		// since balance of contract can be 0, as it can be called by an EOA
		// to claim. So we hardcode the gas limit to 150_000 so that we can
		// create unsigned tx without gas limit estimation from contract bindings
		if common.IsSmartContractAddress(config.CallerAddress, ethClient) {
			// Caller is a smart contract
			noSendTxOpts.GasLimit = 150_000
		}
		unsignedTx, err := contractBindings.RewardsCoordinator.SetClaimerFor(noSendTxOpts, config.ClaimerAddress)
		if err != nil {
			return eigenSdkUtils.WrapError("failed to create unsigned tx", err)
		}

		if config.OutputType == string(common.OutputType_Calldata) {
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
		config.CallerAddress,
		config.SignerConfig,
		ethClient,
		elcontracts.Config{
			RewardsCoordinatorAddress: config.RewardsCoordinatorAddress,
		},
		s.prompter,
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
	callerAddress := common.PopulateCallerAddress(cCtx, logger, earnerAddress)

	rewardsCoordinatorAddress := cCtx.String(RewardsCoordinatorAddressFlag.Name)
	var err error
	if common.IsEmptyString(rewardsCoordinatorAddress) {
		rewardsCoordinatorAddress, err = common.GetRewardCoordinatorAddress(utils.NetworkNameToChainId(network))
		if err != nil {
			return nil, err
		}
	}
	logger.Debugf("Using Rewards Coordinator address: %s", rewardsCoordinatorAddress)

	chainID := utils.NetworkNameToChainId(network)
	logger.Debugf("Using chain ID: %s", chainID.String())

	if common.IsEmptyString(environment) {
		environment = common.GetEnvFromNetwork(network)
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
		CallerAddress:             callerAddress,
		Output:                    output,
		OutputType:                outputType,
	}, nil
}
