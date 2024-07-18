package rewards

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/types"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	contractrewardscoordinator "github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IRewardsCoordinator"

	"github.com/Layr-Labs/eigenlayer-rewards-proofs/pkg/claimgen"
	"github.com/Layr-Labs/eigenlayer-rewards-proofs/pkg/proofDataFetcher/httpProofDataFetcher"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/eth"
	"github.com/Layr-Labs/eigensdk-go/chainio/txmgr"
	rewardscoordinator "github.com/Layr-Labs/eigensdk-go/contracts/bindings/IRewardsCoordinator"
	"github.com/Layr-Labs/eigensdk-go/logging"
	eigenMetrics "github.com/Layr-Labs/eigensdk-go/metrics"
	eigenSdkUtils "github.com/Layr-Labs/eigensdk-go/utils"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/urfave/cli/v2"
)

const LatestClaimTimestamp = "latest"

type ClaimConfig struct {
	Network                   string
	RPCUrl                    string
	EarnerAddress             gethcommon.Address
	RecipientAddress          gethcommon.Address
	Output                    string
	Broadcast                 bool
	TokenAddresses            []gethcommon.Address
	RewardsCoordinatorAddress gethcommon.Address
	ClaimTimestamp            string
	ChainID                   *big.Int
	ProofStoreBaseURL         string
	Environment               string
	SignerConfig              *types.SignerConfig
}

func ClaimCmd(p utils.Prompter) *cli.Command {
	var claimCmd = &cli.Command{
		Name:  "claim",
		Usage: "Claim rewards for the operator",
		Action: func(cCtx *cli.Context) error {
			return Claim(cCtx, p)
		},
		After: telemetry.AfterRunAction(),
		Flags: []cli.Flag{
			&flags.NetworkFlag,
			&flags.ETHRpcUrlFlag,
			&flags.OutputFileFlag,
			&flags.BroadcastFlag,
			&EarnerAddressFlag,
			&EnvironmentFlag,
			&RecipientAddressFlag,
			&TokenAddressesFlag,
			&RewardsCoordinatorAddressFlag,
			&ClaimTimestampFlag,
			&ProofStoreBaseURLFlag,
			&flags.PathToKeyStoreFlag,
			&flags.EcdsaPrivateKeyFlag,
			&flags.FireblocksAPIKeyFlag,
			&flags.FireblocksSecretKeyFlag,
			&flags.FireblocksBaseUrlFlag,
			&flags.FireblocksVaultAccountNameFlag,
			&flags.FireblocksTimeoutFlag,
			&flags.FireblocksSecretStorageTypeFlag,
			&flags.FireblocksAWSRegionFlag,
			&flags.Web3SignerUrlFlag,
			&flags.VerboseFlag,
		},
	}

	return claimCmd
}

