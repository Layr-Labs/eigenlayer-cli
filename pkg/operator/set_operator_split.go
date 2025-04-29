package operator

import (
	"fmt"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/command"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/operator/split"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/rewards"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	contractRewardsCoordinator "github.com/Layr-Labs/eigensdk-go/contracts/bindings/RewardsCoordinator"
	"github.com/Layr-Labs/eigensdk-go/logging"
	eigenSdkUtils "github.com/Layr-Labs/eigensdk-go/utils"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli/v2"
)

type SetOperatorSplitCmd struct {
	prompter                utils.Prompter
	isProgrammaticIncentive bool
	isOperatorSet           bool
}

func NewSetOperatorSplitCmd(p utils.Prompter, isProgrammaticIncentive bool, isOperatorSet bool) *cli.Command {
	delegateCommand := &SetOperatorSplitCmd{
		prompter:                p,
		isProgrammaticIncentive: isProgrammaticIncentive,
		isOperatorSet:           isOperatorSet,
	}
	setOperatorSplitCmd := command.NewWriteableCallDataCommand(
		delegateCommand,
		"set-rewards-split",
		"Set operator rewards split",
		"",
		"",
		getSetOperatorSplitFlags(),
	)

	return setOperatorSplitCmd
}

func (s SetOperatorSplitCmd) Execute(cCtx *cli.Context) error {
	return SetOperatorSplit(cCtx, s.prompter, s.isProgrammaticIncentive, s.isOperatorSet)
}

func SetOperatorSplit(cCtx *cli.Context, p utils.Prompter, isProgrammaticIncentive bool, isOperatorSet bool) error {
	ctx := cCtx.Context
	logger := common.GetLogger(cCtx)

	config, err := readAndValidateSetOperatorSplitConfig(cCtx, logger, isProgrammaticIncentive, isOperatorSet)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to read and validate operator split config", err)
	}

	cCtx.App.Metadata["network"] = config.ChainID.String()

	ethClient, err := ethclient.Dial(config.RPCUrl)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to create new eth client", err)
	}

	if config.Broadcast {

		eLWriter, err := common.GetELWriter(
			config.CallerAddress,
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
		if isOperatorSet {
			operatorSet := contractRewardsCoordinator.OperatorSet{
				Id:  uint32(config.OperatorSetId),
				Avs: config.AVSAddress,
			}
			receipt, err = eLWriter.SetOperatorSetSplit(ctx, config.OperatorAddress, operatorSet, config.Split, true)
		} else if isProgrammaticIncentive {
			receipt, err = eLWriter.SetOperatorPISplit(ctx, config.OperatorAddress, config.Split, true)

		} else {
			receipt, err = eLWriter.SetOperatorAVSSplit(ctx, config.OperatorAddress, config.AVSAddress, config.Split, true)
		}
		if err != nil {
			return eigenSdkUtils.WrapError("failed to process claim", err)
		}

		logger.Infof("Set operator transaction submitted successfully")
		common.PrintTransactionInfo(receipt.TxHash.String(), config.ChainID)
	} else {

		noSendTxOpts := common.GetNoSendTxOpts(config.CallerAddress)
		// If the caller is a smart contract, we can't estimate gas using geth
		// since balance of contract can be 0, as it can be called by an EOA
		// to claim. So we hardcode the gas limit to 150_000 so that we can
		// create unsigned tx without gas limit estimation from contract bindings
		if common.IsSmartContractAddress(config.CallerAddress, ethClient) {
			// Caller is a smart contract
			noSendTxOpts.GasLimit = 150_000
		}
		_, _, contractBindings, err := elcontracts.BuildClients(elcontracts.Config{
			RewardsCoordinatorAddress: config.RewardsCoordinatorAddress,
		}, ethClient, nil, logger, nil)
		if err != nil {
			return err
		}

		var unsignedTx *types.Transaction
		if isOperatorSet {
			operatorSet := contractRewardsCoordinator.OperatorSet{
				Id:  uint32(config.OperatorSetId),
				Avs: config.AVSAddress,
			}
			unsignedTx, err = contractBindings.RewardsCoordinator.SetOperatorSetSplit(noSendTxOpts, config.OperatorAddress, operatorSet, config.Split)
		} else if isProgrammaticIncentive {
			unsignedTx, err = contractBindings.RewardsCoordinator.SetOperatorPISplit(noSendTxOpts, config.OperatorAddress, config.Split)
		} else {
			unsignedTx, err = contractBindings.RewardsCoordinator.SetOperatorAVSSplit(noSendTxOpts, config.OperatorAddress, config.AVSAddress, config.Split)
		}

		if err != nil {
			return eigenSdkUtils.WrapError("failed to create unsigned tx", err)
		}
		if config.OutputType == utils.CallDataOutputType {
			calldataHex := gethcommon.Bytes2Hex(unsignedTx.Data())

			if !common.IsEmptyString(config.OutputFile) {
				err = common.WriteToFile([]byte(calldataHex), config.OutputFile)
				if err != nil {
					return err
				}
				logger.Infof("Call data written to file: %s", config.OutputFile)
			} else {
				fmt.Println(calldataHex)
			}
		} else {
			logger.Infof("This transaction would set the operator split to %d", config.Split)
		}

		if !config.IsSilent {
			txFeeDetails := common.GetTxFeeDetails(unsignedTx)
			fmt.Println()
			txFeeDetails.Print()

			fmt.Println("To broadcast the operator set split, use the --broadcast flag")
		}
	}
	return nil
}

func getSetOperatorSplitFlags() []cli.Flag {
	return []cli.Flag{
		&flags.NetworkFlag,
		&flags.ETHRpcUrlFlag,
		&flags.OperatorAddressFlag,
		&split.OperatorSplitFlag,
		&rewards.RewardsCoordinatorAddressFlag,
		&split.AVSAddressFlag,
		&flags.SilentFlag,
	}
}

func readAndValidateSetOperatorSplitConfig(
	cCtx *cli.Context,
	logger logging.Logger,
	isProgrammaticIncentive bool,
	isOperatorSet bool,
) (*split.SetOperatorAVSSplitConfig, error) {
	network := cCtx.String(flags.NetworkFlag.Name)
	rpcUrl := cCtx.String(flags.ETHRpcUrlFlag.Name)
	opSplit := cCtx.Int(split.OperatorSplitFlag.Name)
	broadcast := cCtx.Bool(flags.BroadcastFlag.Name)
	outputType := cCtx.String(flags.OutputTypeFlag.Name)
	outputFile := cCtx.String(flags.OutputFileFlag.Name)
	isSilent := cCtx.Bool(flags.SilentFlag.Name)

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
	callerAddress := common.PopulateCallerAddress(cCtx, logger, operatorAddress, flags.OperatorAddressFlag.Name)

	avsAddress := gethcommon.HexToAddress(cCtx.String(split.AVSAddressFlag.Name))

	if !isProgrammaticIncentive {
		logger.Infof("Using AVS address: %s", avsAddress.String())
	}

	var operatorSetId int
	if isOperatorSet {
		operatorSetId = cCtx.Int(split.OperatorSetIdFlag.Name)
	}

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
		CallerAddress:             callerAddress,
		Split:                     uint16(opSplit),
		OperatorSetId:             operatorSetId,
		Broadcast:                 broadcast,
		OutputType:                outputType,
		OutputFile:                outputFile,
		IsSilent:                  isSilent,
	}, nil
}
