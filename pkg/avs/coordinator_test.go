package avs

import (
	"testing"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils/codec"

	"github.com/stretchr/testify/assert"
)

func TestLoadAllSpecificationsAndCreateCoordinators(t *testing.T) {
	t.Cleanup(CleanupRepository)
	repo := SetupRepository(t, true)
	specs, err := repo.List()
	assert.NoError(t, err)
	assert.NotNil(t, specs)
	assert.NotEmpty(t, specs)

	for _, base := range *specs {
		if base.Coordinator == "plugin" {
			continue
		}

		spec, err := NewSpecification(repo, base.Name)
		assert.NoError(t, err)
		assert.NotNil(t, spec)

		config := Configuration{
			repo.logger,
			nil,
			map[string]interface{}{
				"eth_rpc_url": "http://rpc.test.com",
				"signer_type": "local_keystore",
			},
			codec.NewJSONCodec(),
		}

		coordinator, err := NewCoordinator(repo, repo.logger, spec, config, false)
		assert.NoError(t, err)
		assert.NotNil(t, coordinator)
	}
}

func TestCreateCoordinatorsAndValidate(t *testing.T) {
	t.Cleanup(CleanupRepository)
	repo := SetupRepository(t, true)
	_, err := repo.List()
	assert.NoError(t, err)

	specs := map[string]string{
		"contract":   "holesky/lagrange-sc",
		"middleware": "holesky/eigenda",
	}

	for k, v := range specs {
		spec, err := NewSpecification(repo, v)
		assert.NoError(t, err)
		assert.NotNil(t, spec)
		assert.Equal(t, k, spec.Type())

		config := Configuration{
			repo.logger,
			nil,
			map[string]interface{}{
				"eth_rpc_url": "http://rpc.test.com",
				"signer_type": "local_keystore",
			},
			codec.NewJSONCodec(),
		}

		coordinator, err := NewCoordinator(repo, repo.logger, spec, config, false)
		assert.NoError(t, err)
		assert.NotNil(t, coordinator)
		assert.Equal(t, k, coordinator.Type())
	}
}
