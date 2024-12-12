package operator

import (
	"sort"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/operator/split"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/rewards"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	"github.com/Layr-Labs/eigensdk-go/logging"
	eigenSdkUtils "github.com/Layr-Labs/eigensdk-go/utils"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli/v2"
)

func SetOperatorSplitCmd(p utils.Prompter) *cli.Command {
	var operatorSplitCmd = &cli.Command{
		Name:  "set-rewards-split",
		Usage: "Set operator rewards split",
		Action: func(cCtx *cli.Context) error {
			return SetOperatorSplit(cCtx, p)
		},
		After: telemetry.AfterRunAction(),
		Flags: getSetOperatorSplitFlags(),
	}

	return operatorSplitCmd
}

func SetOperatorSplit(cCtx *cli.Context, p utils.Prompter) error {
	ctx := cCtx.Context
	logger := common.GetLogger(cCtx)

	config, err := readAndValidateSetOperatorSplitConfig(cCtx, logger)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to read and validate operator split config", err)
	}

	cCtx.App.Metadata["network"] = config.ChainID.String()

	ethClient, err := ethclient.Dial(config.RPCUrl)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to create new eth client", err)
	}

	eLWriter, err := common.GetELWriter(
		config.OperatorAddress,
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

	logger.Infof("Broadcasting set operator transaction...")

	var receipt *types.Receipt

	receipt, err = eLWriter.SetOperatorAVSSplit(ctx, config.OperatorAddress, config.AVSAddress, config.Split, true)

	if err != nil {
		return eigenSdkUtils.WrapError("failed to process claim", err)
	}

	logger.Infof("Set operator transaction submitted successfully")
	common.PrintTransactionInfo(receipt.TxHash.String(), config.ChainID)

	return nil
}

func getSetOperatorSplitFlags() []cli.Flag {
	baseFlags := []cli.Flag{
		&flags.NetworkFlag,
		&flags.ETHRpcUrlFlag,
		&flags.OperatorAddressFlag,
		&split.OperatorSplitFlag,
		&rewards.RewardsCoordinatorAddressFlag,
		&split.AVSAddressFlag,
	}

	allFlags := append(baseFlags, flags.GetSignerFlags()...)
	sort.Sort(cli.FlagsByName(allFlags))
	return allFlags
}

func readAndValidateSetOperatorSplitConfig(
	cCtx *cli.Context,
	logger logging.Logger,
) (*split.SetOperatorAVSSplitConfig, error) {
	network := cCtx.String(flags.NetworkFlag.Name)
	rpcUrl := cCtx.String(flags.ETHRpcUrlFlag.Name)
	opSplit := cCtx.Int(split.OperatorSplitFlag.Name)
	rewardsCoordinatorAddress := cCtx.String(rewards.RewardsCoordinatorAddressFlag.Name)

	var err error
	if common.IsEmptyString(rewardsCoordinatorAddress) {
		rewardsCoordinatorAddress, err = common.GetRewardCoordinatorAddress(utils.NetworkNameToChainId(network))
		if err != nil {
			return nil, err
		}
	}
	logger.Debugf("Using Rewards Coordinator address: %s", rewardsCoordinatorAddress)

	operatorAddress := gethcommon.HexToAddress(cCtx.String(flags.OperatorAddressFlag.Name))
	logger.Infof("Using operator address: %s", operatorAddress.String())

	avsAddress := gethcommon.HexToAddress(cCtx.String(split.AVSAddressFlag.Name))
	logger.Infof("Using AVS address: %s", avsAddress.String())

	chainID := utils.NetworkNameToChainId(network)
	logger.Debugf("Using chain ID: %s", chainID.String())

	// Get SignerConfig
	signerConfig, err := common.GetSignerConfig(cCtx, logger)
	if err != nil {
		// We don't want to throw error since people can still use it to generate the claim
		// without broadcasting it
		logger.Debugf("Failed to get signer config: %s", err)
	}

	return &split.SetOperatorAVSSplitConfig{
		Network:                   network,
		RPCUrl:                    rpcUrl,
		RewardsCoordinatorAddress: gethcommon.HexToAddress(rewardsCoordinatorAddress),
		ChainID:                   chainID,
		SignerConfig:              signerConfig,
		OperatorAddress:           operatorAddress,
		AVSAddress:                avsAddress,
		Split:                     uint16(opSplit),
	}, nil
}
