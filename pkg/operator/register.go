package operator

import (
	"context"
	"fmt"
	"os"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	elContracts "github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/eth"
	"github.com/Layr-Labs/eigensdk-go/chainio/txmgr"
	eigensdkLogger "github.com/Layr-Labs/eigensdk-go/logging"
	eigenMetrics "github.com/Layr-Labs/eigensdk-go/metrics"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethcommon "github.com/ethereum/go-ethereum/common"

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
		Action: func(cCtx *cli.Context) error {
			args := cCtx.Args()
			if args.Len() != 1 {
				return fmt.Errorf("%w: accepts 1 arg, received %d", ErrInvalidNumberOfArgs, args.Len())
			}

			configurationFilePath := args.Get(0)
			operatorCfg, err := common.ValidateAndReturnConfig(configurationFilePath)
			if err != nil {
				return err
			}
			cCtx.App.Metadata["network"] = operatorCfg.ChainId.String()

			fmt.Printf(
				"\r%s Operator configuration file validated successfully %s\n",
				utils.EmojiCheckMark,
				operatorCfg.Operator.Address,
			)

			ctx := context.Background()
			logger := eigensdkLogger.NewTextSLogger(os.Stdout, &eigensdkLogger.SLoggerOptions{})

			ethClient, err := eth.NewClient(operatorCfg.EthRPCUrl)
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

			elWriter, err := elContracts.BuildELChainWriter(
				gethcommon.HexToAddress(operatorCfg.ELDelegationManagerAddress),
				gethcommon.HexToAddress(operatorCfg.ELAVSDirectoryAddress),
				ethClient,
				logger,
				noopMetrics,
				txMgr)

			if err != nil {
				return err
			}

			reader, err := elContracts.BuildELChainReader(
				gethcommon.HexToAddress(operatorCfg.ELDelegationManagerAddress),
				gethcommon.HexToAddress(operatorCfg.ELAVSDirectoryAddress),
				ethClient,
				logger,
			)
			if err != nil {
				return err
			}

			status, err := reader.IsOperatorRegistered(&bind.CallOpts{Context: ctx}, operatorCfg.Operator)
			if err != nil {
				return err
			}

			if !status {
				receipt, err := elWriter.RegisterAsOperator(ctx, operatorCfg.Operator)
				if err != nil {
					fmt.Printf("%s Error while registering operator\n", utils.EmojiCrossMark)
					return err
				}
				common.PrintRegistrationInfo(
					receipt.TxHash.String(),
					gethcommon.HexToAddress(operatorCfg.Operator.Address),
					&operatorCfg.ChainId,
				)
			} else {
				fmt.Printf("%s Operator is already registered on EigenLayer\n", utils.EmojiCheckMark)
				return nil
			}

			fmt.Printf(
				"%s Operator is registered successfully to EigenLayer. There is a 30 minute delay between registration and operator details being shown in our webapp.\n",
				utils.EmojiCheckMark,
			)
			return nil
		},
	}

	return registerCmd
}