func Claim(cCtx *cli.Context, p utils.Prompter) error {
	ctx := cCtx.Context
	logger := common.GetLogger(cCtx)

	config, err := readAndValidateClaimConfig(cCtx, logger)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to read and validate claim config", err)
	}
	cCtx.App.Metadata["network"] = config.ChainID.String()
	if config.ChainID.Int64() == utils.MainnetChainId {
		return fmt.Errorf("rewards currently unsupported on mainnet")
	}

	ethClient, err := eth.NewClient(config.RPCUrl)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to create new eth client", err)
	}

	elReader, err := elcontracts.NewReaderFromConfig(
		elcontracts.Config{
			RewardsCoordinatorAddress: config.RewardsCoordinatorAddress,
		},
		ethClient, logger,
	)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to create new reader from config", err)
	}

	df := httpProofDataFetcher.NewHttpProofDataFetcher(
		config.ProofStoreBaseURL,
		config.Environment,
		config.Network,
		http.DefaultClient,
	)

	latestSubmittedTimestamp, err := elReader.CurrRewardsCalculationEndTimestamp(&bind.CallOpts{})
	if err != nil {
		return eigenSdkUtils.WrapError("failed to get latest submitted timestamp", err)
	}
	claimDate := time.Unix(int64(latestSubmittedTimestamp), 0).UTC().Format(time.DateOnly)
	logger.Debugf("Latest submitted timestamp: %s", claimDate)

	rootCount, err := elReader.GetDistributionRootsLength(&bind.CallOpts{})
	if err != nil {
		return eigenSdkUtils.WrapError("failed to get number of published roots", err)
	}

	rootIndex := uint32(rootCount.Uint64() - 1)

	proofData, err := df.FetchClaimAmountsForDate(ctx, claimDate)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to fetch claim amounts for date", err)
	}

	cg := claimgen.NewClaimgen(proofData.Distribution)
	accounts, claim, err := cg.GenerateClaimProofForEarner(
		config.EarnerAddress,
		config.TokenAddresses,
		rootIndex,
	)

	if err != nil {
		return eigenSdkUtils.WrapError("failed to generate claim proof for earner", err)
	}

	if config.Broadcast {
		if config.SignerConfig == nil {
			return errors.New("signer is required for broadcasting")
		}
		logger.Info("Broadcasting claim...")
		keyWallet, sender, err := common.GetWallet(
			*config.SignerConfig,
			config.EarnerAddress.String(),
			ethClient,
			p,
			*config.ChainID,
			logger,
		)
		if err != nil {
			return eigenSdkUtils.WrapError("failed to get wallet", err)
		}

		txMgr := txmgr.NewSimpleTxManager(keyWallet, ethClient, logger, sender)
		noopMetrics := eigenMetrics.NewNoopMetrics()
		eLWriter, err := elcontracts.NewWriterFromConfig(
			elcontracts.Config{
				RewardsCoordinatorAddress: config.RewardsCoordinatorAddress,
			},
			ethClient,
			logger,
			noopMetrics,
			txMgr,
		)
		if err != nil {
			return eigenSdkUtils.WrapError("failed to create new writer from config", err)
		}

		elClaim := rewardscoordinator.IRewardsCoordinatorRewardsMerkleClaim{
			RootIndex:       claim.RootIndex,
			EarnerIndex:     claim.EarnerIndex,
			EarnerTreeProof: claim.EarnerTreeProof,
			EarnerLeaf: rewardscoordinator.IRewardsCoordinatorEarnerTreeMerkleLeaf{
				Earner:          claim.EarnerLeaf.Earner,
				EarnerTokenRoot: claim.EarnerLeaf.EarnerTokenRoot,
			},
			TokenIndices:    claim.TokenIndices,
			TokenTreeProofs: claim.TokenTreeProofs,
			TokenLeaves:     convertClaimTokenLeaves(claim.TokenLeaves),
		}
		receipt, err := eLWriter.ProcessClaim(ctx, elClaim, config.RecipientAddress)
		if err != nil {
			return eigenSdkUtils.WrapError("failed to process claim", err)
		}

		logger.Infof("Claim transaction submitted successfully")
		common.PrintTransactionInfo(receipt.TxHash.String(), config.ChainID)
	} else {
		solidityClaim := claimgen.FormatProofForSolidity(accounts.Root(), claim)
		if !common.IsEmptyString(config.Output) {
			jsonData, err := json.MarshalIndent(solidityClaim, "", "  ")
			if err != nil {
				fmt.Println("Error marshaling JSON:", err)
				return err
			}

			err = common.WriteToJSON(jsonData, config.Output)
			if err != nil {
				return err
			}
			logger.Infof("Claim written to file: %s", config.Output)
		} else {
			fmt.Println("------- Claim generated -------")
			common.PrettyPrintStruct(*solidityClaim)
			fmt.Println("-------------------------------")
			fmt.Println("To write to a file, use the --output flag")
		}
		fmt.Println("To broadcast the claim, use the --broadcast flag")
	}

	return nil
}

