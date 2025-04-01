package release

import (
	"context"
	"fmt"
	"time"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/command"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/Layr-Labs/release-management-service-client/pkg/client"
	"github.com/Layr-Labs/release-management-service-client/pkg/model"

	"github.com/urfave/cli/v2"
)

type ListOperatorReleasesCmd struct {
	prompter utils.Prompter
}

func NewListOperatorReleasesCmd(prompter utils.Prompter) *cli.Command {
	delegateCommand := &ListOperatorReleasesCmd{prompter}
	listOperatorReleasesCmd := command.NewServiceCommand(
		delegateCommand,
		"list-operator-releases",
		"List AVS application releases a given operator is registered to run",
		"list-operator-releases --operator-address",
		"",
		getListOperatorReleasesFlags(),
	)

	return listOperatorReleasesCmd
}

func (l *ListOperatorReleasesCmd) Execute(c *cli.Context) error {
	logger := common.GetLogger(c)
	config, err := readAndValidateListOperatorReleasesConfig(c)
	if err != nil {
		return err
	}

	releases, err := l.listOperatorReleases(config)
	if err != nil {
		return err
	}

	if len(releases.OperatorRequirements) == 0 {
		logger.Info("No releases found for the operator")
		return nil
	}

	printReleases(releases, logger)

	return nil
}

func readAndValidateListOperatorReleasesConfig(
	c *cli.Context,
) (*listOperatorReleasesConfig, error) {
	network := c.String(flags.NetworkFlag.Name)

	operatorId := c.String(flags.OperatorAddressFlag.Name)
	if operatorId == "" {
		return nil, fmt.Errorf("operator Id is required")
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

	return &listOperatorReleasesConfig{
		Network:     network,
		OperatorId:  operatorId,
		Environment: environment,
		OutputType:  outputType,
		OutputFile:  outputFile,
		RmsClient:   rmsClient,
	}, nil
}

func (l *ListOperatorReleasesCmd) listOperatorReleases(
	config *listOperatorReleasesConfig,
) (*model.ListOperatorRequirementsResponse, error) {
	ctx := context.Background()

	resp, err := config.RmsClient.ListOperatorReleases(ctx, config.OperatorId)
	if err != nil {
		return nil, err
	}

	var operatorApps []model.OperatorApplication
	for _, app := range resp.OperatorRequirements {
		var components []model.Component
		for _, c := range app.Components {
			components = append(components, model.Component{
				Name:             c.Name,
				Description:      c.Description,
				Location:         c.Location,
				LatestArtifactId: c.LatestArtifactId,
				ReleaseTimestamp: c.ReleaseTimestamp,
			})
		}

		operatorApps = append(operatorApps, model.OperatorApplication{
			ApplicationName: app.ApplicationName,
			OperatorSetId:   app.OperatorSetId,
			Description:     app.Description,
			Components:      components,
		})
	}

	return &model.ListOperatorRequirementsResponse{
		OperatorRequirements: operatorApps,
	}, nil
}

func printReleases(releases *model.ListOperatorRequirementsResponse, logger logging.Logger) {
	logger.Info("Found releases:")
	for _, app := range releases.OperatorRequirements {
		logger.Infof("- %s (Operator Set ID: %s)", app.ApplicationName, app.OperatorSetId)
		logger.Infof("  Description: %s", app.Description)
		logger.Info("  Components:")
		for _, component := range app.Components {
			logger.Infof("    - %s", component.Name)
			logger.Infof("      Description: %s", component.Description)
			logger.Infof("      Location: %s", component.Location)
			logger.Infof("      Latest Artifact ID: %s", component.LatestArtifactId)
			logger.Infof("      Release Timestamp: %s", component.ReleaseTimestamp)
		}
	}
}

func getListOperatorReleasesFlags() []cli.Flag {
	return []cli.Flag{
		&flags.OutputTypeFlag,
		&flags.OutputFileFlag,
		&flags.OperatorAddressFlag,
	}
}
