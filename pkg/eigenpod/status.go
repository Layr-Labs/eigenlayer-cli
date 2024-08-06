package eigenpod

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	"github.com/Layr-Labs/eigenpod-proofs-generation/cli/core"

	"github.com/Layr-Labs/eigensdk-go/logging"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

type statusConfig struct {
	network      string
	ethClient    *ethclient.Client
	beaconClient core.BeaconClient
	podAddress   string
	outputType   string
	outputFile   string
	chainID      *big.Int
}

func StatusCmd(p utils.Prompter) *cli.Command {
	return &cli.Command{
		Name:  "status",
		Usage: "Get the status of an EigenPod",
		Action: func(c *cli.Context) error {
			return status(c, p)
		},
		After: telemetry.AfterRunAction(),
		Flags: []cli.Flag{
			&flags.NetworkFlag,
			&flags.ETHRpcUrlFlag,
			&flags.BeaconRpcUrlFlag,
			&flags.OutputFileFlag,
			&flags.OutputTypeFlag,
			&PodAddressFlag,
		},
	}
}

func status(cCtx *cli.Context, p utils.Prompter) error {
	ctx := cCtx.Context
	logger := common.GetLogger(cCtx)

	cfg, err := readAndValidateConfig(cCtx, logger)
	if err != nil {
		return err
	}
	cCtx.App.Metadata["network"] = cfg.chainID.String()
	eigenPodStatus := core.GetStatus(ctx, cfg.podAddress, cfg.ethClient, cfg.beaconClient)
	if cfg.outputType == "json" {
		jsonData, err := json.MarshalIndent(eigenPodStatus, "", "  ")
		if err != nil {
			return err
		}

		if cfg.outputFile != "" {
			err := common.WriteToFile(jsonData, cfg.outputFile)
			if err != nil {
				return err
			}
		} else {
			fmt.Println()
			fmt.Println(string(jsonData))
		}
	} else {
		fmt.Println()
		color.Green("EigenPod Address: %s\n", cfg.podAddress)
		color.Green("EigenPod Proof Submitted Address: %s\n", eigenPodStatus.ProofSubmitter.String())
		fmt.Println()
		inactiveValidators, activeValidators, withdrawnValidators := core.SortByStatus(eigenPodStatus.Validators)
		if len(inactiveValidators) > 0 {
			color.Yellow("Inactive Validators. Run `credentials` to verify these %d validators' withdrawal credentials \n", len(inactiveValidators))
			prettyPrintValidator(inactiveValidators)
			fmt.Println()
		}

		if len(activeValidators) > 0 {
			color.Green("Active Validators. Run `checkpoint` to update these %d validators' balances \n", len(activeValidators))
			prettyPrintValidator(activeValidators)
			fmt.Println()
		}

		if len(withdrawnValidators) > 0 {
			color.Red("Withdrawn Validators \n")
			prettyPrintValidator(withdrawnValidators)
			fmt.Println()
		}

		// Calculate the change in shares for completing a checkpoint
		deltaETH := new(big.Float).Sub(
			eigenPodStatus.TotalSharesAfterCheckpointETH,
			eigenPodStatus.CurrentTotalSharesETH,
		)

		if eigenPodStatus.ActiveCheckpoint != nil {
			startTime := time.Unix(int64(eigenPodStatus.ActiveCheckpoint.StartedAt), 0)

			color.Blue("!NOTE: There is a checkpoint active! (started at: %s)\n", startTime.String())
			color.Blue("\t- If you finish it, you may receive up to %f shares. (%f -> %f)\n", deltaETH, eigenPodStatus.CurrentTotalSharesETH, eigenPodStatus.TotalSharesAfterCheckpointETH)
			color.Blue("\t- %d proof(s) remaining until completion.\n", eigenPodStatus.ActiveCheckpoint.ProofsRemaining)
		} else {
			color.Blue("Running a `checkpoint` right now will result in: \n")
			color.Blue("\t%f new shares issued (%f ==> %f)\n", deltaETH, eigenPodStatus.CurrentTotalSharesETH, eigenPodStatus.TotalSharesAfterCheckpointETH)

			if eigenPodStatus.MustForceCheckpoint {
				color.Yellow("\tNote: pod does not have checkpointable native ETH. To checkpoint anyway, run `checkpoint` with the `--force` flag.\n")
			}

			color.Blue("Batching %d proofs per txn, this will require:\n\t", DefaultBatchCheckpoint)
			color.Blue("- 1x startCheckpoint() transaction, and \n\t- %dx EigenPod.verifyCheckpointProofs() transaction(s)\n\n", int(math.Ceil(float64(eigenPodStatus.NumberValidatorsToCheckpoint)/float64(DefaultBatchCheckpoint))))
		}
	}
	return nil
}

func prettyPrintValidator(validators []core.Validator) {
	column := formatStringColumns("Validator Index", 15) +
		" | " + formatStringColumns("Public Key", 98) +
		" | " + formatStringColumns("Effective Balance (GWei)", 24) +
		" | " + formatStringColumns("Current Balance (GWei)", 24) +
		" | " + formatStringColumns("Slashed", 8)

	fmt.Println(strings.Repeat("-", len(column)))
	fmt.Println(column)
	fmt.Println(strings.Repeat("-", len(column)))

	for _, validator := range validators {
		fmt.Printf(
			"%s | %s | %s | %s | %s\n",
			formatUint64Columns(validator.Index, 15),
			formatStringColumns(validator.PublicKey, 98),
			formatUint64Columns(validator.EffectiveBalance, 24),
			formatUint64Columns(validator.CurrentBalance, 24),
			formatBoolColumns(validator.Slashed, 8),
		)
	}
	fmt.Println(strings.Repeat("-", len(column)))
}

func formatBoolColumns(value bool, size int32) string {
	return fmt.Sprintf("%-*t", size, value)
}

func formatUint64Columns(value uint64, size int32) string {
	return fmt.Sprintf("%-*d", size, value)
}

func formatStringColumns(columnName string, size int32) string {
	return fmt.Sprintf("%-*s", size, columnName)
}

func readAndValidateConfig(c *cli.Context, logger logging.Logger) (*statusConfig, error) {
	network := c.String(flags.NetworkFlag.Name)
	logger.Debugf("Using Network: %s", network)

	chainID := utils.NetworkNameToChainId(network)

	ethRpcUrl := c.String(flags.ETHRpcUrlFlag.Name)
	ethRpcClient, err := ethclient.Dial(ethRpcUrl)
	if err != nil {
		return nil, err
	}

	beaconRpcUrl := c.String(flags.BeaconRpcUrlFlag.Name)
	beaconClient, err := core.GetBeaconClient(beaconRpcUrl)
	if err != nil {
		return nil, err
	}

	podAddress := c.String(PodAddressFlag.Name)
	logger.Debugf("Using Pod Address: %s", podAddress)

	outputType := c.String(flags.OutputTypeFlag.Name)
	outputFile := c.String(flags.OutputFileFlag.Name)

	config := &statusConfig{
		network:      network,
		ethClient:    ethRpcClient,
		beaconClient: beaconClient,
		podAddress:   podAddress,
		outputType:   outputType,
		outputFile:   outputFile,
		chainID:      chainID,
	}

	return config, nil
}
