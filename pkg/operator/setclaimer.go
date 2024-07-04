package operator

import (
	"context"
	"fmt"
	"os"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/eth"
	"github.com/Layr-Labs/eigensdk-go/chainio/txmgr"
	eigensdkLogger "github.com/Layr-Labs/eigensdk-go/logging"
	eigenMetrics "github.com/Layr-Labs/eigensdk-go/metrics"

	gethcommon "github.com/ethereum/go-ethereum/common"

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

			operatorCfg, err := common.ValidateAndReturnConfig(configurationFilePath)
			if err != nil {
				return err
			}
			cCtx.App.Metadata["network"] = operatorCfg.ChainId.String()
			if operatorCfg.ChainId.Int64() == utils.MainnetChainId {
				return fmt.Errorf("set claimer currently unsupported on mainnet")
			}

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

			contractCfg := elcontracts.Config{
				DelegationManagerAddress:  gethcommon.HexToAddress(operatorCfg.ELDelegationManagerAddress),
				AvsDirectoryAddress:       gethcommon.HexToAddress(operatorCfg.ELAVSDirectoryAddress),
				RewardsCoordinatorAddress: gethcommon.HexToAddress(operatorCfg.ELRewardsCoordinatorAddress),
			}

			elWriter, err := elcontracts.NewWriterFromConfig(contractCfg, ethClient, logger, noopMetrics, txMgr)
			if err != nil {
				return err
			}

			receipt, err := elWriter.SetClaimerFor(context.Background(), gethcommon.HexToAddress(claimerAddress))
			if err != nil {
				return err
			}

			fmt.Printf(
				"%s Claimer address %s set successfully for operator %s\n",
				utils.EmojiCheckMark,
				claimerAddress,
				operatorCfg.Operator.Address,
			)

			common.PrintRegistrationInfo(
				receipt.TxHash.String(),
				gethcommon.HexToAddress(operatorCfg.Operator.Address),
				&operatorCfg.ChainId,
			)

			return nil
		},
	}

	return setClaimerCmd
}
