package avs

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/operator/config"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"go.uber.org/mock/gomock"
)

type ConfigCreateTestCase struct {
	name   string
	config string
	args   []string
	err    error
	reset  bool
	prompt func(p *mocks.MockPrompter)
}

var (
	tempDir = "./config-test"
)

func beforeAll(t *testing.T) {
	RepositorySubFolder = ".eigenlayer-test/avs/specs"

	assert.NoError(t, os.RemoveAll(tempDir))
	assert.NoError(t, os.MkdirAll(tempDir, os.ModePerm))
}

func beforeEachConfigCreate(t *testing.T, tt ConfigCreateTestCase) *cli.Command {
	controller := gomock.NewController(t)
	p := mocks.NewMockPrompter(controller)
	if tt.prompt != nil {
		tt.prompt(p)
	}

	assert.NoError(t, runCmd(t, ResetCmd(), []string{""}))
	assert.NoError(t, runCmd(t, config.CreateCmd(p), []string{"-y"}))

	return CreateCmd(p)
}

func afterEachConfigCreate(t *testing.T) {
	assert.NoError(t, os.RemoveAll(OperatorConfigFile))
	assert.NoError(t, os.RemoveAll(OperatorMetaFile))
}

func afterAll(t *testing.T) {
	home, _ := os.UserHomeDir()
	path := filepath.Clean(filepath.Join(home, ".eigenlayer-test"))
	assert.NoError(t, os.RemoveAll(path))
	assert.NoError(t, os.RemoveAll(tempDir))

	RepositorySubFolder = ".eigenlayer/avs/specs"
}

func runCmd(t *testing.T, cmd *cli.Command, flags []string) error {
	app := cli.NewApp()
	args := append([]string{""}, flags...)
	ctx := cli.NewContext(app, nil, &cli.Context{Context: context.Background()})
	err := cmd.Run(ctx, args...)
	assert.NoError(t, err)

	return nil
}

func TestConfigCreate(t *testing.T) {
	beforeAll(t)
	t.Cleanup(func() {
		afterAll(t)
	})

	tests := []ConfigCreateTestCase{
		{
			name:   "required argument not provided",
			config: filepath.Join(tempDir, "test.yaml"),
			args:   []string{},
			err:    errors.New("invalid number of arguments: accepts 1 arg, received 0"),
			reset:  true,
		},
		{
			name:   "provided additional arguments",
			config: filepath.Join(tempDir, "test.yaml"),
			args:   []string{"holesky/eigenda", filepath.Join(tempDir, "test.yaml"), "additional"},
			err:    errors.New("invalid number of arguments: accepts 1 arg, received 3"),
			reset:  true,
		},
		{
			name:   "empty avs id",
			config: filepath.Join(tempDir, "test.yaml"),
			args:   []string{""},
			err:    errors.New("argument value cannot be empty: provided avs id is empty"),
			reset:  true,
		},
		{
			name:   "empty avs config file",
			config: filepath.Join(tempDir, "test.yaml"),
			args:   []string{"holesky/eigenda", ""},
			err:    errors.New("failed: provided avs config is empty"),
			reset:  true,
		},
		{
			name:   "create config file for invalid avs",
			config: filepath.Join(tempDir, "test.yaml"),
			args:   []string{"invalid", filepath.Join(tempDir, "test.yaml")},
			err: errors.New(
				"failed: config template file expected in directory .eigenlayer-test/avs/specs/invalid",
			),
			reset: true,
		},
		{
			name:   "create new config file",
			config: filepath.Join(tempDir, "test.yaml"),
			args:   []string{"holesky/eigenda", filepath.Join(tempDir, "test.yaml")},
			err:    nil,
		},
		{
			name:   "confirm when creating the same config file",
			config: filepath.Join(tempDir, "test.yaml"),
			args:   []string{"holesky/eigenda", filepath.Join(tempDir, "test.yaml")},
			err:    nil,
			reset:  false,
			prompt: func(p *mocks.MockPrompter) {
				p.EXPECT().
					Confirm("This will overwrite existing avs config file. Are you sure you want to continue?").
					Return(true, nil)
			},
		},
		{
			name:   "not confirmed when creating the same config file",
			config: filepath.Join(tempDir, "test.yaml"),
			args:   []string{"holesky/eigenda", filepath.Join(tempDir, "test.yaml")},
			err:    nil,
			reset:  false,
			prompt: func(p *mocks.MockPrompter) {
				p.EXPECT().
					Confirm("This will overwrite existing avs config file. Are you sure you want to continue?").
					Return(false, nil)
			},
		},
		{
			name:   "create new config file without config name",
			config: "holesky-eigenda-config.yaml",
			args:   []string{"holesky/eigenda"},
			err:    nil,
			reset:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := beforeEachConfigCreate(t, tt)
			app := cli.NewApp()
			args := append([]string{""}, tt.args...)
			ctx := cli.NewContext(app, nil, &cli.Context{Context: context.Background()})
			err := cmd.Run(ctx, args...)

			if tt.err == nil {
				assert.NoError(t, err)
				assert.FileExists(t, tt.config)
				if tt.reset {
					_ = os.RemoveAll(tt.config)
				}
			} else {
				assert.ErrorContains(t, err, tt.err.Error())
			}

			afterEachConfigCreate(t)
		})
	}
}
