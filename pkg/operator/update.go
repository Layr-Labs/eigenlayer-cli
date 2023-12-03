package operator

import (
	"context"
	"fmt"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/types"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	eigenChainio "github.com/Layr-Labs/eigensdk-go/chainio/clients"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/eth"
	elContracts "github.com/Layr-Labs/eigensdk-go/chainio/elcontracts"
	eigensdkLogger "github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/Layr-Labs/eigensdk-go/metrics"
	eigensdkUtils "github.com/Layr-Labs/eigensdk-go/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
)

func UpdateCmd(p utils.Prompter) *cli.Command {
	updateCmd := &cli.Command{
		Name:      "update",
		Usage:     "Update the operator metadata onchain",
		UsageText: "update <configuration-file>",
		Description: `
		Updates the operator metadata onchain which includes 
			- metadata url
			- delegation approver address
			- earnings receiver address
			- staker opt out window blocks

		Requires the same file used for registration as argument
		`,
		Action: func(cCtx *cli.Context) error {
			args := cCtx.Args()
			if args.Len() != 1 {
				return fmt.Errorf("%w: accepts 1 arg, received %d", ErrInvalidNumberOfArgs, args.Len())
			}

			configurationFilePath := args.Get(0)
			var operatorCfg types.OperatorConfig
			err := eigensdkUtils.ReadYamlConfig(configurationFilePath, &operatorCfg)
			if err != nil {
				return err
			}
			fmt.Printf(
				"Operator configuration file read successfully %s %s\n",
				operatorCfg.Operator.Address,
				utils.EmojiCheckMark,
			)
			signerType, err := validateSignerType(operatorCfg)
			if err != nil {
				return err
			}

			logger, err := eigensdkLogger.NewZapLogger(eigensdkLogger.Development)
			if err != nil {
				return err
			}

			localSigner, err := getSigner(p, signerType, operatorCfg)
			if err != nil {
				return err
			}

			ethClient, err := eth.NewClient(operatorCfg.EthRPCUrl)
			if err != nil {
				return err
			}

			elContractsClient, err := eigenChainio.NewELContractsChainClient(
				common.HexToAddress(operatorCfg.ELSlasherAddress),
				common.HexToAddress(operatorCfg.BlsPublicKeyCompendiumAddress),
				ethClient,
				ethClient,
				logger,
			)
			if err != nil {
				return err
			}

			noopMetrics := metrics.NewNoopMetrics()
			elWriter := elContracts.NewELChainWriter(
				elContractsClient,
				ethClient,
				localSigner,
				logger,
				noopMetrics,
			)

			if err != nil {
				return err
			}
			receipt, err := elWriter.UpdateOperatorDetails(context.Background(), operatorCfg.Operator)
			if err != nil {
				logger.Errorf("Error while updating operator details: %s", utils.EmojiCrossMark)
				return err
			}
			logger.Infof(
				"Operator details updated at: %s %s",
				getTransactionLink(receipt.TxHash.String(), &operatorCfg.ChainId),
				utils.EmojiCheckMark,
			)

			logger.Infof("Operator updated successfully %s", utils.EmojiCheckMark)
			return nil
		},
	}

	return updateCmd
}
