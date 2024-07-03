package rewards

import (
	"errors"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/Layr-Labs/eigenlayer-rewards-proofs/pkg/claimgen"
	"github.com/Layr-Labs/eigenlayer-rewards-updater/pkg/proofDataFetcher/httpProofDataFetcher"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/eth"
	"github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"

	rConfig "github.com/Layr-Labs/eigenlayer-rewards-updater/pkg/config"
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
	Output                    string
	PathToKeyStore            string
	SubmitClaim               bool
	TokenAddresses            []common.Address
	RewardsCoordinatorAddress common.Address
	ClaimTimestamp            string
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
			&SubmitClaimFlag,
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
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}

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
	env := rConfig.Environment_PRE_PROD
	e, err := rConfig.StringEnvironmentFromEnum(env)
	if err != nil {
		return err
	}
	// TODO: data fetcher after Sean implements it
	df := httpProofDataFetcher.NewHttpProofDataFetcher("", e, config.Network, http.DefaultClient, zapLogger)

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

	if config.SubmitClaim {

	}
}

func readAndValidateClaimConfig(cCtx *cli.Context) (*ClaimConfig, error) {
	network := cCtx.String(flags.NetworkFlag.Name)
	rpcUrl := cCtx.String(flags.ETHRpcUrlFlag.Name)
	earnerAddress := common.HexToAddress(cCtx.String(flags.EarnerAddressFlag.Name))
	output := cCtx.String(flags.OutputFileFlag.Name)
	pathToKeyStore := cCtx.String(flags.PathToKeyStoreFlag.Name)
	submitClaim := cCtx.Bool(SubmitClaimFlag.Name)
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

	return &ClaimConfig{
		Network:                   network,
		RPCUrl:                    rpcUrl,
		EarnerAddress:             earnerAddress,
		Output:                    output,
		PathToKeyStore:            pathToKeyStore,
		SubmitClaim:               submitClaim,
		TokenAddresses:            tokenAddressArray,
		RewardsCoordinatorAddress: common.HexToAddress(rewardsCoordinatorAddress),
	}, nil
}

func stringToAddressArray(addresses []string) []common.Address {
	var addressArray []common.Address
	for _, address := range addresses {
		addressArray = append(addressArray, common.HexToAddress(address))
	}
	return addressArray
}
