package rewards

import (
	"errors"
	"sort"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
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
		Name:  "set-operator-split",
		Usage: "Set operator split",
		Action: func(cCtx *cli.Context) error {
			return SetOperatorSplit(cCtx, p)
		},
		After: telemetry.AfterRunAction(),
		Flags: getOperatorSplitFlags(),
	}

	return operatorSplitCmd
}

func GetOperatorSplitCmd(p utils.Prompter) *cli.Command {
	var operatorSplitCmd = &cli.Command{
		Name:  "get-operator-split",
		Usage: "Get operator split",
		Action: func(cCtx *cli.Context) error {
			return GetOperatorSplit(cCtx, p)
		},
		After: telemetry.AfterRunAction(),
		Flags: getOperatorSplitFlags(),
	}

	return operatorSplitCmd
}

func GetOperatorSplit(cCtx *cli.Context, p utils.Prompter) error {
	ctx := cCtx.Context
	logger := common.GetLogger(cCtx)

	config, err := readAndValidateOperatorSplitConfig(cCtx, logger)
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

	split, err := elReader.GetOperatorAVSSplit(ctx, config.OperatorAddress, config.AVSAddress)

	if err != nil || split == nil {
		return eigenSdkUtils.WrapError("failed to get operator split", err)
	}

	logger.Infof("Operator split is %d", *split)

	return nil
}

func SetOperatorSplit(cCtx *cli.Context, p utils.Prompter) error {
	ctx := cCtx.Context
	logger := common.GetLogger(cCtx)

	config, err := readAndValidateOperatorSplitConfig(cCtx, logger)
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

func readAndValidateOperatorSplitConfig(cCtx *cli.Context, logger logging.Logger) (*SetOperatorAVSSplitConfig, error) {
	network := cCtx.String(flags.NetworkFlag.Name)
	environment := cCtx.String(EnvironmentFlag.Name)
	rpcUrl := cCtx.String(flags.ETHRpcUrlFlag.Name)
	broadcast := cCtx.Bool(flags.BroadcastFlag.Name)
	split := cCtx.Int(OperatorSplitFlag.Name)
	rewardsCoordinatorAddress := cCtx.String(RewardsCoordinatorAddressFlag.Name)

	var err error
	if common.IsEmptyString(rewardsCoordinatorAddress) {
		rewardsCoordinatorAddress, err = common.GetRewardCoordinatorAddress(utils.NetworkNameToChainId(network))
		if err != nil {
			return nil, err
		}
	}
	logger.Debugf("Using Rewards Coordinator address: %s", rewardsCoordinatorAddress)

	claimTimestamp := cCtx.String(ClaimTimestampFlag.Name)
	logger.Debugf("Using claim timestamp from user: %s", claimTimestamp)

	operatorAddress := gethcommon.HexToAddress(cCtx.String(OperatorAddressFlag.Name))
	logger.Infof("Using operator address: %s", operatorAddress.String())

	avsAddress := gethcommon.HexToAddress(cCtx.String(AVSAddressesFlag.Name))
	logger.Infof("Using AVS address: %s", avsAddress.String())

	chainID := utils.NetworkNameToChainId(network)
	logger.Debugf("Using chain ID: %s", chainID.String())

	proofStoreBaseURL := cCtx.String(ProofStoreBaseURLFlag.Name)

	// If empty get from utils
	if common.IsEmptyString(proofStoreBaseURL) {
		proofStoreBaseURL = getProofStoreBaseURL(network)

		// If still empty return error
		if common.IsEmptyString(proofStoreBaseURL) {
			return nil, errors.New("proof store base URL not provided")
		}
	}
	logger.Debugf("Using Proof store base URL: %s", proofStoreBaseURL)

	if common.IsEmptyString(environment) {
		environment = getEnvFromNetwork(network)
	}
	logger.Debugf("Using network %s and environment: %s", network, environment)

	// Get SignerConfig
	signerConfig, err := common.GetSignerConfig(cCtx, logger)
	if err != nil {
		// We don't want to throw error since people can still use it to generate the claim
		// without broadcasting it
		logger.Debugf("Failed to get signer config: %s", err)
	}

	// TODO(shrimalmadhur): Fix to make sure correct S3 bucket is used. Clean up later
	if network == utils.MainnetNetworkName {
		network = "ethereum"
	}

	return &SetOperatorAVSSplitConfig{
		Network:                   network,
		RPCUrl:                    rpcUrl,
		Broadcast:                 broadcast,
		RewardsCoordinatorAddress: gethcommon.HexToAddress(rewardsCoordinatorAddress),
		ChainID:                   chainID,
		SignerConfig:              signerConfig,
		OperatorAddress:           operatorAddress,
		AVSAddress:                avsAddress,
		Split:                     uint16(split),
	}, nil
}

func getOperatorSplitFlags() []cli.Flag {
	baseFlags := []cli.Flag{
		&flags.NetworkFlag,
		&flags.ETHRpcUrlFlag,
		&flags.OutputFileFlag,
		&OperatorAddressFlag,
		&OperatorSplitFlag,
		&RewardsCoordinatorAddressFlag,
		&ClaimTimestampFlag,
		&AVSAddressesFlag,
	}

	allFlags := append(baseFlags, flags.GetSignerFlags()...)
	sort.Sort(cli.FlagsByName(allFlags))
	return allFlags
}