func convertClaimTokenLeaves(
	claimTokenLeaves []contractrewardscoordinator.IRewardsCoordinatorTokenTreeMerkleLeaf,
) []rewardscoordinator.IRewardsCoordinatorTokenTreeMerkleLeaf {
	var tokenLeaves []rewardscoordinator.IRewardsCoordinatorTokenTreeMerkleLeaf
	for _, claimTokenLeaf := range claimTokenLeaves {
		tokenLeaves = append(tokenLeaves, rewardscoordinator.IRewardsCoordinatorTokenTreeMerkleLeaf{
			Token:              claimTokenLeaf.Token,
			CumulativeEarnings: claimTokenLeaf.CumulativeEarnings,
		})
	}
	return tokenLeaves

}
func readAndValidateClaimConfig(cCtx *cli.Context, logger logging.Logger) (*ClaimConfig, error) {
	network := cCtx.String(flags.NetworkFlag.Name)
	environment := cCtx.String(EnvironmentFlag.Name)
	rpcUrl := cCtx.String(flags.ETHRpcUrlFlag.Name)
	earnerAddress := gethcommon.HexToAddress(cCtx.String(EarnerAddressFlag.Name))
	output := cCtx.String(flags.OutputFileFlag.Name)
	broadcast := cCtx.Bool(flags.BroadcastFlag.Name)
	tokenAddresses := cCtx.String(TokenAddressesFlag.Name)
	tokenAddressArray := stringToAddressArray(strings.Split(tokenAddresses, ","))
	rewardsCoordinatorAddress := cCtx.String(RewardsCoordinatorAddressFlag.Name)
	var err error
	if rewardsCoordinatorAddress == "" {
		rewardsCoordinatorAddress, err = utils.GetRewardCoordinatorAddress(utils.NetworkNameToChainId(network))
		if err != nil {
			return nil, err
		}
	}
	logger.Debugf("Using Rewards Coordinator address: %s", rewardsCoordinatorAddress)

	claimTimestamp := cCtx.String(ClaimTimestampFlag.Name)
	if claimTimestamp != LatestClaimTimestamp {
		return nil, errors.New("claim-timestamp must be 'latest'")
	}

	recipientAddress := gethcommon.HexToAddress(cCtx.String(RecipientAddressFlag.Name))
	if recipientAddress == utils.ZeroAddress {
		logger.Infof(
			"Recipient address not provided, using earner address (%s) as recipient address",
			earnerAddress.String(),
		)
		recipientAddress = earnerAddress
	}
	logger.Infof("Using rewards recipient address: %s", recipientAddress.String())

	chainID := utils.NetworkNameToChainId(network)
	logger.Debugf("Using chain ID: %s", chainID.String())

	proofStoreBaseURL := cCtx.String(ProofStoreBaseURLFlag.Name)

	// If empty get from utils
	if proofStoreBaseURL == "" {
		proofStoreBaseURL = utils.GetProofStoreBaseURL(network)

		// If still empty return error
		if proofStoreBaseURL == "" {
			return nil, errors.New("proof store base URL not provided")
		}
	}
	logger.Debugf("Using Proof store base URL: %s", proofStoreBaseURL)

	if environment == "" {
		environment = getEnvFromNetwork(network)
	}
	logger.Debugf("Using network %s and environment: %s", network, environment)

	// Get SignerConfig
	signerConfig, err := common.GetSignerConfig(cCtx, logger)
	if err != nil {
		// We don't want to throw error since people can still use it to generate the claim
		// without broadcasting it
		logger.Debugf("Failed to get signer config: %s", err)
	}

	return &ClaimConfig{
		Network:                   network,
		RPCUrl:                    rpcUrl,
		EarnerAddress:             earnerAddress,
		Output:                    output,
		Broadcast:                 broadcast,
		TokenAddresses:            tokenAddressArray,
		RewardsCoordinatorAddress: gethcommon.HexToAddress(rewardsCoordinatorAddress),
		ChainID:                   chainID,
		ProofStoreBaseURL:         proofStoreBaseURL,
		Environment:               environment,
		RecipientAddress:          recipientAddress,
		SignerConfig:              signerConfig,
	}, nil
}

func getEnvFromNetwork(network string) string {
	switch network {
	case utils.HoleskyNetworkName:
		return "testnet"
	case utils.MainnetNetworkName:
		return "prod"
	default:
		return "local"
	}
}

func stringToAddressArray(addresses []string) []gethcommon.Address {
	var addressArray []gethcommon.Address
	for _, address := range addresses {
		addressArray = append(addressArray, gethcommon.HexToAddress(address))
	}
	return addressArray
}
