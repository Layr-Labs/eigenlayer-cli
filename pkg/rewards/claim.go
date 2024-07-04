package rewards

import (
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/common/flags"
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
	SignerConfig              types.SignerConfig
}

func ClaimCmd(p utils.Prompter) *cli.Command {
	var claimCmd = &cli.Command{
		Name:  "claim",
		Usage: "Claim rewards for the operator",
		Action: func(cCtx *cli.Context) error {
			return Claim(cCtx, p)
		},
		Flags: []cli.Flag{
			&flags.NetworkFlag,
			&flags.ETHRpcUrlFlag,
			&flags.EarnerAddressFlag,
			&flags.OutputFileFlag,
			&flags.PathToKeyStoreFlag,
			&flags.EcdsaPrivateKeyFlag,
			&flags.BroadcastFlag,
			&RecipientAddressFlag,
			&TokenAddressesFlag,
			&RewardsCoordinatorAddressFlag,
			&ClaimTimestampFlag,
			&ProofStoreBaseURLFlag,
		},
	}

	return claimCmd
}

func Claim(cCtx *cli.Context, p utils.Prompter) error {
	ctx := cCtx.Context
	config, err := readAndValidateClaimConfig(cCtx)
	if err != nil {
		return err
	}

	ethClient, err := eth.NewClient(config.RPCUrl)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to create new eth client", err)
	}

	logger := logging.NewTextSLogger(os.Stdout, &logging.SLoggerOptions{})

	elReader, err := elcontracts.NewReaderFromConfig(
		elcontracts.Config{
			RewardsCoordinatorAddress: config.RewardsCoordinatorAddress,
		},
		ethClient, logger,
	)
	if err != nil {
		return err
	}

	//fmt.Println(config)
	df := httpProofDataFetcher.NewHttpProofDataFetcher(
		config.ProofStoreBaseURL,
		config.Environment,
		config.Network,
		http.DefaultClient,
	)

	latestSubmittedTimestamp, err := elReader.CurrRewardsCalculationEndTimestamp(&bind.CallOpts{})
	if err != nil {
		return err
	}
	claimDate := time.Unix(int64(latestSubmittedTimestamp), 0).UTC().Format(time.DateOnly)
	rootCount, err := elReader.GetDistributionRootsLength(&bind.CallOpts{})
	if err != nil {
		return err
	}

	rootIndex := uint32(rootCount.Uint64() - 1)

	fmt.Println(claimDate)
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
		return err
	}

	solidityClaim := claimgen.FormatProofForSolidity(accounts.Root(), claim)

	if config.Broadcast {
		keyWallet, sender, err := common.GetWallet(
			config.SignerConfig,
			config.EarnerAddress.String(),
			ethClient,
			p,
			*config.ChainID,
			logger,
		)
		if err != nil {
			return err
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
			return err
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
		_, err = eLWriter.ProcessClaim(ctx, elClaim, config.RecipientAddress)
		if err != nil {
			return eigenSdkUtils.WrapError("failed to process claim", err)
		}
	} else {
		// Write to file
		//err = utils.WriteToFile(config.Output, solidityClaim)
		fmt.Println(solidityClaim)
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
func readAndValidateClaimConfig(cCtx *cli.Context) (*ClaimConfig, error) {
	network := cCtx.String(flags.NetworkFlag.Name)
	rpcUrl := cCtx.String(flags.ETHRpcUrlFlag.Name)
	earnerAddress := gethcommon.HexToAddress(cCtx.String(flags.EarnerAddressFlag.Name))
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

	claimTimestamp := cCtx.String(ClaimTimestampFlag.Name)
	fmt.Println(claimTimestamp)
	if claimTimestamp != LatestClaimTimestamp {
		return nil, errors.New("claim-timestamp must be 'latest'")
	}

	recipientAddress := gethcommon.HexToAddress(cCtx.String(RecipientAddressFlag.Name))
	if recipientAddress == utils.ZeroAddress {
		fmt.Println("Recipient address not provided, using earner address as recipient address")
		recipientAddress = earnerAddress
	}
	chainID := utils.NetworkNameToChainId(network)

	proofStoreBaseURL := cCtx.String(ProofStoreBaseURLFlag.Name)

	// If empty get from utils
	if proofStoreBaseURL == "" {
		proofStoreBaseURL = utils.GetProofStoreBaseURL(network)

		// If still empty return error
		if proofStoreBaseURL == "" {
			return nil, errors.New("proof store base URL not provided")
		}
	}

	env := utils.GetEnvironmentFromNetwork(network)
	//env = "preprod"

	// Get SignerConfig
	signerConfig, err := common.GetSignerConfig(cCtx)
	if err != nil {
		return nil, err
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
		Environment:               env,
		RecipientAddress:          recipientAddress,
		SignerConfig:              *signerConfig,
	}, nil
}

func stringToAddressArray(addresses []string) []gethcommon.Address {
	var addressArray []gethcommon.Address
	for _, address := range addresses {
		addressArray = append(addressArray, gethcommon.HexToAddress(address))
	}
	return addressArray
}
