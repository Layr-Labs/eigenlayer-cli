package rewards

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/wealdtech/go-merkletree/v2"
	"gopkg.in/yaml.v2"

	contractrewardscoordinator "github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IRewardsCoordinator"

	"github.com/Layr-Labs/eigenlayer-rewards-proofs/pkg/claimgen"
	"github.com/Layr-Labs/eigenlayer-rewards-proofs/pkg/distribution"
	"github.com/Layr-Labs/eigenlayer-rewards-proofs/pkg/proofDataFetcher"
	"github.com/Layr-Labs/eigenlayer-rewards-proofs/pkg/proofDataFetcher/httpProofDataFetcher"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	rewardscoordinator "github.com/Layr-Labs/eigensdk-go/contracts/bindings/IRewardsCoordinator"
	"github.com/Layr-Labs/eigensdk-go/logging"
	eigenSdkUtils "github.com/Layr-Labs/eigensdk-go/utils"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/urfave/cli/v2"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type elChainReader interface {
	GetDistributionRootsLength(ctx context.Context) (*big.Int, error)
	GetRootIndexFromHash(ctx context.Context, hash [32]byte) (uint32, error)
	GetCurrentClaimableDistributionRoot(
		ctx context.Context,
	) (rewardscoordinator.IRewardsCoordinatorDistributionRoot, error)
	CurrRewardsCalculationEndTimestamp(ctx context.Context) (uint32, error)
	GetCumulativeClaimed(ctx context.Context, earnerAddress, tokenAddress gethcommon.Address) (*big.Int, error)
}

func ClaimCmd(p utils.Prompter) *cli.Command {
	var claimCmd = &cli.Command{
		Name:  "claim",
		Usage: "Claim rewards for any earner",
		Action: func(cCtx *cli.Context) error {
			return Claim(cCtx, p)
		},
		After: telemetry.AfterRunAction(),
		Flags: getClaimFlags(),
	}

	return claimCmd
}

func getClaimFlags() []cli.Flag {
	baseFlags := []cli.Flag{
		&flags.NetworkFlag,
		&flags.ETHRpcUrlFlag,
		&flags.OutputFileFlag,
		&flags.OutputTypeFlag,
		&flags.BroadcastFlag,
		&EarnerAddressFlag,
		&EnvironmentFlag,
		&RecipientAddressFlag,
		&TokenAddressesFlag,
		&ClaimerAddressFlag,
		&RewardsCoordinatorAddressFlag,
		&ClaimTimestampFlag,
		&ProofStoreBaseURLFlag,
		&flags.VerboseFlag,
		&flags.SilentFlag,
		&flags.BatchClaimFile,
	}

	allFlags := append(baseFlags, flags.GetSignerFlags()...)
	sort.Sort(cli.FlagsByName(allFlags))
	return allFlags
}

func batchClaim(
	ctx context.Context,
	logger logging.Logger,
	ethClient *ethclient.Client,
	elReader *elcontracts.ChainReader,
	config *ClaimConfig,
	p utils.Prompter,
	rootIndex uint32,
	proofData *proofDataFetcher.RewardProofData,
) error {

	yamlFile, err := os.ReadFile(config.BatchClaimFile)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to read YAML config file", err)
	}

	var claimConfigs []struct {
		EarnerAddress  string   `yaml:"earner_address"`
		TokenAddresses []string `yaml:"token_addresses"`
	}

	err = yaml.Unmarshal(yamlFile, &claimConfigs)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to parse YAML config", err)
	}

	var elClaims []rewardscoordinator.IRewardsCoordinatorRewardsMerkleClaim
	var claims []contractrewardscoordinator.IRewardsCoordinatorRewardsMerkleClaim
	var accounts []merkletree.MerkleTree

	for _, claimConfig := range claimConfigs {
		earnerAddr := gethcommon.HexToAddress(claimConfig.EarnerAddress)

		var tokenAddrs []gethcommon.Address

		// Empty token addresses list will create a claim for all tokens claimable
		// by the earner address.
		if len(claimConfig.TokenAddresses) != 0 {
			for _, addr := range claimConfig.TokenAddresses {
				tokenAddrs = append(tokenAddrs, gethcommon.HexToAddress(addr))
			}
		}

		elClaim, claim, account, err := generateClaimPayload(
			ctx,
			rootIndex,
			proofData,
			elReader,
			logger,
			earnerAddr,
			tokenAddrs,
		)

		if err != nil {
			logger.Warnf("Failed to process claim for earner %s: %v", earnerAddr.String(), err)
			continue
		}

		elClaims = append(elClaims, *elClaim)
		claims = append(claims, *claim)
		accounts = append(accounts, *account)

	}

	return broadcastClaims(config, ethClient, logger, p, ctx, elClaims, claims, accounts)
}

