package avs

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/avs/adapters"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/avs/adapters/contract"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/avs/adapters/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func SetupRepository(t *testing.T, reset bool) *Repository {
	RepositorySubFolder = ".eigenlayer-test/avs/specs"

	home, err := os.UserHomeDir()
	assert.NoError(t, err)

	path := filepath.Clean(filepath.Join(home, RepositorySubFolder))
	if reset {
		err = os.RemoveAll(path)
		assert.NoError(t, err)
	}

	app := cli.NewApp()
	ctx := cli.NewContext(app, nil, &cli.Context{Context: context.Background()})

	repo, err := NewRepository(ctx)
	assert.NoError(t, err)

	return repo
}

func CleanupRepository() {
	RepositorySubFolder = ".eigenlayer/avs/specs"

	home, _ := os.UserHomeDir()
	path := filepath.Clean(filepath.Join(home, ".eigenlayer-test"))
	_ = os.RemoveAll(path)
}

func ContainsSpec(t *testing.T, repo *Repository, specs *[]*adapters.BaseSpecification, name string) {
	ok := false
	for _, spec := range *specs {
		if spec.Name == name {
			ok = true
			break
		}
	}

	assert.True(t, ok)
	assert.FileExists(t, filepath.Join(repo.path, name, "avs.json"))
}

func NotContainsSpec(t *testing.T, repo *Repository, specs *[]*adapters.BaseSpecification, name string) {
	ok := false
	for _, spec := range *specs {
		if spec.Name == name {
			ok = true
			break
		}
	}

	assert.False(t, ok)
	assert.NoFileExists(t, filepath.Join(repo.path, name, "avs.json"))
}

func TestRepositoryListOnEmptyRepository(t *testing.T) {
	t.Cleanup(CleanupRepository)
	repo := SetupRepository(t, true)

	specs, err := repo.List()
	assert.NoError(t, err)
	ContainsSpec(t, repo, specs, "holesky/eigenda")
	ContainsSpec(t, repo, specs, "mainnet/eigenda")
}

func TestRepositoryListOnNonEmptyRepository(t *testing.T) {
	t.Cleanup(CleanupRepository)
	repo := SetupRepository(t, true)
	specs, err := repo.List()
	assert.NoError(t, err)
	ContainsSpec(t, repo, specs, "holesky/eigenda")
	ContainsSpec(t, repo, specs, "mainnet/eigenda")

	err = os.RemoveAll(filepath.Join(repo.path, "holesky/eigenda"))
	assert.NoError(t, err)

	specs, err = repo.List()
	assert.NoError(t, err)
	NotContainsSpec(t, repo, specs, "holesky/eigenda")
	ContainsSpec(t, repo, specs, "mainnet/eigenda")
}

func TestRepositoryResetOnEmptyRepository(t *testing.T) {
	t.Cleanup(CleanupRepository)
	repo := SetupRepository(t, true)

	err := repo.Reset()
	assert.NoError(t, err)
	assert.FileExists(t, filepath.Join(repo.path, "holesky/eigenda/avs.json"))
	assert.FileExists(t, filepath.Join(repo.path, "mainnet/eigenda/avs.json"))
}

func TestRepositoryResetOnNonEmptyRepository(t *testing.T) {
	t.Cleanup(CleanupRepository)
	repo := SetupRepository(t, true)
	_, err := repo.List()
	assert.NoError(t, err)

	err = os.RemoveAll(filepath.Join(repo.path, "holesky/eigenda"))
	assert.NoError(t, err)

	err = repo.Reset()
	assert.NoError(t, err)
	assert.FileExists(t, filepath.Join(repo.path, "holesky/eigenda/avs.json"))
	assert.FileExists(t, filepath.Join(repo.path, "mainnet/eigenda/avs.json"))
}

func TestLoadNonExistingSpecification(t *testing.T) {
	t.Cleanup(CleanupRepository)
	repo := SetupRepository(t, true)
	_, err := repo.List()
	assert.NoError(t, err)

	_, err = NewSpecification(repo, "invalid")
	assert.ErrorContains(t, err, "no such file or directory")
}

func TestLoadContractSpecification(t *testing.T) {
	t.Cleanup(CleanupRepository)
	repo := SetupRepository(t, true)
	_, err := repo.List()
	assert.NoError(t, err)

	spec, err := NewSpecification(repo, "holesky/lagrange-sc")
	assert.NoError(t, err)
	assert.NotNil(t, spec)

	contract, ok := spec.(contract.Specification)
	assert.True(t, ok)
	assert.Equal(t, "contract", contract.Type())

	assert.Equal(t, "holesky/lagrange-sc", contract.Name)
	assert.Equal(t, "Lagrange State Committees AVS on Holesky Testnet", contract.Description)
	assert.Equal(t, "holesky", contract.Network)
	assert.Equal(t, "0x18A74E66cc90F0B1744Da27E72Df338cEa0A542b", contract.ContractAddress)
	assert.Equal(t, "contract", contract.Coordinator)
	assert.Equal(t, true, contract.RemoteSigning)
	assert.Equal(t, "service_manager.json", contract.ABI)
	assert.Equal(t, 5, len(contract.Functions))
	assert.Equal(t, 2, len(contract.Delegates))
}

func TestLoadMiddlewareSpecification(t *testing.T) {
	t.Cleanup(CleanupRepository)
	repo := SetupRepository(t, true)
	_, err := repo.List()
	assert.NoError(t, err)

	spec, err := NewSpecification(repo, "holesky/eigenda")
	assert.NoError(t, err)
	assert.NotNil(t, spec)

	middleware, ok := spec.(middleware.Specification)
	assert.True(t, ok)
	assert.Equal(t, "middleware", middleware.Type())

	assert.Equal(t, "holesky/eigenda", middleware.Name)
	assert.Equal(t, "EigenDA on Holesky Testnet", middleware.Description)
	assert.Equal(t, "holesky", middleware.Network)
	assert.Equal(t, "0xD4A7E1Bd8015057293f0D0A557088c286942e84b", middleware.ContractAddress)
	assert.Equal(t, "middleware", middleware.Coordinator)
	assert.Equal(t, true, middleware.RemoteSigning)
	assert.Equal(t, "churner-holesky.eigenda.xyz:443", middleware.ChurnerURL)
}
