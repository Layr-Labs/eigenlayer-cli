package rewards

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"math/big"
	"net/http"
	"strings"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	"github.com/Layr-Labs/eigensdk-go/logging"

	gethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/urfave/cli/v2"
)

var (
	preprodUrl = ""
	testnetUrl = ""
	mainnetUrl = ""
)

type ClaimType string

const (
	All       ClaimType = "all"
	Unclaimed ClaimType = "unclaimed"
	Claimed   ClaimType = "claimed"

	GetClaimableRewardsEndpoint        = "grpc/eigenlayer.RewardsService/GetClaimableRewards"
	GetEarnedTokensForStrategyEndpoint = "grpc/eigenlayer.RewardsService/GetEarnedTokensForStrategy"
)

func ShowCmd(p utils.Prompter) *cli.Command {
	showCmd := &cli.Command{
		Name:      "show",
		Usage:     "Show rewards for an address",
		UsageText: "show",
		Description: `
Command to show rewards for earners

Currently supports past total rewards (claimed and unclaimed) and past unclaimed rewards
		`,
		After: telemetry.AfterRunAction(),
		Flags: getShowFlags(),
		Action: func(cCtx *cli.Context) error {
			return ShowRewards(cCtx)
		},
	}

	return showCmd
}

func getShowFlags() []cli.Flag {
	baseFlags := []cli.Flag{
		&flags.NetworkFlag,
		&flags.OutputFileFlag,
		&flags.VerboseFlag,
		&EarnerAddressFlag,
		&NumberOfDaysFlag,
		&AVSAddressesFlag,
		&EnvironmentFlag,
		&ClaimTypeFlag,
	}

	sort.Sort(cli.FlagsByName(baseFlags))
	return baseFlags
}

func ShowRewards(cCtx *cli.Context) error {
	logger := common.GetLogger(cCtx)

	config, err := readAndValidateConfig(cCtx, logger)
	if err != nil {
		return fmt.Errorf("error reading and validating config: %s", err)
	}
	cCtx.App.Metadata["network"] = config.ChainID.String()

	url := testnetUrl
	if config.Environment == "mainnet" {
		url = mainnetUrl
	} else if config.Environment == "preprod" {
		url = preprodUrl
	}

	if config.ClaimType == All {
		requestBody := map[string]string{
			"earnerAddress": config.EarnerAddress.String(),
			"days":          fmt.Sprintf("%d", absInt64(config.NumberOfDays)),
		}
		resp, err := post(
			fmt.Sprintf("%s/%s", url, GetEarnedTokensForStrategyEndpoint),
			requestBody,
		)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		var responseBody RewardResponse
		err = json.NewDecoder(resp.Body).Decode(&responseBody)
		if err != nil {
			return err
		}
		normalizedRewards := normalizeRewardResponse(responseBody)
		if common.IsEmptyString(config.Output) {
			printNormalizedRewardsAsTable(normalizedRewards)
		} else {
			logger.Debugf("Writing total rewards to %s", config.Output)
			err = common.WriteToCSV(normalizedRewards, config.Output)
			if err != nil {
				return err
			}
			logger.Infof("Total rewards written to %s", config.Output)
		}
	} else if config.ClaimType == Unclaimed {
		requestBody := map[string]string{
			"earnerAddress": config.EarnerAddress.String(),
		}
		claimableRewardsUrl := fmt.Sprintf("%s/%s", url, GetClaimableRewardsEndpoint)
		resp, err := post(claimableRewardsUrl, requestBody)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		var response UnclaimedRewardResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			return err
		}
		unclaimedNormalizedRewards := normalizeUnclaimedRewardResponse(response)
		if common.IsEmptyString(config.Output) {
			printUnclaimedNormalizedRewardsAsTable(unclaimedNormalizedRewards)
		} else {
			logger.Debugf("Writing unclaimed rewards to %s", config.Output)
			err = common.WriteToCSV(unclaimedNormalizedRewards, config.Output)
			if err != nil {
				return err
			}
			logger.Infof("Unclaimed rewards written to %s", config.Output)
		}
	} else {
		return fmt.Errorf("claim type %s not supported", config.ClaimType)
	}

	return nil
}

func post(url string, requestBody map[string]string) (*http.Response, error) {
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}
	return http.Post(url, "application/json", bytes.NewBuffer(jsonData))
}

