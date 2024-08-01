package rewards

import (
	"context"
	"fmt"
	"math/big"
	"sort"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/types"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	"github.com/Layr-Labs/eigensdk-go/chainio/txmgr"
	"github.com/Layr-Labs/eigensdk-go/logging"
	eigenMetrics "github.com/Layr-Labs/eigensdk-go/metrics"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/urfave/cli/v2"
)

type SetClaimerConfig struct {
	ClaimerAddress            gethcommon.Address
	Network                   string
	RPCUrl                    string
	Broadcast                 bool
	RewardsCoordinatorAddress gethcommon.Address
	ChainID                   *big.Int
	SignerConfig              *types.SignerConfig
	EarnerAddress             gethcommon.Address
	Output                    string
	OutputType                string
}

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
	if config.ChainID.Int64() == utils.MainnetChainId {
		return fmt.Errorf("set claimer currently unsupported on mainnet")
	}

	ethClient, err := ethclient.Dial(config.RPCUrl)
	if err != nil {
		return err
	}

	if !config.Broadcast {
		if config.OutputType == string(common.OutputType_Calldata) {
			_, _, contractBindings, err := elcontracts.BuildClients(elcontracts.Config{
				RewardsCoordinatorAddress: config.RewardsCoordinatorAddress,
			}, ethClient, nil, logger, nil)
			if err != nil {
				return err
			}

			noSendTxOpts := common.GetNoSendTxOpts(config.EarnerAddress)
			unsignedTx, err := contractBindings.RewardsCoordinator.SetClaimerFor(noSendTxOpts, config.ClaimerAddress)
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

		return nil
	}

	keyWallet, sender, err := common.GetWallet(
		*config.SignerConfig,
		config.EarnerAddress.Hex(),
		ethClient,
		p,
		*config.ChainID,
		logger,
	)
	if err != nil {
		return err
	}

	if sender != config.EarnerAddress {
		return fmt.Errorf(
			"signer address(%s) and earner addresses(%s) do not match",
			sender.String(),
			config.EarnerAddress.String(),
		)
	}

	txMgr := txmgr.NewSimpleTxManager(keyWallet, ethClient, logger, sender)
	noopMetrics := eigenMetrics.NewNoopMetrics()
	contractCfg := elcontracts.Config{
		RewardsCoordinatorAddress: config.RewardsCoordinatorAddress,
	}

	elWriter, err := elcontracts.NewWriterFromConfig(contractCfg, ethClient, logger, noopMetrics, txMgr)
	if err != nil {
		return err
	}

	receipt, err := elWriter.SetClaimerFor(context.Background(), config.ClaimerAddress)
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
