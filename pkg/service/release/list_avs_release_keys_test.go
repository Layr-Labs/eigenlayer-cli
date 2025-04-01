package release

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
)

func TestReadAndValidateListAvsReleaseKeysConfig_FullInput(t *testing.T) {
	app := cli.NewApp()
	set := flagSet(map[string]string{
		flags.NetworkFlag.Name:      "holesky",
		flags.AVSAddressesFlag.Name: "avs-123",
		flags.EnvironmentFlag.Name:  "testnet",
		flags.OutputTypeFlag.Name:   "json",
		flags.OutputFileFlag.Name:   "out.json",
	})
	ctx := cli.NewContext(app, set, nil)

	cfg, err := readAndValidateListAvsReleaseKeysConfig(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "holesky", cfg.Network)
	assert.Equal(t, "avs-123", cfg.AvsId)
	assert.Equal(t, "testnet", cfg.Environment)
	assert.Equal(t, "json", cfg.OutputType)
	assert.Equal(t, "out.json", cfg.OutputFile)
	assert.NotNil(t, cfg.RmsClient)
}

func TestReadAndValidateListAvsReleaseKeysConfig_DefaultEnv(t *testing.T) {
	app := cli.NewApp()
	set := flagSet(map[string]string{
		flags.NetworkFlag.Name:      "holesky",
		flags.AVSAddressesFlag.Name: "avs-456",
		flags.EnvironmentFlag.Name:  "",
		flags.OutputTypeFlag.Name:   "json",
		flags.OutputFileFlag.Name:   "",
	})
	ctx := cli.NewContext(app, set, nil)

	cfg, err := readAndValidateListAvsReleaseKeysConfig(ctx)
	assert.NoError(t, err)
	assert.Equal(t, common.GetEnvFromNetwork("holesky"), cfg.Environment)
}

func TestReadAndValidateListAvsReleaseKeysConfig_MissingAvsId(t *testing.T) {
	app := cli.NewApp()
	set := flagSet(map[string]string{
		flags.NetworkFlag.Name:     "holesky",
		flags.AVSAddressFlag.Name:  "",
		flags.EnvironmentFlag.Name: "testnet",
		flags.OutputTypeFlag.Name:  "json",
		flags.OutputFileFlag.Name:  "",
	})
	ctx := cli.NewContext(app, set, nil)

	cfg, err := readAndValidateListAvsReleaseKeysConfig(ctx)
	assert.Nil(t, cfg)
	assert.ErrorContains(t, err, "AVS Id is required")
}
