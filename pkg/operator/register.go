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
	eigenSdkUtils "github.com/Layr-Labs/eigensdk-go/utils"

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
			logger.Debugf("operatorCfg: %v", operatorCfg)
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

			contractCfg := elcontracts.Config{
				DelegationManagerAddress: gethcommon.HexToAddress(operatorCfg.ELDelegationManagerAddress),
				AvsDirectoryAddress:      gethcommon.HexToAddress(operatorCfg.ELAVSDirectoryAddress),
			}

			elWriter, err := common.GetELWriter(
				gethcommon.HexToAddress(operatorCfg.Operator.Address),
				&operatorCfg.SignerConfig,
				ethClient,
				contractCfg,
				p,
				&operatorCfg.ChainId,
				logger,
			)

			if err != nil {
				return eigenSdkUtils.WrapError("failed to get EL writer", err)
			}

			elReader, err := elContracts.NewReaderFromConfig(
				contractCfg,
				ethClient,
				logger,
			)
			if err != nil {
				return err
			}

			status, err := elReader.IsOperatorRegistered(ctx, operatorCfg.Operator)
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
