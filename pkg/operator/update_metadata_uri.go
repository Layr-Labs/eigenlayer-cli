package operator

import (
	"context"
	"fmt"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	eigenSdkUtils "github.com/Layr-Labs/eigensdk-go/utils"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/urfave/cli/v2"
)

func UpdateMetadataURICmd(p utils.Prompter) *cli.Command {
	updateMetadataURICmd := &cli.Command{
		Name:      "update-metadata-uri",
		Usage:     "Update the operator metadata uri onchain",
		UsageText: "update-metadata-uri <configuration-file>",
		Description: `
Updates the operator metadata uri onchain

Requires the same file used for registration as argument
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

			receipt, err := elWriter.UpdateMetadataURI(context.Background(), operatorCfg.Operator.MetadataUrl, operatorCfg.Operator.Address, true)
			if err != nil {
				fmt.Printf("%s Error while updating operator metadata uri\n", utils.EmojiCrossMark)
				return err
			}
			logger.Infof(
				"%s Operator metadata uri updated at: %s",
				utils.EmojiCheckMark,
				common.GetTransactionLink(receipt.TxHash.String(), &operatorCfg.ChainId),
			)

			common.PrintRegistrationInfo(
				"",
				gethcommon.HexToAddress(operatorCfg.Operator.Address),
				&operatorCfg.ChainId,
			)

			logger.Infof(
				"%s Operator metadata uri successfully. There is a 30 minute delay between update and operator metadata being shown in our webapp.",
				utils.EmojiCheckMark,
			)
			return nil
		},
	}

	return updateMetadataURICmd
}
