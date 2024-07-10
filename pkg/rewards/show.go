package rewards

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/common/flags"
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

type ShowConfig struct {
	EarnerAddress gethcommon.Address
	NumberOfDays  int64
	Network       string
	Environment   string
}

func ShowCmd(p utils.Prompter) *cli.Command {
	showCmd := &cli.Command{
		Name:      "show",
		Usage:     "Show rewards for an address",
		UsageText: "show",
		Description: `
Command to show rewards for earners
		`,
		After: telemetry.AfterRunAction(),
		Flags: []cli.Flag{
			&flags.NetworkFlag,
			&flags.OutputFileFlag,
			&flags.VerboseFlag,
			&EarnerAddressFlag,
			&NumberOfDaysFlag,
			&AVSAddressesFlag,
			&EnvironmentFlag,
		},
		Action: func(cCtx *cli.Context) error {
			return ShowRewards(cCtx)
		},
	}

	return showCmd
}

func ShowRewards(cCtx *cli.Context) error {
	//ctx := cCtx.Context

	verbose := cCtx.Bool(flags.VerboseFlag.Name)
	logLevel := slog.LevelInfo
	if verbose {
		logLevel = slog.LevelDebug
	}
	logger := logging.NewTextSLogger(os.Stdout, &logging.SLoggerOptions{Level: logLevel})

	config, err := readAndValidateConfig(cCtx, logger)
	if err != nil {
		return fmt.Errorf("error reading and validating config: %s", err)
	}

	// Data to be sent in the request body
	requestBody := map[string]string{
		"earnerAddress": config.EarnerAddress.String(),
		"days":          fmt.Sprintf("%d", AbsInt64(config.NumberOfDays)),
	}

	// Convert the request body to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return nil
	}

	var url = preprodUrl
	if config.Environment == "prod" {
		url = mainnetUrl
	} else if config.Environment == "testnet" {
		url = testnetUrl
	}

	showRewardsURL := fmt.Sprintf("%s/%s", url, "grpc/eigenlayer.RewardsService/GetEarnedTokensForStrategy")
	resp, err := http.Post(
		showRewardsURL,
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	var responseBody RewardResponse
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		return err
	}
	printRewards(responseBody)

	return nil
}

func printRewards(rewardResponse RewardResponse) {
	space := "--"
	for _, reward := range rewardResponse.Rewards {
		fmt.Println(space + "Strategy Address: " + reward.StrategyAddress)
		for _, rewardsPerAVS := range reward.RewardsPerStrategy {
			avsSpace := space + space
			fmt.Println(avsSpace + "AVS Address: " + rewardsPerAVS.AVSAddress)
			for _, token := range rewardsPerAVS.Tokens {
				tokenSpace := avsSpace + space
				fmt.Println(tokenSpace + "Token Address: " + token.TokenAddress)
				fmt.Println(tokenSpace + "Token Amount (in Wei): " + token.WeiAmount)
			}
		}
	}
}

func readAndValidateConfig(cCtx *cli.Context, logger logging.Logger) (*ShowConfig, error) {
	earnerAddress := gethcommon.HexToAddress(cCtx.String(EarnerAddressFlag.Name))
	numberOfDays := cCtx.Int64(NumberOfDaysFlag.Name)
	if numberOfDays >= 0 {
		return nil, errors.New("future rewards projection is not supported yet. Please provide a negative number of days for past rewards")
	}
	network := cCtx.String(flags.NetworkFlag.Name)
	env := cCtx.String(EnvironmentFlag.Name)
	if env != "" {
		network = envToNetwork(env)
		logger.Debugf("Env: %s", env)
	}
	logger.Debugf("Network: %s", network)

	return &ShowConfig{
		EarnerAddress: earnerAddress,
		NumberOfDays:  numberOfDays,
		Network:       network,
		Environment:   env,
	}, nil
}

func envToNetwork(env string) string {
	switch env {
	case "preprod", "testnet":
		return "holesky"
	case "prod":
		return "mainnet"
	default:
		return "local"
	}
}

// AbsInt64 returns the absolute value of an int64.
func AbsInt64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
