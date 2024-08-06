package eigenpod

import (
	"context"
	"errors"
	"fmt"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/types"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	eigenpodproofs "github.com/Layr-Labs/eigenpod-proofs-generation"
	"github.com/Layr-Labs/eigenpod-proofs-generation/cli/core"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/eigenpod"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/eigenpod/bindings"
	"github.com/Layr-Labs/eigensdk-go/chainio/txmgr"
	"github.com/Layr-Labs/eigensdk-go/logging"
	eigenSdkUtils "github.com/Layr-Labs/eigensdk-go/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethcommon "github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli/v2"
	"math/big"
	"sort"
)

type checkpointConfig struct {
	network                    string
	ethClient                  *ethclient.Client
	beaconClient               core.BeaconClient
	podAddress                 gethcommon.Address
	outputType                 string
	outputFile                 string
	batchSize                  uint64
	proofPath                  string
	chainID                    *big.Int
	verbose                    bool
	forceCheckpoint            bool
	signerCfg                  *types.SignerConfig
	checkpointSubmittedAddress string
}

func CheckpointCmd(p utils.Prompter) *cli.Command {
	return &cli.Command{
		Name:  "checkpoint",
		Usage: "Checkpoint the current state of an EigenPod",
		After: telemetry.AfterRunAction(),
		Flags: getCheckpointCmdFlags(),
		Action: func(c *cli.Context) error {
			return checkpoint(c, p)
		},
	}
}

func getCheckpointCmdFlags() []cli.Flag {

	BatchSizeFlag.Value = DefaultBatchCheckpoint
	baseFlags := []cli.Flag{
		&flags.NetworkFlag,
		&flags.ETHRpcUrlFlag,
		&flags.BeaconRpcUrlFlag,
		&flags.OutputFileFlag,
		&flags.OutputTypeFlag,
		&flags.VerboseFlag,
		&PodAddressFlag,
		&BatchSizeFlag,
		&ProofPathFlag,
	}

	allFlags := append(baseFlags, flags.GetSignerFlags()...)
	sort.Sort(cli.FlagsByName(allFlags))

	return allFlags
}

func checkpoint(cCtx *cli.Context, p utils.Prompter) error {
	ctx := cCtx.Context
	logger := common.GetLogger(cCtx)

	cfg, err := readAndValidateCheckpointConfig(cCtx, logger)
	if err != nil {
		return err
	}
	cCtx.App.Metadata["network"] = cfg.chainID.String()
	if !common.IsEmptyString(cfg.proofPath) {
		if cfg.signerCfg == nil {
			return fmt.Errorf("signer config is required to submit checkpoint proof")
		}
		keyWallet, sender, err := common.GetWallet(
			*cfg.signerCfg,
			cfg.checkpointSubmittedAddress,
			cfg.ethClient,
			p,
			*cfg.chainID,
			logger,
		)
		if err != nil {
			return eigenSdkUtils.WrapError("failed to get wallet", err)
		}

		txMgr := txmgr.NewSimpleTxManager(keyWallet, cfg.ethClient, logger, sender)
		epWriter, err := eigenpod.NewWriter(cfg.podAddress, cfg.ethClient, txMgr, logger)
		if err != nil {
			return eigenSdkUtils.WrapError("unable to initialize eigenpod writer", err)
		}
		logger.Infof("Loading checkpoint proof from file: %s", cfg.proofPath)
		proof, err := core.LoadCheckpointProofFromFile(cfg.proofPath)
		if err != nil {
			return eigenSdkUtils.WrapError("unable to load proof from file", err)
		}

		receipts, err := submitCheckpointProof(ctx, cfg, epWriter, proof)
		for _, receipt := range receipts {
			fmt.Printf("%s Transaction Link: %s\n", utils.EmojiLink, common.GetTransactionLink(receipt.TxHash.String(), cfg.chainID))
		}
		if err != nil {
			return err
		}
		return nil
	}

	epReader, err := eigenpod.NewReader(cfg.podAddress, cfg.ethClient, logger)
	if err != nil {
		return eigenSdkUtils.WrapError("unable to build eigenpod reader", err)
	}

	currentCheckpointTimestamp, err := epReader.CurrentCheckpointTimestamp(&bind.CallOpts{})
	if err != nil {
		return eigenSdkUtils.WrapError("unable to get current checkpoint", err)
	}

	if currentCheckpointTimestamp == 0 {
		if cfg.signerCfg != nil {
			// consent stuff

		} else {
			return errors.New("no checkpoint active of no signer provided")
		}
	}

}

