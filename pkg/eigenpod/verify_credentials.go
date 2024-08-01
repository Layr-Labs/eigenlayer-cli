package eigenpod

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sort"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/eigenpod/bindings"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/types"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/eigenpod"
	"github.com/Layr-Labs/eigensdk-go/chainio/txmgr"
	"github.com/Layr-Labs/eigensdk-go/logging"
	eigenSdkUtils "github.com/Layr-Labs/eigensdk-go/utils"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	eigenpodproofs "github.com/Layr-Labs/eigenpod-proofs-generation"
	"github.com/Layr-Labs/eigenpod-proofs-generation/cli/core"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

type verifyCredentialsConfig struct {
	network               string
	ethClient             *ethclient.Client
	beaconClient          core.BeaconClient
	podAddress            string
	batchSize             uint64
	outputType            string
	outputFile            string
	broadcast             bool
	chainID               *big.Int
	signerCfg             *types.SignerConfig
	proofSubmitterAddress string
}

func VerifyCredentialsCmd(p utils.Prompter) *cli.Command {
	return &cli.Command{
		Name:  "verify-credentials",
		Usage: "Verify the credentials of an EigenPod onchain",
		Action: func(c *cli.Context) error {
			return verifyCredentials(c, p)
		},
		Aliases: []string{"vc", "creds"},
		After:   telemetry.AfterRunAction(),
		Flags:   getVerifyCredentialsFlags(),
	}
}

func getVerifyCredentialsFlags() []cli.Flag {
	// Set default batch size for credentials to 60
	BatchSizeFlag.Value = 60
	baseFlags := []cli.Flag{
		&PodAddressFlag,
		&flags.NetworkFlag,
		&flags.ETHRpcUrlFlag,
		&flags.BeaconRpcUrlFlag,
		&flags.OutputFileFlag,
		&flags.VerboseFlag,
		&flags.BroadcastFlag,
		&BatchSizeFlag,
		&ProofSubmitterAddress,
	}

	allFlags := append(baseFlags, flags.GetSignerFlags()...)
	sort.Sort(cli.FlagsByName(allFlags))
	return allFlags
}

func verifyCredentials(cCtx *cli.Context, p utils.Prompter) error {
	ctx := cCtx.Context
	logger := common.GetLogger(cCtx)

	cfg, err := readAndValidateVerifyCredsConfig(cCtx, logger)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to read and validate verify credentials config", err)
	}

	validatorProofs, oracleBeaconTimestamp, err := core.GenerateValidatorProof(
		ctx,
		cfg.podAddress,
		cfg.ethClient,
		cfg.chainID,
		cfg.beaconClient,
	)
	if err != nil || validatorProofs == nil {
		return eigenSdkUtils.WrapError("failed to generate validator proof or no inactive validators", err)
	}

	if cfg.broadcast {
		// Broadcast the proofs to the network
		err = BroadcastValidatorProofs(ctx, p, cfg, validatorProofs, oracleBeaconTimestamp, logger)
		if err != nil {
			return eigenSdkUtils.WrapError("failed to broadcast validator proofs", err)
		}
	} else {
		data := map[string]any{
			"validatorProofs": validatorProofs,
		}
		out, err := json.MarshalIndent(data, "", "   ")
		if err != nil {
			return eigenSdkUtils.WrapError("failed to marshal validator proofs", err)
		}

		if cfg.outputFile != "" {
			err = common.WriteToJSON(out, cfg.outputFile)
			if err != nil {
				return eigenSdkUtils.WrapError("failed to write validator proofs to file", err)
			}
		} else {
			logger.Infof("Validator proofs: %s", string(out))
		}
	}

	return nil
}

