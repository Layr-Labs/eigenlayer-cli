package adapters

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func NewSpecification() *BaseSpecification {
	spec := BaseSpecification{
		"base",
		"Base AVS Specification",
		"mainnet",
		"0x0000000000000000000000000000000000000000",
		"base",
		false,
	}

	return &spec
}

func TestSpecification(t *testing.T) {
	spec := NewSpecification()
	assert.NoError(t, spec.Validate())
}

func TestSpecificationWithMissingName(t *testing.T) {
	spec := NewSpecification()
	spec.Name = ""
	assert.ErrorContains(t, spec.Validate(), "name is required")
}

func TestSpecificationWithMissingNetwork(t *testing.T) {
	spec := NewSpecification()
	spec.Network = ""
	assert.ErrorContains(t, spec.Validate(), "network is required")
}

func TestSpecificationWithMissingContractAddress(t *testing.T) {
	spec := NewSpecification()
	spec.ContractAddress = ""
	assert.ErrorContains(t, spec.Validate(), "contract address is required")
}

func TestSpecificationWithInvalidContractAddress(t *testing.T) {
	spec := NewSpecification()
	spec.ContractAddress = "invalid"
	assert.ErrorContains(t, spec.Validate(), "contract address is invalid")
}
