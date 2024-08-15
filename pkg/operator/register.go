package operator

import (
	"context"
	"fmt"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	elContracts "github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	"github.com/Layr-Labs/eigensdk-go/chainio/txmgr"
	eigenMetrics "github.com/Layr-Labs/eigensdk-go/metrics"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/urfave/cli/v2"
)

func RegisterCmd(p utils.Prompter) *cli.Command {
	registerCmd := &cli.Command{
		Name:      "register",
		Usage:     "Register the operator to EigenLayer contracts",
		UsageText: "register <configuration-file>",
		Description: `
		Register command expects a yaml config file as an argument
		to successfully register an operator address to eigenlayer

		This will register operator to DelegationManager
		`,
		After: telemetry.AfterRunAction(),
		Flags: []cli.Flag{
			&flags.VerboseFlag,
		},
		Action: func(cCtx *cli.Context) error {
			logger := common.GetLogger(cCtx)

			args := cCtx.Args()
			if args.Len() != 1 {
				return fmt.Errorf("%w: accepts 1 arg, received %d", ErrInvalidNumberOfArgs, args.Len())
			}

			configurationFilePath := args.Get(0)
			operatorCfg, err := common.ValidateAndReturnConfig(configurationFilePath, logger)
			if err != nil {
				return err
			}
			cCtx.App.Metadata["network"] = operatorCfg.ChainId.String()

			logger.Infof(
				"%s Operator configuration file validated successfully %s",
				utils.EmojiCheckMark,
				operatorCfg.Operator.Address,
			)

			ctx := context.Background()

			ethClient, err := ethclient.Dial(operatorCfg.EthRPCUrl)
			if err != nil {
				return err
			}

			keyWallet, sender, err := common.GetWallet(
				operatorCfg.SignerConfig,
				operatorCfg.Operator.Address,
				ethClient,
				p,
				operatorCfg.ChainId,
				logger,
			)
			if err != nil {
				return err
			}

			txMgr := txmgr.NewSimpleTxManager(keyWallet, ethClient, logger, sender)
			noopMetrics := eigenMetrics.NewNoopMetrics()
			contractCfg := elcontracts.Config{
				DelegationManagerAddress: gethcommon.HexToAddress(operatorCfg.ELDelegationManagerAddress),
				AvsDirectoryAddress:      gethcommon.HexToAddress(operatorCfg.ELAVSDirectoryAddress),
			}
			elWriter, err := elcontracts.NewWriterFromConfig(contractCfg, ethClient, logger, noopMetrics, txMgr)
			if err != nil {
				return err
			}

			elReader, err := elContracts.NewReaderFromConfig(
				contractCfg,
				ethClient,
				logger,
			)
			if err != nil {
				return err
			}

			status, err := elReader.IsOperatorRegistered(&bind.CallOpts{Context: ctx}, operatorCfg.Operator)
			if err != nil {
				return err
			}

			if !status {
				receipt, err := elWriter.RegisterAsOperator(ctx, operatorCfg.Operator, true)
				if err != nil {
					return err
				}
				common.PrintRegistrationInfo(
					receipt.TxHash.String(),
					gethcommon.HexToAddress(operatorCfg.Operator.Address),
					&operatorCfg.ChainId,
				)
			} else {
				logger.Infof("%s Operator is already registered on EigenLayer", utils.EmojiCheckMark)
				return nil
			}

			logger.Infof(
				"%s Operator is registered successfully to EigenLayer. There is a 30 minute delay between registration and operator details being shown in our webapp.",
				utils.EmojiCheckMark,
			)
			return nil
		},
	}

	return registerCmd
}