func generateClaimPayload(
	ctx context.Context,
	rootIndex uint32,
	proofData *proofDataFetcher.RewardProofData,
	elReader *elcontracts.ChainReader,
	logger logging.Logger,
	earnerAddress gethcommon.Address,
	tokenAddresses []gethcommon.Address,
) (*rewardscoordinator.IRewardsCoordinatorRewardsMerkleClaim, *contractrewardscoordinator.IRewardsCoordinatorRewardsMerkleClaim, *merkletree.MerkleTree, error) {

	claimableTokensOrderMap, present := proofData.Distribution.GetTokensForEarner(earnerAddress)
	if !present {
		return nil, nil, nil, errors.New("no tokens claimable by earner")
	}

	claimableTokensMap := getTokensToClaim(claimableTokensOrderMap, tokenAddresses)

	claimableTokens, err := filterClaimableTokens(ctx, elReader, earnerAddress, claimableTokensMap)
	if err != nil {
		return nil, nil, nil, eigenSdkUtils.WrapError("failed to get claimable tokens", err)
	}

	cg := claimgen.NewClaimgen(proofData.Distribution)
	accounts, claim, err := cg.GenerateClaimProofForEarner(
		earnerAddress,
		claimableTokens,
		rootIndex,
	)
	if err != nil {
		return nil, nil, nil, eigenSdkUtils.WrapError("failed to generate claim proof for earner", err)
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

	logger.Infof("Validating claim proof for earner %s...", earnerAddress)
	ok, err := elReader.CheckClaim(ctx, elClaim)
	if err != nil {
		return nil, nil, nil, err
	}
	if !ok {
		return nil, nil, nil, errors.New("failed to validate claim")
	}
	logger.Infof("Claim proof for earner %s validated successfully", earnerAddress)

	return &elClaim, claim, accounts, nil
}

func Claim(cCtx *cli.Context, p utils.Prompter) error {
	ctx := cCtx.Context
	logger := common.GetLogger(cCtx)

	config, err := readAndValidateClaimConfig(cCtx, logger)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to read and validate claim config", err)
	}

	if config.Network != "holesky" && config.BatchClaimFile != "" {
		return errors.New("batch claim only supported on holesky with this version of CLI")
	}

	cCtx.App.Metadata["network"] = config.ChainID.String()

	ethClient, err := ethclient.Dial(config.RPCUrl)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to create new eth client", err)
	}

	elReader, err := elcontracts.NewReaderFromConfig(
		elcontracts.Config{
			RewardsCoordinatorAddress: config.RewardsCoordinatorAddress,
		},
		ethClient,
		logger,
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

	claimDate, rootIndex, err := getClaimDistributionRoot(ctx, config.ClaimTimestamp, elReader, logger)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to get claim distribution root", err)
	}

	proofData, err := df.FetchClaimAmountsForDate(ctx, claimDate)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to fetch claim amounts for date", err)
	}

	if config.BatchClaimFile != "" {
		return batchClaim(ctx, logger, ethClient, elReader, config, p, rootIndex, proofData)
	}

	elClaim, claim, account, err := generateClaimPayload(
		ctx,
		rootIndex,
		proofData,
		elReader,
		logger,
		config.EarnerAddress,
		config.TokenAddresses,
	)

	if err != nil {
		return err
	}

	elClaims := []rewardscoordinator.IRewardsCoordinatorRewardsMerkleClaim{*elClaim}
	claims := []contractrewardscoordinator.IRewardsCoordinatorRewardsMerkleClaim{*claim}
	accounts := []merkletree.MerkleTree{*account}
	err = broadcastClaims(config, ethClient, logger, p, ctx, elClaims, claims, accounts)

	return err
}

