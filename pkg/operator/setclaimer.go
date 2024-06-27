package operator

import (
	"context"
	"fmt"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/eth"
	"github.com/Layr-Labs/eigensdk-go/chainio/txmgr"
	eigensdkLogger "github.com/Layr-Labs/eigensdk-go/logging"
	eigenMetrics "github.com/Layr-Labs/eigensdk-go/metrics"

	"github.com/ethereum/go-ethereum/common"

	"github.com/urfave/cli/v2"
)

func SetClaimerCmd(p utils.Prompter) *cli.Command {
	setClaimerCmd := &cli.Command{
		Name:      "set-claimer",
		Usage:     "Set the claimer address for the operator",
		UsageText: "set-claimer",
		Description: `
		Set the rewards claimer address for the operator.
		`,
		After: telemetry.AfterRunAction(),
		Flags: []cli.Flag{
			&ConfigurationFilePathFlag,
			&ClaimerAddressFlag,
		},
		Action: func(cCtx *cli.Context) error {
			configurationFilePath := cCtx.String(ConfigurationFilePathFlag.Name)
			claimerAddress := cCtx.String(ClaimerAddressFlag.Name)

			operatorCfg, err := validateAndReturnConfig(configurationFilePath)
			if err != nil {
				return err
			}
			cCtx.App.Metadata["network"] = operatorCfg.ChainId.String()

			logger, err := eigensdkLogger.NewZapLogger(eigensdkLogger.Development)
			if err != nil {
				return err
			}

			ethClient, err := eth.NewClient(operatorCfg.EthRPCUrl)
			if err != nil {
				return err
			}

			keyWallet, sender, err := getWallet(operatorCfg, ethClient, p, logger)
			if err != nil {
				return err
			}

			txMgr := txmgr.NewSimpleTxManager(keyWallet, ethClient, logger, sender)

			noopMetrics := eigenMetrics.NewNoopMetrics()

			contractCfg := elcontracts.Config{
				DelegationManagerAddress:  common.HexToAddress(operatorCfg.ELDelegationManagerAddress),
				AvsDirectoryAddress:       common.HexToAddress(operatorCfg.ELAVSDirectoryAddress),
				RewardsCoordinatorAddress: common.HexToAddress(operatorCfg.ELRewardsCoordinatorAddress),
			}
			fmt.Println(operatorCfg)

			elWriter, err := elcontracts.NewWriterFromConfig(contractCfg, ethClient, logger, noopMetrics, txMgr)
			if err != nil {
				return err
			}

			receipt, err := elWriter.SetClaimerFor(context.Background(), common.HexToAddress(claimerAddress))
			if err != nil {
				return err
			}

			fmt.Printf(
				"%s Claimer address %s set successfully for operator %s\n",
				utils.EmojiCheckMark,
				claimerAddress,
				operatorCfg.Operator.Address,
			)

			printRegistrationInfo(
				receipt.TxHash.String(),
				common.HexToAddress(operatorCfg.Operator.Address),
				&operatorCfg.ChainId,
			)

			return nil
		},
	}

	return setClaimerCmd
}
