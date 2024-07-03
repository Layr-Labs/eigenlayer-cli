package rewards

import (
	"errors"
	"fmt"
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
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

/**
type ClaimConfig struct {
	GlobalConfig
	Environment               Environment `mapstructure:"environment"`
	Network                   string      `mapstructure:"network"`
	RPCUrl                    string      `mapstructure:"rpc_url"`
	PrivateKey                string      `mapstructure:"private_key"`
	RewardsCoordinatorAddress string      `mapstructure:"rewards_coordinator_address"`
	Output                    string      `mapstructure:"output"`
	EarnerAddress             string      `mapstructure:"earner_address"`
	Tokens                    []string    `mapstructure:"tokens"`
	ProofStoreBaseUrl         string      `mapstructure:"proof_store_base_url"`
	ClaimTimestamp            string      `mapstructure:"claim_timestamp"`
	SubmitClaim               bool        `mapstructure:"submit_claim"`
}
*/

const LatestClaimTimestamp = "latest"

type ClaimConfig struct {
	Network                   string
	RPCUrl                    string
	EarnerAddress             common.Address
	RecipientAddress          common.Address
	Output                    string
	PathToKeyStore            string
	Broadcast                 bool
	TokenAddresses            []common.Address
	RewardsCoordinatorAddress common.Address
	ClaimTimestamp            string
	ChainID                   *big.Int
}

func ClaimCmd(p utils.Prompter) cli.Command {
	var claimCmd = cli.Command{
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
			&flags.BroadcastFlag,
			&TokenAddressesFlag,
			&RewardsCoordinatorAddressFlag,
			&ClaimTimestampFlag,
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
	logger, err := logging.NewZapLogger(logging.Development)

	elReader, err := elcontracts.NewReaderFromConfig(
		elcontracts.Config{
			RewardsCoordinatorAddress: config.RewardsCoordinatorAddress,
		},
		ethClient, logger,
	)
	if err != nil {
		return err
	}

	// TODO(shrimalmadhur): to change this from config/network
	df := httpProofDataFetcher.NewHttpProofDataFetcher("", "preprod", config.Network, http.DefaultClient)

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

	proofData, err := df.FetchClaimAmountsForDate(ctx, claimDate)
	if err != nil {
		return err
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
		signerSfg := types.SignerConfig{PrivateKeyStorePath: config.PathToKeyStore}
		keyWallet, sender, err := getWallet(
			signerSfg,
			config.EarnerAddress,
			ethClient,
			p,
			config.ChainID,
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
		_, err = eLWriter.ProcessClaim(ctx, elClaim, config.EarnerAddress)
	} else {
		// Write to file
		//err = utils.WriteToFile(config.Output, solidityClaim)
		fmt.Println(solidityClaim)
	}

	return nil
}

func convertClaimTokenLeaves(claimTokenLeaves []contractrewardscoordinator.IRewardsCoordinatorTokenTreeMerkleLeaf) []rewardscoordinator.IRewardsCoordinatorTokenTreeMerkleLeaf {
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
	earnerAddress := common.HexToAddress(cCtx.String(flags.EarnerAddressFlag.Name))
	output := cCtx.String(flags.OutputFileFlag.Name)
	pathToKeyStore := cCtx.String(flags.PathToKeyStoreFlag.Name)
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
	if claimTimestamp != LatestClaimTimestamp {
		return nil, errors.New("claim-timestamp must be 'latest'")
	}

	recipientAddress := common.HexToAddress(cCtx.String(RecipientAddressFlag.Name))
	if recipientAddress == utils.ZeroAddress {
		fmt.Println("Recipient address not provided, using earner address as recipient address")
		recipientAddress = earnerAddress
	}

	chainID := utils.NetworkNameToChainId(network)

	return &ClaimConfig{
		Network:                   network,
		RPCUrl:                    rpcUrl,
		EarnerAddress:             earnerAddress,
		Output:                    output,
		PathToKeyStore:            pathToKeyStore,
		Broadcast:                 broadcast,
		TokenAddresses:            tokenAddressArray,
		RewardsCoordinatorAddress: common.HexToAddress(rewardsCoordinatorAddress),
		ChainID:                   chainID,
	}, nil
}

func stringToAddressArray(addresses []string) []common.Address {
	var addressArray []common.Address
	for _, address := range addresses {
		addressArray = append(addressArray, common.HexToAddress(address))
	}
	return addressArray
}
