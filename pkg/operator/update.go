package operator

import (
	"context"
	"fmt"

	"github.com/Layr-Labs/eigensdk-go/chainio/txmgr"
	"github.com/Layr-Labs/eigensdk-go/signerv2"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	elContracts "github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/eth"
	eigensdkLogger "github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/Layr-Labs/eigensdk-go/metrics"
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
			operatorCfg, err := validateAndMigrateConfigFile(configurationFilePath)
			if err != nil {
				return err
			}
			fmt.Printf(
				"%s Operator configuration file read successfully %s\n",
				utils.EmojiCheckMark,
				operatorCfg.Operator.Address,
			)

			logger, err := eigensdkLogger.NewZapLogger(eigensdkLogger.Development)
			if err != nil {
				return err
			}

			ethClient, err := eth.NewClient(operatorCfg.EthRPCUrl)
			if err != nil {
				return err
			}

			ecdsaPassword, err := p.InputHiddenString("Enter password to decrypt the ecdsa private key:", "",
				func(password string) error {
					return nil
				},
			)
			if err != nil {
				fmt.Printf("%s Error while reading ecdsa key password\n", utils.EmojiCrossMark)
				return err
			}

			signerCfg := signerv2.Config{
				KeystorePath: operatorCfg.PrivateKeyStorePath,
				Password:     ecdsaPassword,
			}
			sgn, sender, err := signerv2.SignerFromConfig(signerCfg, &operatorCfg.ChainId)
			if err != nil {
				return err
			}
			txMgr := txmgr.NewSimpleTxManager(ethClient, logger, sgn, sender)

			noopMetrics := metrics.NewNoopMetrics()

			elWriter, err := elContracts.BuildELChainWriter(
				common.HexToAddress(operatorCfg.ELDelegationManagerAddress),
				ethClient,
				logger,
				noopMetrics,
				txMgr)

			if err != nil {
				return err
			}

			receipt, err := elWriter.UpdateOperatorDetails(context.Background(), operatorCfg.Operator)
			if err != nil {
				fmt.Printf("%s Error while updating operator details\n", utils.EmojiCrossMark)
				return err
			}
			fmt.Printf(
				"%s Operator details updated at: %s\n",
				utils.EmojiCheckMark,
				getTransactionLink(receipt.TxHash.String(), &operatorCfg.ChainId),
			)

			fmt.Printf("%s Operator updated successfully\n", utils.EmojiCheckMark)
			return nil
		},
	}

	return updateCmd
}
