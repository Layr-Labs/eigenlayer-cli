package release

import (
	"context"
	"flag"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"testing"
	"time"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/release-management-service-client/pkg/model"
)

type stubRMSClient struct {
	listOperatorReleasesFunc func(ctx context.Context, operatorId string) (*model.ListOperatorRequirementsResponse, error)
}

func (s *stubRMSClient) ListOperatorReleases(ctx context.Context, operatorId string) (*model.ListOperatorRequirementsResponse, error) {
	return s.listOperatorReleasesFunc(ctx, operatorId)
}

func (s *stubRMSClient) ListAvsReleaseKeys(ctx context.Context, avsId string) (*model.ListAvsReleaseKeysResponse, error) {
	panic("not implemented")
}

func TestListOperatorReleases(t *testing.T) {
	now := time.Now().UTC().Format(time.RFC3339)

	client := &stubRMSClient{
		listOperatorReleasesFunc: func(ctx context.Context, operatorId string) (*model.ListOperatorRequirementsResponse, error) {
			return &model.ListOperatorRequirementsResponse{
				OperatorRequirements: []model.OperatorApplication{
					{
						ApplicationName: "TestApp",
						OperatorSetId:   "os-123",
						Description:     "Test description",
						Components: []model.Component{
							{
								Name:             "CompA",
								Description:      "Component A",
								Location:         "ghcr.io/location",
								LatestArtifactId: "artifact-abc",
								ReleaseTimestamp: now,
							},
						},
					},
				},
			}, nil
		},
	}

	cmd := &ListOperatorReleasesCmd{}
	cfg := &listOperatorReleasesConfig{
		OperatorId: "0xabc",
		RmsClient:  client,
	}

	resp, err := cmd.listOperatorReleases(cfg)

	assert.NoError(t, err)
	assert.Len(t, resp.OperatorRequirements, 1)
	assert.Equal(t, "TestApp", resp.OperatorRequirements[0].ApplicationName)
	assert.Equal(t, "CompA", resp.OperatorRequirements[0].Components[0].Name)
}

func flagSet(values map[string]string) *flag.FlagSet {
	set := flag.NewFlagSet("test", 0)
	for key, value := range values {
		set.String(key, value, "")
		_ = set.Set(key, value)
	}
	return set
}

func TestReadAndValidateListOperatorReleasesConfig_FullInput(t *testing.T) {
	app := cli.NewApp()
	set := flagSet(map[string]string{
		flags.NetworkFlag.Name:         "holesky",
		flags.OperatorAddressFlag.Name: "0xabc123",
		flags.EnvironmentFlag.Name:     "testnet",
		flags.OutputTypeFlag.Name:      "json",
		flags.OutputFileFlag.Name:      "out.json",
	})
	ctx := cli.NewContext(app, set, nil)

	cfg, err := readAndValidateListOperatorReleasesConfig(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "holesky", cfg.Network)
	assert.Equal(t, "0xabc123", cfg.OperatorId)
	assert.Equal(t, "testnet", cfg.Environment)
	assert.Equal(t, "json", cfg.OutputType)
	assert.Equal(t, "out.json", cfg.OutputFile)
	assert.NotNil(t, cfg.RmsClient)
}

func TestReadAndValidateListOperatorReleasesConfig_DefaultEnv(t *testing.T) {
	app := cli.NewApp()
	set := flagSet(map[string]string{
		flags.NetworkFlag.Name:         "holesky",
		flags.OperatorAddressFlag.Name: "0xdef456",
		flags.EnvironmentFlag.Name:     "",
		flags.OutputTypeFlag.Name:      "json",
		flags.OutputFileFlag.Name:      "out.json",
	})
	ctx := cli.NewContext(app, set, nil)

	cfg, err := readAndValidateListOperatorReleasesConfig(ctx)
	assert.NoError(t, err)
	assert.Equal(t, common.GetEnvFromNetwork("holesky"), cfg.Environment)
}

func TestReadAndValidateListOperatorReleasesConfig_MissingOperatorId(t *testing.T) {
	app := cli.NewApp()
	set := flagSet(map[string]string{
		flags.NetworkFlag.Name:         "holesky",
		flags.OperatorAddressFlag.Name: "",
		flags.EnvironmentFlag.Name:     "testnet",
		flags.OutputTypeFlag.Name:      "json",
		flags.OutputFileFlag.Name:      "out.json",
	})
	ctx := cli.NewContext(app, set, nil)

	cfg, err := readAndValidateListOperatorReleasesConfig(ctx)
	assert.Nil(t, cfg)
	assert.ErrorContains(t, err, "operator Id is required")
}