func normalizeUnclaimedRewardResponse(unclaimedRewardResponse UnclaimedRewardResponse) []NormalizedUnclaimedReward {
	var normalizedUnclaimedRewards []NormalizedUnclaimedReward
	for _, rewardsPerAVS := range unclaimedRewardResponse.Rewards {
		for _, token := range rewardsPerAVS.Tokens {
			amount := new(big.Int)
			amount.SetString(token.WeiAmount, 10)
			normalizedUnclaimedRewards = append(normalizedUnclaimedRewards, NormalizedUnclaimedReward{
				TokenAddress: token.TokenAddress,
				WeiAmount:    amount,
			})
		}
	}
	return normalizedUnclaimedRewards
}

func normalizeRewardResponse(rewardResponse RewardResponse) []NormalizedReward {
	var normalizedRewards []NormalizedReward
	for _, reward := range rewardResponse.Rewards {
		for _, rewardsPerAVS := range reward.RewardsPerStrategy {
			for _, token := range rewardsPerAVS.Tokens {
				amount := new(big.Int)
				amount.SetString(token.WeiAmount, 10)
				normalizedRewards = append(normalizedRewards, NormalizedReward{
					StrategyAddress: reward.StrategyAddress,
					AVSAddress:      rewardsPerAVS.AVSAddress,
					TokenAddress:    token.TokenAddress,
					WeiAmount:       amount,
				})
			}
		}
	}
	return normalizedRewards
}

func printUnclaimedNormalizedRewardsAsTable(normalizedRewards []NormalizedUnclaimedReward) {
	column := formatColumns(
		"Token Address",
		common.MaxAddressLength,
	) + " | " + formatColumns(
		"Wei Amount",
		common.MaxAddressLength,
	)
	fmt.Println(strings.Repeat("-", len(column)))
	fmt.Println(column)
	fmt.Println(strings.Repeat("-", len(column)))
	for _, reward := range normalizedRewards {
		if reward.WeiAmount.Cmp(big.NewInt(0)) == 0 {
			continue
		}
		fmt.Printf(
			"%s | %s\n",
			reward.TokenAddress,
			reward.WeiAmount.String(),
		)
	}
	fmt.Println(strings.Repeat("-", len(column)))
}

func printNormalizedRewardsAsTable(normalizedRewards []NormalizedReward) {
	column := formatColumns(
		"Strategy Address",
		common.MaxAddressLength,
	) + " | " + formatColumns(
		"AVS Address",
		common.MaxAddressLength,
	) + " | " + formatColumns(
		"Token Address",
		common.MaxAddressLength,
	) + " | " + formatColumns(
		"Wei Amount",
		common.MaxAddressLength,
	)
	fmt.Println(strings.Repeat("-", len(column)))
	fmt.Println(column)
	fmt.Println(strings.Repeat("-", len(column)))
	for _, reward := range normalizedRewards {
		fmt.Printf(
			"%s | %s | %s | %s\n",
			reward.StrategyAddress,
			reward.AVSAddress,
			reward.TokenAddress,
			reward.WeiAmount.String(),
		)
	}
	fmt.Println(strings.Repeat("-", len(column)))
}

func formatColumns(columnName string, size int32) string {
	return fmt.Sprintf("%-*s", size, columnName)
}

func readAndValidateConfig(cCtx *cli.Context, logger logging.Logger) (*ShowConfig, error) {
	earnerAddress := gethcommon.HexToAddress(cCtx.String(EarnerAddressFlag.Name))
	output := cCtx.String(flags.OutputFileFlag.Name)
	numberOfDays := cCtx.Int64(NumberOfDaysFlag.Name)
	if numberOfDays >= 0 {
		return nil, errors.New(
			"future rewards projection is not supported yet. Please provide a negative number of days for past rewards",
		)
	}
	network := cCtx.String(flags.NetworkFlag.Name)
	env := cCtx.String(EnvironmentFlag.Name)
	if env == "" {
		env = getEnvFromNetwork(network)
	}
	logger.Debugf("Network: %s, Env: %s", network, env)

	claimType := ClaimType(cCtx.String(ClaimTypeFlag.Name))
	if claimType != All && claimType != Unclaimed && claimType != Claimed {
		return nil, errors.New("claim type must be 'all', 'unclaimed' or 'claimed'")
	}
	logger.Debugf("Claim Type: %s", claimType)
	chainID := utils.NetworkNameToChainId(network)
	logger.Debugf("Using chain ID: %s", chainID.String())

	return &ShowConfig{
		EarnerAddress: earnerAddress,
		NumberOfDays:  numberOfDays,
		Network:       network,
		Environment:   env,
		ClaimType:     claimType,
		ChainID:       chainID,
		Output:        output,
	}, nil
}

// absInt64 returns the absolute value of an int64.
func absInt64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
