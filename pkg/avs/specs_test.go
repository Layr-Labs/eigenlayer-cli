package avs

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
	"testing"
)

func runSpecsCmd(t *testing.T, cmd *cli.Command, reset bool) string {
	RepositorySubFolder = ".eigenlayer-test/avs/specs"

	home, err := os.UserHomeDir()
	assert.NoError(t, err)

	path := filepath.Clean(filepath.Join(home, RepositorySubFolder))
	if reset {
		err = os.RemoveAll(path)
		assert.NoError(t, err)
	}

	app := cli.NewApp()
	flags := []string{"-v"}
	args := append([]string{""}, flags...)
	ctx := cli.NewContext(app, nil, &cli.Context{Context: context.Background()})

	err = cmd.Run(ctx, args...)
	assert.NoError(t, err)

	return path
}

func cleanupSpecsCmd() {
	RepositorySubFolder = ".eigenlayer/avs/specs"

	home, _ := os.UserHomeDir()
	path := filepath.Clean(filepath.Join(home, ".eigenlayer-test"))
	_ = os.RemoveAll(path)
}

func TestSpecsListOnEmptyRepository(t *testing.T) {
	t.Cleanup(cleanupSpecsCmd)
	repo := runSpecsCmd(t, ListCmd(), true)
	assert.FileExists(t, filepath.Join(repo, "holesky/eigenda/avs.json"))
	assert.FileExists(t, filepath.Join(repo, "mainnet/eigenda/avs.json"))
}

func TestSpecsListOnNonEmptyRepository(t *testing.T) {
	t.Cleanup(cleanupSpecsCmd)
	repo := runSpecsCmd(t, ListCmd(), true)
	err := os.RemoveAll(filepath.Join(repo, "holesky/eigenda"))
	assert.NoError(t, err)

	_ = runSpecsCmd(t, ListCmd(), false)
	assert.NoFileExists(t, filepath.Join(repo, "holesky/eigenda/avs.json"))
	assert.FileExists(t, filepath.Join(repo, "mainnet/eigenda/avs.json"))
}

func TestSpecsResetOnEmptyRepository(t *testing.T) {
	t.Cleanup(cleanupSpecsCmd)
	repo := runSpecsCmd(t, ResetCmd(), true)
	assert.FileExists(t, filepath.Join(repo, "holesky/eigenda/avs.json"))
	assert.FileExists(t, filepath.Join(repo, "mainnet/eigenda/avs.json"))
}

func TestSpecsResetOnNonEmptyRepository(t *testing.T) {
	t.Cleanup(cleanupSpecsCmd)
	repo := runSpecsCmd(t, ListCmd(), true)
	err := os.RemoveAll(filepath.Join(repo, "holesky/eigenda"))
	assert.NoError(t, err)

	_ = runSpecsCmd(t, ResetCmd(), false)
	assert.FileExists(t, filepath.Join(repo, "holesky/eigenda/avs.json"))
	assert.FileExists(t, filepath.Join(repo, "mainnet/eigenda/avs.json"))
}
