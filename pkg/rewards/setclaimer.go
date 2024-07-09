package rewards

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"os"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/types"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/Layr-Labs/eigensdk-go/logging"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/eth"
	"github.com/Layr-Labs/eigensdk-go/chainio/txmgr"
	eigensdkLogger "github.com/Layr-Labs/eigensdk-go/logging"
	eigenMetrics "github.com/Layr-Labs/eigensdk-go/metrics"

	gethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/urfave/cli/v2"
)

type SetClaimerConfig struct {
	ClaimerAddress            string
	Network                   string
	RPCUrl                    string
	Broadcast                 bool
	RewardsCoordinatorAddress gethcommon.Address
	ChainID                   *big.Int
	SignerConfig              *types.SignerConfig
	EarnerAddress             gethcommon.Address
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
		Flags: []cli.Flag{
			&flags.VerboseFlag,
			&flags.NetworkFlag,
			&flags.ETHRpcUrlFlag,
			&flags.EarnerAddressFlag,
			&flags.OutputFileFlag,
			&flags.PathToKeyStoreFlag,
			&flags.EcdsaPrivateKeyFlag,
			&flags.BroadcastFlag,
			&RewardsCoordinatorAddressFlag,
			&ClaimerAddressFlag,
		},
		Action: func(cCtx *cli.Context) error {
			return SetClaimer(cCtx, p)
		},
	}

	return setClaimerCmd
}

func SetClaimer(cCtx *cli.Context, p utils.Prompter) error {
	verbose := cCtx.Bool(flags.VerboseFlag.Name)
	logLevel := slog.LevelInfo
	if verbose {
		logLevel = slog.LevelDebug
	}
	logger := logging.NewTextSLogger(os.Stdout, &eigensdkLogger.SLoggerOptions{Level: logLevel})
	config, err := readAndValidateSetClaimerConfig(cCtx, logger)
	if err != nil {
		return err
	}

	cCtx.App.Metadata["network"] = config.ChainID.String()
	if config.ChainID.Int64() == utils.MainnetChainId {
		return fmt.Errorf("set claimer currently unsupported on mainnet")
	}

	if !config.Broadcast {
		fmt.Printf(
			"Claimer address %s will be set for earner %s\n",
			config.ClaimerAddress,
			config.EarnerAddress.Hex(),
		)
		return nil
	}

	ethClient, err := eth.NewClient(config.RPCUrl)
	if err != nil {
		return err
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
			sender.Hex(),
			config.EarnerAddress.Hex(),
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

	receipt, err := elWriter.SetClaimerFor(context.Background(), gethcommon.HexToAddress(config.ClaimerAddress))
	if err != nil {
		return err
	}

	fmt.Printf(
		"%s Claimer address %s set successfully for operator %s\n",
		utils.EmojiCheckMark,
		config.ClaimerAddress,
		config.EarnerAddress.Hex(),
	)

	common.PrintRegistrationInfo(
		receipt.TxHash.String(),
		config.EarnerAddress,
		config.ChainID,
	)

	return nil
}

func readAndValidateSetClaimerConfig(cCtx *cli.Context, logger logging.Logger) (*SetClaimerConfig, error) {
	network := cCtx.String(flags.NetworkFlag.Name)
	environment := cCtx.String(EnvironmentFlag.Name)
	rpcUrl := cCtx.String(flags.ETHRpcUrlFlag.Name)
	earnerAddress := gethcommon.HexToAddress(cCtx.String(flags.EarnerAddressFlag.Name))
	broadcast := cCtx.Bool(flags.BroadcastFlag.Name)
	claimerAddress := cCtx.String(ClaimerAddressFlag.Name)
	rewardsCoordinatorAddress := cCtx.String(RewardsCoordinatorAddressFlag.Name)
	var err error
	if rewardsCoordinatorAddress == "" {
		rewardsCoordinatorAddress, err = utils.GetRewardCoordinatorAddress(utils.NetworkNameToChainId(network))
		if err != nil {
			return nil, err
		}
	}
	logger.Debugf("Using Rewards Coordinator address: %s", rewardsCoordinatorAddress)

	chainID := utils.NetworkNameToChainId(network)
	logger.Debugf("Using chain ID: %s", chainID.String())

	if environment == "" {
		environment = getEnvFromNetwork(network)
	}
	logger.Debugf("Using network %s and environment: %s", network, environment)

	// Get SignerConfig
	signerConfig, err := common.GetSignerConfig(cCtx)
	if err != nil {
		return nil, err
	}

	return &SetClaimerConfig{
		ClaimerAddress:            claimerAddress,
		Network:                   network,
		RPCUrl:                    rpcUrl,
		Broadcast:                 broadcast,
		RewardsCoordinatorAddress: gethcommon.HexToAddress(rewardsCoordinatorAddress),
		ChainID:                   chainID,
		SignerConfig:              signerConfig,
		EarnerAddress:             earnerAddress,
	}, nil
}
