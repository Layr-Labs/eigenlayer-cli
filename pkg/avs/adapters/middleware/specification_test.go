package middleware

import (
	"testing"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/avs/adapters"
	"github.com/stretchr/testify/assert"
)

func NewTestSpecification() *Specification {
	return &Specification{
		BaseSpecification: adapters.BaseSpecification{
			Name:            "holesky/eigenda",
			Description:     "EigenDA on Holesky Testnet",
			Network:         "holesky",
			ContractAddress: "0xD4A7E1Bd8015057293f0D0A557088c286942e84b",
			Coordinator:     "middleware",
			RemoteSigning:   false,
		},
		ChurnerURL: "churner-holesky.eigenda.xyz:443",
	}
}

func TestSpecification(t *testing.T) {
	spec := NewTestSpecification()
	assert.NoError(t, spec.Validate())
}

func TestSpecificationWithMissingName(t *testing.T) {
	spec := NewTestSpecification()
	spec.Name = ""
	assert.ErrorContains(t, spec.Validate(), "name is required")
}

func TestSpecificationWithMissingNetwork(t *testing.T) {
	spec := NewTestSpecification()
	spec.Network = ""
	assert.ErrorContains(t, spec.Validate(), "network is required")
}

func TestSpecificationWithMissingContractAddress(t *testing.T) {
	spec := NewTestSpecification()
	spec.ContractAddress = ""
	assert.ErrorContains(t, spec.Validate(), "contract address is required")
}

func TestSpecificationWithInvalidContractAddress(t *testing.T) {
	spec := NewTestSpecification()
	spec.ContractAddress = "invalid"
	assert.ErrorContains(t, spec.Validate(), "contract address is invalid")
}

func TestSpecificationWithInvalidCoordinator(t *testing.T) {
	spec := NewTestSpecification()
	spec.Coordinator = "invalid"
	assert.ErrorContains(t, spec.Validate(), "coordinator is invalid")
}
