package release

import (
	"context"
	"fmt"
	"time"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/command"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	"github.com/Layr-Labs/release-management-service-client/pkg/client"

	"github.com/urfave/cli/v2"
)

type ListAvsReleaseKeysCmd struct {
	prompter utils.Prompter
}

func NewListAvsReleaseKeysCmd(prompter utils.Prompter) *cli.Command {
	delegateCommand := &ListAvsReleaseKeysCmd{prompter}
	listAvsReleaseKeysCmd := command.NewServiceCommand(
		delegateCommand,
		"list-avs-release-keys",
		"List valid keys used by AVSs to sign released artifacts",
		"list-avs-release-keys --avs-address",
		"",
		getListAvsReleaseKeysFlags(),
	)

	return listAvsReleaseKeysCmd
}

func (l *ListAvsReleaseKeysCmd) Execute(c *cli.Context) error {
	logger := common.GetLogger(c)
	config, err := readAndValidateListAvsReleaseKeysConfig(c)
	if err != nil {
		return err
	}

	keys, err := l.listAvsReleaseKeys(config)
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		logger.Info("No release keys found for the AVS")
		return nil
	}

	logger.Info("Found release keys:")
	for _, key := range keys {
		logger.Infof("- %s", key)
	}

	return nil
}

func readAndValidateListAvsReleaseKeysConfig(c *cli.Context) (*listAvsReleaseKeysConfig, error) {
	network := c.String(flags.NetworkFlag.Name)

	avsId := c.String(flags.AVSAddressesFlag.Name)
	if avsId == "" {
		return nil, fmt.Errorf("AVS Id is required")
	}

	environment := c.String(flags.EnvironmentFlag.Name)
	if environment == "" {
		environment = common.GetEnvFromNetwork(network)
	}

	outputType := c.String(flags.OutputTypeFlag.Name)
	outputFile := c.String(flags.OutputFileFlag.Name)

	clientConfig := client.NewClientConfig("", environment, 500*time.Millisecond, nil)
	rmsClient, err := client.NewClient(clientConfig)
	if err != nil {
		return nil, err
	}

	return &listAvsReleaseKeysConfig{
		Network:     network,
		AvsId:       avsId,
		Environment: environment,
		OutputType:  outputType,
		OutputFile:  outputFile,
		RmsClient:   rmsClient,
	}, nil
}

func (l *ListAvsReleaseKeysCmd) listAvsReleaseKeys(config *listAvsReleaseKeysConfig) ([]string, error) {
	ctx := context.Background()

	resp, err := config.RmsClient.ListAvsReleaseKeys(ctx, config.AvsId)
	if err != nil {
		return nil, err
	}

	return resp.Keys, nil
}

func getListAvsReleaseKeysFlags() []cli.Flag {
	return []cli.Flag{
		&flags.OutputTypeFlag,
		&flags.OutputFileFlag,
		&flags.AVSAddressesFlag,
	}
}