func broadcastClaims(
	config *ClaimConfig,
	ethClient *ethclient.Client,
	logger logging.Logger,
	p utils.Prompter,
	ctx context.Context,
	elClaims []rewardscoordinator.IRewardsCoordinatorRewardsMerkleClaim,
	claims []contractrewardscoordinator.IRewardsCoordinatorRewardsMerkleClaim,
	accounts []merkletree.MerkleTree,
) error {
	if len(elClaims) == 0 {
		return errors.New("at least one claim is required")
	}
	if config.Broadcast {
		eLWriter, err := common.GetELWriter(
			config.ClaimerAddress,
			config.SignerConfig,
			ethClient,
			elcontracts.Config{
				RewardsCoordinatorAddress: config.RewardsCoordinatorAddress,
			},
			p,
			config.ChainID,
			logger,
		)

		if err != nil {
			return eigenSdkUtils.WrapError("failed to get EL writer", err)
		}

		logger.Infof("Broadcasting claim transaction...")

		var receipt *types.Receipt

		if len(elClaims) > 1 {
			receipt, err = eLWriter.ProcessClaims(ctx, elClaims, config.RecipientAddress, true)
		} else {
			receipt, err = eLWriter.ProcessClaim(ctx, elClaims[0], config.RecipientAddress, true)
		}

		if err != nil {
			return eigenSdkUtils.WrapError("failed to process claim", err)
		}

		logger.Infof("Claim transaction submitted successfully")
		common.PrintTransactionInfo(receipt.TxHash.String(), config.ChainID)
	} else {
		noSendTxOpts := common.GetNoSendTxOpts(config.ClaimerAddress)
		_, _, contractBindings, err := elcontracts.BuildClients(elcontracts.Config{
			RewardsCoordinatorAddress: config.RewardsCoordinatorAddress,
		}, ethClient, nil, logger, nil)
		if err != nil {
			return err
		}

		// If claimer is a smart contract, we can't estimate gas using geth
		// since balance of contract can be 0, as it can be called by an EOA
		// to claim. So we hardcode the gas limit to 150_000 so that we can
		// create unsigned tx without gas limit estimation from contract bindings
		code, err := ethClient.CodeAt(ctx, config.ClaimerAddress, nil)
		if err != nil {
			return eigenSdkUtils.WrapError("failed to get code at address", err)
		}
		if len(code) > 0 {
			// Claimer is a smart contract
			noSendTxOpts.GasLimit = 150_000
		}
		var unsignedTx *types.Transaction
		if len(elClaims) > 1 {
			unsignedTx, err = contractBindings.RewardsCoordinator.ProcessClaims(noSendTxOpts, elClaims, config.RecipientAddress)
		} else {
			unsignedTx, err = contractBindings.RewardsCoordinator.ProcessClaim(noSendTxOpts, elClaims[0], config.RecipientAddress)
		}
		if err != nil {
			return eigenSdkUtils.WrapError("failed to create unsigned tx", err)
		}

		if config.OutputType == string(common.OutputType_Calldata) {
			calldataHex := gethcommon.Bytes2Hex(unsignedTx.Data())

			if !common.IsEmptyString(config.Output) {
				err = common.WriteToFile([]byte(calldataHex), config.Output)
				if err != nil {
					return err
				}
				logger.Infof("Call data written to file: %s", config.Output)
			} else {
				fmt.Println(calldataHex)
			}
		} else if config.OutputType == string(common.OutputType_Json) {
			for idx, claim := range claims {
				solidityClaim := claimgen.FormatProofForSolidity(accounts[idx].Root(), &claim)
				jsonData, err := json.MarshalIndent(solidityClaim, "", "  ")
				if err != nil {
					return err
				}
				if !common.IsEmptyString(config.Output) {
					err = common.WriteToFile(jsonData, config.Output)
					if err != nil {
						return err
					}
					logger.Infof("Claim written to file: %s", config.Output)
				} else {
					fmt.Println(string(jsonData))
					fmt.Println()
					fmt.Println("To write to a file, use the --output flag")
				}
			}
		} else {
			if !common.IsEmptyString(config.Output) {
				fmt.Println("output file not supported for pretty output type")
				fmt.Println()
			}
			for idx, claim := range claims {
				solidityClaim := claimgen.FormatProofForSolidity(accounts[idx].Root(), &claim)
				if !config.IsSilent {
					fmt.Println("------- Claim generated -------")
				}
				common.PrettyPrintStruct(*solidityClaim)
			}
			if !config.IsSilent {
				fmt.Println("-------------------------------")
				fmt.Println("To write to a file, use the --output flag")
			}
		}
		if !config.IsSilent {
			txFeeDetails := common.GetTxFeeDetails(unsignedTx)
			fmt.Println()
			txFeeDetails.Print()

			fmt.Println("To broadcast the claim, use the --broadcast flag")
		}
	}

	return nil
}