func BroadcastValidatorProofs(
	ctx context.Context,
	p utils.Prompter,
	cfg *verifyCredentialsConfig,
	proofs *eigenpodproofs.VerifyValidatorFieldsCallParams,
	oracleBeaconTimestamp uint64,
	logger logging.Logger,
) error {
	if cfg.signerCfg == nil {
		return errors.New("signer is required for broadcasting")
	}
	logger.Info("Broadcasting claim...")
	keyWallet, sender, err := common.GetWallet(
		*cfg.signerCfg,
		cfg.proofSubmitterAddress,
		cfg.ethClient,
		p,
		*cfg.chainID,
		logger,
	)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to get wallet", err)
	}

	txMgr := txmgr.NewSimpleTxManager(keyWallet, cfg.ethClient, logger, sender)
	epWriter, err := eigenpod.NewWriter(gethcommon.HexToAddress(cfg.podAddress), cfg.ethClient, txMgr, logger)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to create eigenpod writer", err)
	}

	indices := Uint64ArrayToBigIntArray(proofs.ValidatorIndices)
	validatorIndicesChunks := chunk(indices, cfg.batchSize)
	validatorProofsChunks := chunk(proofs.ValidatorFieldsProofs, cfg.batchSize)
	validatorFieldsChunks := chunk(proofs.ValidatorFields, cfg.batchSize)

	// Do I need to add consent? THis is already happening in broadcast

	numChunks := len(validatorIndicesChunks)

	color.Green(
		"calling EigenPod.VerifyWithdrawalCredentials() (using %d txn(s), max(%d) proofs per txn)",
		numChunks,
		cfg.batchSize,
	)
	color.Green("Submitting proofs with %d transactions", numChunks)

	for i := 0; i < numChunks; i++ {
		curValidatorIndices := validatorIndicesChunks[i]
		curValidatorProofs := validatorProofsChunks[i]

		var validatorFieldsProofs [][]byte
		for i := 0; i < len(curValidatorProofs); i++ {
			pr := curValidatorProofs[i].ToByteSlice()
			validatorFieldsProofs = append(validatorFieldsProofs, pr)
		}
		curValidatorFields := castValidatorFields(validatorFieldsChunks[i])

		fmt.Printf("Submitted chunk %d/%d -- waiting for transaction...: \n", i+1, numChunks)
		receipt, err := epWriter.VerifyWithdrawalCredentials(
			ctx,
			oracleBeaconTimestamp,
			bindings.BeaconChainProofsStateRootProof{
				Proof:           proofs.StateRootProof.Proof.ToByteSlice(),
				BeaconStateRoot: proofs.StateRootProof.BeaconStateRoot,
			},
			curValidatorIndices,
			validatorFieldsProofs,
			curValidatorFields,
		)
		if err != nil {
			return eigenSdkUtils.WrapError("failed to submit validator proof chunk", err)
		}

		fmt.Printf(
			"%s Transaction Link: %s\n",
			utils.EmojiLink,
			common.GetTransactionLink(receipt.TxHash.String(), cfg.chainID),
		)
	}

	return nil
}

func castValidatorFields(proof [][]eigenpodproofs.Bytes32) [][][32]byte {
	result := make([][][32]byte, len(proof))

	for i, slice := range proof {
		result[i] = make([][32]byte, len(slice))
		for j, bytes := range slice {
			result[i][j] = bytes
		}
	}

	return result
}

func Uint64ArrayToBigIntArray(nums []uint64) []*big.Int {
	var out []*big.Int
	for i := 0; i < len(nums); i++ {
		bigInt := new(big.Int).SetUint64(nums[i])
		out = append(out, bigInt)
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
		end := i + chunkSize
		if end > arrLen {
			end = arrLen
		}
		chunks = append(chunks, arr[i:end])
	}

	return chunks
}

func readAndValidateVerifyCredsConfig(cCtx *cli.Context, logger logging.Logger) (*verifyCredentialsConfig, error) {
	network := cCtx.String("network")
	logger.Debugf("Using Network: %s", network)
	ethRpcUrl := cCtx.String(flags.ETHRpcUrlFlag.Name)
	ethRpcClient, err := ethclient.Dial(ethRpcUrl)
	if err != nil {
		return nil, err
	}

	beaconRpcUrl := cCtx.String(flags.BeaconRpcUrlFlag.Name)
	beaconClient, err := core.GetBeaconClient(beaconRpcUrl)
	if err != nil {
		return nil, err
	}

	podAddress := cCtx.String(PodAddressFlag.Name)
	logger.Debugf("Using Pod Address: %s", podAddress)

	// TODO(shrimalmadhur): Implement pretty version of output
	//outputType := cCtx.String(flags.OutputTypeFlag.Name)

	outputFile := cCtx.String(flags.OutputFileFlag.Name)

	batchSize := cCtx.Uint64(BatchSizeFlag.Name)
	logger.Debugf("Using Batch Size: %d", batchSize)

	broadcast := cCtx.Bool(flags.BroadcastFlag.Name)

	chainID := utils.NetworkNameToChainId(network)

	// Get SignerConfig
	signerConfig, err := common.GetSignerConfig(cCtx, logger)
	if err != nil {
		// We don't want to throw error since people can still use it to generate the claim
		// without broadcasting it
		logger.Debugf("Failed to get signer config: %s", err)
	}

	proofSubmittedAddress := ""
	if signerConfig != nil && signerConfig.SignerType == types.Web3Signer {
		proofSubmittedAddress = cCtx.String(ProofSubmitterAddress.Name)
	}

	return &verifyCredentialsConfig{
		network:      network,
		ethClient:    ethRpcClient,
		beaconClient: beaconClient,
		podAddress:   podAddress,
		batchSize:    batchSize,
		//outputType:   outputType,
		outputFile:            outputFile,
		broadcast:             broadcast,
		chainID:               chainID,
		signerCfg:             signerConfig,
		proofSubmitterAddress: proofSubmittedAddress,
	}, nil
}