func startCheckpoint(
	ctx context.Context,
	cfg *checkpointConfig,
	epReader *eigenpod.ChainReader,
	p utils.Prompter,
	logger logging.Logger,
) (*gethtypes.Receipt, error) {

	keyWallet, sender, err := common.GetWallet(
		*cfg.signerCfg,
		cfg.checkpointSubmittedAddress,
		cfg.ethClient,
		p,
		*cfg.chainID,
		logger,
	)
	if err != nil {
		return nil, eigenSdkUtils.WrapError("failed to get wallet", err)
	}

	txMgr := txmgr.NewSimpleTxManager(keyWallet, cfg.ethClient, logger, sender)
	epWriter, err := eigenpod.NewWriter(cfg.podAddress, cfg.ethClient, txMgr, logger)
	if err != nil {
		return nil, eigenSdkUtils.WrapError("unable to initialize eigenpod writer", err)
	}

	receipt, err := epWriter.StartCheckpoint(ctx, !cfg.forceCheckpoint)
	if err != nil {
		return nil, eigenSdkUtils.WrapError("unable to start checkpoint", err)
	}

	return proof, nil
}

func submitCheckpointProof(ctx context.Context, cfg *checkpointConfig, epWriter *eigenpod.ChainWriter, proof *eigenpodproofs.VerifyCheckpointProofsCallParams) (gethtypes.Receipts, error) {
	allProofChunks := chunk(proof.BalanceProofs, cfg.batchSize)
	var receipts gethtypes.Receipts
	for i := 0; i < len(allProofChunks); i++ {
		balanceProofs := allProofChunks[i]
		receipt, err := epWriter.VerifyCheckpointProofs(
			ctx,
			bindings.BeaconChainProofsBalanceContainerProof{
				BalanceContainerRoot: proof.ValidatorBalancesRootProof.ValidatorBalancesRoot,
				Proof:                proof.ValidatorBalancesRootProof.Proof.ToByteSlice(),
			},
			castBalanceProofs(balanceProofs),
		)
		if err != nil {
			// failed to submit batch.
			return receipts, err
		}
		receipts = append(receipts, receipt)
		fmt.Printf("Submitted chunk %d/%d -- waiting for transaction...: ", i+1, len(allProofChunks))
	}

	return receipts, nil
}

func castBalanceProofs(proofs []*eigenpodproofs.BalanceProof) []bindings.BeaconChainProofsBalanceProof {
	var out []bindings.BeaconChainProofsBalanceProof

	for i := 0; i < len(proofs); i++ {
		proof := proofs[i]
		out = append(out, bindings.BeaconChainProofsBalanceProof{
			PubkeyHash:  proof.PubkeyHash,
			BalanceRoot: proof.BalanceRoot,
			Proof:       proof.Proof.ToByteSlice(),
		})
	}

	return out
}

func chunk[T any](arr []T, chunkSize uint64) [][]T {
	// Validate the chunkSize to ensure it's positive
	if chunkSize <= 0 {
		panic("chunkSize must be greater than 0")
	}

	// Create a slice to hold the chunks
	var chunks [][]T

	// Loop through the input slice and create chunks
	arrLen := uint64(len(arr))
	for i := uint64(0); i < arrLen; i += chunkSize {
		end := uint64(i + chunkSize)
		if end > arrLen {
			end = arrLen
		}
		chunks = append(chunks, arr[i:end])
	}

	return chunks
}

func readAndValidateCheckpointConfig(c *cli.Context, logger logging.Logger) (*checkpointConfig, error) {
	network := c.String(flags.NetworkFlag.Name)
	logger.Debugf("Using Network: %s", network)

	verbose := c.Bool(flags.VerboseFlag.Name)
	chainID := utils.NetworkNameToChainId(network)
	batchSize := c.Uint64(BatchSizeFlag.Name)
	proofPath := c.String(ProofPathFlag.Name)
	forceCheckpoint := c.Bool(ForceCheckpointFlag.Name)
	checkpointSubmitterAddress := c.String(CheckpointSubmitterAddressFlag.Name)

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

	signerCfg, err := common.GetSignerConfig(c, logger)
	if err != nil {
		// We don't want to throw error since people can still use it to generate the claim
		// without broadcasting it
		logger.Debugf("Failed to get signer config: %s", err)
	}

	config := &checkpointConfig{
		network:                    network,
		ethClient:                  ethRpcClient,
		beaconClient:               beaconClient,
		podAddress:                 gethcommon.HexToAddress(podAddress),
		outputType:                 outputType,
		outputFile:                 outputFile,
		batchSize:                  batchSize,
		proofPath:                  proofPath,
		chainID:                    chainID,
		verbose:                    verbose,
		forceCheckpoint:            forceCheckpoint,
		signerCfg:                  signerCfg,
		checkpointSubmittedAddress: checkpointSubmitterAddress,
	}

	return config, nil
}