// filterClaimableTokens to filter out tokens that have been fully claimed
func filterClaimableTokens(
	ctx context.Context,
	elReader elChainReader,
	earnerAddress gethcommon.Address,
	claimableTokensMap map[gethcommon.Address]*big.Int,
) ([]gethcommon.Address, error) {
	claimableTokens := make([]gethcommon.Address, 0)
	for token, claimedAmount := range claimableTokensMap {
		amount, err := getCummulativeClaimedRewards(ctx, elReader, earnerAddress, token)
		if err != nil {
			return nil, err
		}
		// If the token has been claimed fully, we don't need to include it in the claim
		// This is because contracts reject claims for tokens that have been fully claimed
		// https://github.com/Layr-Labs/eigenlayer-contracts/blob/ac57bc1b28c83d9d7143c0da19167c148c3596a3/src/contracts/core/RewardsCoordinator.sol#L575-L578
		if claimedAmount.Cmp(amount) == 0 {
			continue
		}
		claimableTokens = append(claimableTokens, token)
	}
	return claimableTokens, nil
}

func getClaimDistributionRoot(
	ctx context.Context,
	claimTimestamp string,
	elReader elChainReader,
	logger logging.Logger,
) (string, uint32, error) {
	if claimTimestamp == LatestTimestamp {
		latestSubmittedTimestamp, err := elReader.CurrRewardsCalculationEndTimestamp(ctx)
		if err != nil {
			return "", 0, eigenSdkUtils.WrapError("failed to get latest submitted timestamp", err)
		}
		claimDate := time.Unix(int64(latestSubmittedTimestamp), 0).UTC().Format(time.DateOnly)

		rootCount, err := elReader.GetDistributionRootsLength(ctx)
		if err != nil {
			return "", 0, eigenSdkUtils.WrapError("failed to get number of published roots", err)
		}

		rootIndex := uint32(rootCount.Uint64() - 1)
		logger.Debugf("Latest active rewards snapshot timestamp: %s, root index: %d", claimDate, rootIndex)
		return claimDate, rootIndex, nil
	} else if claimTimestamp == LatestActiveTimestamp {
		latestClaimableRoot, err := elReader.GetCurrentClaimableDistributionRoot(ctx)
		if err != nil {
			return "", 0, eigenSdkUtils.WrapError("failed to get latest claimable root", err)
		}
		rootIndex, err := elReader.GetRootIndexFromHash(ctx, latestClaimableRoot.Root)
		if err != nil {
			return "", 0, eigenSdkUtils.WrapError("failed to get root index from hash", err)
		}

		ts := time.Unix(int64(latestClaimableRoot.RewardsCalculationEndTimestamp), 0).UTC().Format(time.DateOnly)
		logger.Debugf("Latest rewards snapshot timestamp: %s, root index: %d", ts, rootIndex)

		return ts, rootIndex, nil
	}
	return "", 0, errors.New("invalid claim timestamp")
}

func getTokensToClaim(
	claimableTokens *orderedmap.OrderedMap[gethcommon.Address, *distribution.BigInt],
	tokenAddresses []gethcommon.Address,
) map[gethcommon.Address]*big.Int {
	var tokenMap map[gethcommon.Address]*big.Int
	if len(tokenAddresses) == 0 {
		tokenMap = getAllClaimableTokenAddresses(claimableTokens)
	} else {
		tokenMap = filterClaimableTokenAddresses(claimableTokens, tokenAddresses)
	}

	return tokenMap
}

func getAllClaimableTokenAddresses(
	addressesMap *orderedmap.OrderedMap[gethcommon.Address, *distribution.BigInt],
) map[gethcommon.Address]*big.Int {
	tokens := make(map[gethcommon.Address]*big.Int)
	for pair := addressesMap.Oldest(); pair != nil; pair = pair.Next() {
		tokens[pair.Key] = pair.Value.Int
	}

	return tokens
}

