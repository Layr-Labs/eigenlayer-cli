package operator

import (
	"errors"
	"sort"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/operator/split"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/rewards"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	contractIRewardsCoordinator "github.com/Layr-Labs/eigensdk-go/contracts/bindings/IRewardsCoordinator"
	"github.com/Layr-Labs/eigensdk-go/logging"
	eigenSdkUtils "github.com/Layr-Labs/eigensdk-go/utils"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli/v2"
)

func GetOperatorSplitCmd(p utils.Prompter) *cli.Command {
	var operatorSplitCmd = &cli.Command{
		Name:  "get-rewards-split",
		Usage: "Get operator rewards split",
		Action: func(cCtx *cli.Context) error {
			return GetOperatorSplit(cCtx, false, false)
		},
		After: telemetry.AfterRunAction(),
		Flags: getGetOperatorSplitFlags(),
	}

	return operatorSplitCmd
}

func getGetOperatorSplitFlags() []cli.Flag {
	baseFlags := []cli.Flag{
		&flags.NetworkFlag,
		&flags.ETHRpcUrlFlag,
		&flags.OperatorAddressFlag,
		&split.OperatorSplitFlag,
		&rewards.RewardsCoordinatorAddressFlag,
		&split.AVSAddressFlag,
	}

	sort.Sort(cli.FlagsByName(baseFlags))
	return baseFlags
}

func GetOperatorSplit(cCtx *cli.Context, isProgrammaticIncentive bool, isOperatorSet bool) error {
	ctx := cCtx.Context
	logger := common.GetLogger(cCtx)

	config, err := readAndValidateGetOperatorSplitConfig(cCtx, logger, isProgrammaticIncentive, isOperatorSet)
	if config.Network != "holesky" {
		return errors.New("getting operator split only supported on holesky with this version of CLI")
	}

	if err != nil {
		return eigenSdkUtils.WrapError("failed to read and validate operator split config", err)
	}

	cCtx.App.Metadata["network"] = config.ChainID.String()

	ethClient, err := ethclient.Dial(config.RPCUrl)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to create new eth client", err)
	}

	elReader, err := elcontracts.NewReaderFromConfig(
		elcontracts.Config{
			RewardsCoordinatorAddress: config.RewardsCoordinatorAddress,
		},
		ethClient,
		logger,
	)

	if err != nil {
		return eigenSdkUtils.WrapError("failed to get EL writer", err)
	}

	logger.Infof("Getting operator split...")

	var split uint16
	if isOperatorSet {
		operatorSet := contractIRewardsCoordinator.OperatorSet{
			Id:  uint32(config.OperatorSetId),
			Avs: config.AVSAddress,
		}
		split, err = elReader.GetOperatorSetSplit(ctx, config.OperatorAddress, operatorSet)
	} else if isProgrammaticIncentive {
		split, err = elReader.GetOperatorPISplit(ctx, config.OperatorAddress)
	} else {
		split, err = elReader.GetOperatorAVSSplit(ctx, config.OperatorAddress, config.AVSAddress)
	}
	if err != nil {
		return eigenSdkUtils.WrapError("failed to get operator split", err)
	}

	logger.Infof("Operator split is %d", split)

	return nil
}

func readAndValidateGetOperatorSplitConfig(
	cCtx *cli.Context,
	logger logging.Logger,
	isProgrammaticIncentive bool,
	isOperatorSet bool,
) (*split.GetOperatorAVSSplitConfig, error) {
	network := cCtx.String(flags.NetworkFlag.Name)
	rpcUrl := cCtx.String(flags.ETHRpcUrlFlag.Name)

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
	if !isProgrammaticIncentive {
		logger.Infof("Using AVS address: %s", avsAddress.String())
	}

	var operatorSetId int
	if isOperatorSet {
		operatorSetId = cCtx.Int(split.OperatorSetIdFlag.Name)
		if operatorSetId == 0 {
			return nil, errors.New("operator set ID is required")
		}
	}

	chainID := utils.NetworkNameToChainId(network)
	logger.Debugf("Using chain ID: %s", chainID.String())

	return &split.GetOperatorAVSSplitConfig{
		Network:                   network,
		RPCUrl:                    rpcUrl,
		RewardsCoordinatorAddress: gethcommon.HexToAddress(rewardsCoordinatorAddress),
		ChainID:                   chainID,
		OperatorAddress:           operatorAddress,
		AVSAddress:                avsAddress,
		OperatorSetId:             operatorSetId,
	}, nil
}
