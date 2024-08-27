package rewards

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	contractrewardscoordinator "github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IRewardsCoordinator"

	"github.com/Layr-Labs/eigenlayer-rewards-proofs/pkg/claimgen"
	"github.com/Layr-Labs/eigenlayer-rewards-proofs/pkg/proofDataFetcher/httpProofDataFetcher"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	rewardscoordinator "github.com/Layr-Labs/eigensdk-go/contracts/bindings/IRewardsCoordinator"
	"github.com/Layr-Labs/eigensdk-go/logging"
	eigenSdkUtils "github.com/Layr-Labs/eigensdk-go/utils"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/urfave/cli/v2"
)

type elChainReader interface {
	GetDistributionRootsLength(opts *bind.CallOpts) (*big.Int, error)
	GetRootIndexFromHash(opts *bind.CallOpts, hash [32]byte) (uint32, error)
	GetCurrentClaimableDistributionRoot(
		opts *bind.CallOpts,
	) (rewardscoordinator.IRewardsCoordinatorDistributionRoot, error)
	CurrRewardsCalculationEndTimestamp(opts *bind.CallOpts) (uint32, error)
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
	}

	allFlags := append(baseFlags, flags.GetSignerFlags()...)
	sort.Sort(cli.FlagsByName(allFlags))
	return allFlags
}

func Claim(cCtx *cli.Context, p utils.Prompter) error {
	ctx := cCtx.Context
	logger := common.GetLogger(cCtx)

	config, err := readAndValidateClaimConfig(cCtx, logger)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to read and validate claim config", err)
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

	claimDate, rootIndex, err := getClaimDistributionRoot(ctx, config.ClaimTimestamp, elReader, logger)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to get claim distribution root", err)
	}

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
		receipt, err := eLWriter.ProcessClaim(ctx, elClaim, config.RecipientAddress, true)
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

		unsignedTx, err := contractBindings.RewardsCoordinator.ProcessClaim(noSendTxOpts, elClaim, config.RecipientAddress)
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
			solidityClaim := claimgen.FormatProofForSolidity(accounts.Root(), claim)
			jsonData, err := json.MarshalIndent(solidityClaim, "", "  ")
			if err != nil {
				fmt.Println("Error marshaling JSON:", err)
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
		} else {
			if !common.IsEmptyString(config.Output) {
				fmt.Println("output file not supported for pretty output type")
				fmt.Println()
			}
			solidityClaim := claimgen.FormatProofForSolidity(accounts.Root(), claim)
			fmt.Println("------- Claim generated -------")
			common.PrettyPrintStruct(*solidityClaim)
			fmt.Println("-------------------------------")
			fmt.Println("To write to a file, use the --output flag")
		}
		txFeeDetails := common.GetTxFeeDetails(unsignedTx)
		fmt.Println()
		txFeeDetails.Print()
		fmt.Println("To broadcast the claim, use the --broadcast flag")
	}

	return nil
}

func getClaimDistributionRoot(
	ctx context.Context,
	claimTimestamp string,
	elReader elChainReader,
	logger logging.Logger,
) (string, uint32, error) {
	if claimTimestamp == "latest" {
		latestSubmittedTimestamp, err := elReader.CurrRewardsCalculationEndTimestamp(&bind.CallOpts{Context: ctx})
		if err != nil {
			return "", 0, eigenSdkUtils.WrapError("failed to get latest submitted timestamp", err)
		}
		claimDate := time.Unix(int64(latestSubmittedTimestamp), 0).UTC().Format(time.DateOnly)

		rootCount, err := elReader.GetDistributionRootsLength(&bind.CallOpts{Context: ctx})
		if err != nil {
			return "", 0, eigenSdkUtils.WrapError("failed to get number of published roots", err)
		}

		rootIndex := uint32(rootCount.Uint64() - 1)
		logger.Debugf("Latest active rewards snapshot timestamp: %s, root index: %d", claimDate, rootIndex)
		return claimDate, rootIndex, nil
	} else if claimTimestamp == "latest_active" {
		latestClaimableRoot, err := elReader.GetCurrentClaimableDistributionRoot(&bind.CallOpts{Context: ctx})
		if err != nil {
			return "", 0, eigenSdkUtils.WrapError("failed to get latest claimable root", err)
		}

		rootIndex, err := elReader.GetRootIndexFromHash(&bind.CallOpts{Context: ctx}, latestClaimableRoot.Root)
		if err != nil {
			return "", 0, eigenSdkUtils.WrapError("failed to get root index from hash", err)
		}

		ts := time.Unix(int64(latestClaimableRoot.RewardsCalculationEndTimestamp), 0).UTC().Format(time.DateOnly)
		logger.Debugf("Latest rewards snapshot timestamp: %s, root index: %d", ts, rootIndex)

		return ts, rootIndex, nil
	}
	return "", 0, errors.New("invalid claim timestamp")
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
	tokenAddressArray := stringToAddressArray(strings.Split(tokenAddresses, ","))
	rewardsCoordinatorAddress := cCtx.String(RewardsCoordinatorAddressFlag.Name)

	var err error
	if common.IsEmptyString(rewardsCoordinatorAddress) {
		rewardsCoordinatorAddress, err = utils.GetRewardCoordinatorAddress(utils.NetworkNameToChainId(network))
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
		proofStoreBaseURL = utils.GetProofStoreBaseURL(network)

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
		TokenAddresses:            tokenAddressArray,
		RewardsCoordinatorAddress: gethcommon.HexToAddress(rewardsCoordinatorAddress),
		ChainID:                   chainID,
		ProofStoreBaseURL:         proofStoreBaseURL,
		Environment:               environment,
		RecipientAddress:          recipientAddress,
		SignerConfig:              signerConfig,
		ClaimTimestamp:            claimTimestamp,
		ClaimerAddress:            claimerAddress,
	}, nil
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

func stringToAddressArray(addresses []string) []gethcommon.Address {
	var addressArray []gethcommon.Address
	for _, address := range addresses {
		addressArray = append(addressArray, gethcommon.HexToAddress(address))
	}
	return addressArray
}