func filterClaimableTokenAddresses(
	addressesMap *orderedmap.OrderedMap[gethcommon.Address, *distribution.BigInt],
	providedAddresses []gethcommon.Address,
) map[gethcommon.Address]*big.Int {
	tokens := make(map[gethcommon.Address]*big.Int)
	for _, address := range providedAddresses {
		if val, ok := addressesMap.Get(address); ok {
			tokens[address] = val.Int
		}
	}

	return tokens
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
	outputType := cCtx.String(flags.OutputTypeFlag.Name)
	broadcast := cCtx.Bool(flags.BroadcastFlag.Name)
	tokenAddresses := cCtx.String(TokenAddressesFlag.Name)
	splitTokenAddresses := strings.Split(tokenAddresses, ",")
	validTokenAddresses := getValidHexAddresses(splitTokenAddresses)
	rewardsCoordinatorAddress := cCtx.String(RewardsCoordinatorAddressFlag.Name)
	isSilent := cCtx.Bool(flags.SilentFlag.Name)
	batchClaimFile := cCtx.String(flags.BatchClaimFile.Name)

	var err error
	if common.IsEmptyString(rewardsCoordinatorAddress) {
		rewardsCoordinatorAddress, err = common.GetRewardCoordinatorAddress(utils.NetworkNameToChainId(network))
		if err != nil {
			return nil, err
		}
	}
	logger.Debugf("Using Rewards Coordinator address: %s", rewardsCoordinatorAddress)

	claimTimestamp := cCtx.String(ClaimTimestampFlag.Name)
	logger.Debugf("Using claim timestamp from user: %s", claimTimestamp)

	recipientAddress := gethcommon.HexToAddress(cCtx.String(RecipientAddressFlag.Name))
	if recipientAddress == utils.ZeroAddress {
		logger.Infof(
			"Recipient address not provided, using earner address (%s) as recipient address",
			earnerAddress.String(),
		)
		recipientAddress = earnerAddress
	}
	logger.Infof("Using rewards recipient address: %s", recipientAddress.String())

	claimerAddress := gethcommon.HexToAddress(cCtx.String(ClaimerAddressFlag.Name))
	if claimerAddress == utils.ZeroAddress {
		logger.Infof(
			"Claimer address not provided, using earner address (%s) as claimer address",
			earnerAddress.String(),
		)
		claimerAddress = earnerAddress
	}
	logger.Infof("Using rewards claimer address: %s", claimerAddress.String())

	chainID := utils.NetworkNameToChainId(network)
	logger.Debugf("Using chain ID: %s", chainID.String())

	proofStoreBaseURL := cCtx.String(ProofStoreBaseURLFlag.Name)

	// If empty get from utils
	if common.IsEmptyString(proofStoreBaseURL) {
		proofStoreBaseURL = getProofStoreBaseURL(network)

		// If still empty return error
		if common.IsEmptyString(proofStoreBaseURL) {
			return nil, errors.New("proof store base URL not provided")
		}
	}
	logger.Debugf("Using Proof store base URL: %s", proofStoreBaseURL)

	if common.IsEmptyString(environment) {
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

	// TODO(shrimalmadhur): Fix to make sure correct S3 bucket is used. Clean up later
	if network == utils.MainnetNetworkName {
		network = "ethereum"
	}

	return &ClaimConfig{
		Network:                   network,
		RPCUrl:                    rpcUrl,
		EarnerAddress:             earnerAddress,
		Output:                    output,
		OutputType:                outputType,
		Broadcast:                 broadcast,
		TokenAddresses:            validTokenAddresses,
		RewardsCoordinatorAddress: gethcommon.HexToAddress(rewardsCoordinatorAddress),
		ChainID:                   chainID,
		ProofStoreBaseURL:         proofStoreBaseURL,
		Environment:               environment,
		RecipientAddress:          recipientAddress,
		SignerConfig:              signerConfig,
		ClaimTimestamp:            claimTimestamp,
		ClaimerAddress:            claimerAddress,
		IsSilent:                  isSilent,
		BatchClaimFile:            batchClaimFile,
	}, nil
}

func getProofStoreBaseURL(network string) string {
	chainId := utils.NetworkNameToChainId(network)
	chainMetadata, ok := common.ChainMetadataMap[chainId.Int64()]
	if !ok {
		return ""
	} else {
		return chainMetadata.ProofStoreBaseURL
	}
}

func getEnvFromNetwork(network string) string {
	switch network {
	case utils.HoleskyNetworkName:
		return "testnet"
	case utils.MainnetNetworkName:
		return "mainnet"
	default:
		return "local"
	}
}

func getValidHexAddresses(addresses []string) []gethcommon.Address {
	var addressArray []gethcommon.Address
	for _, address := range addresses {
		if gethcommon.IsHexAddress(address) && address != utils.ZeroAddress.String() {
			addressArray = append(addressArray, gethcommon.HexToAddress(address))
		}
	}
	return addressArray
}
