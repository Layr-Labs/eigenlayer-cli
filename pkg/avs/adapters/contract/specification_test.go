package contract

import (
	"testing"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/avs/adapters"
	"github.com/stretchr/testify/assert"
)

func NewTestSpecification() *Specification {
	return &Specification{
		BaseSpecification: adapters.BaseSpecification{
			Name:            "holesky/lagrange-sc",
			Description:     "Lagrange State Committees AVS on Holesky Testnet",
			Network:         "holesky",
			ContractAddress: "0x18A74E66cc90F0B1744Da27E72Df338cEa0A542b",
			Coordinator:     "contract",
			RemoteSigning:   false,
		},
		ABI: "service_manager.json",
		Functions: map[string]Function{
			"register": {
				Name: "register",
				Parameters: []string{
					"config:operator.address",
					"func:bls_sign(type=config:signer_type,file=config:bls_key_file,password=passwd:bls_key_password,hash=call:committee.calculateKeyWithProofHash,salt=last:salt,expiry=last:expiry)->struct(BlsG1PublicKeys:g1,AggG2PublicKey:g2,Signature:signature,Salt:salt,Expiry:expiry)",
					"func:ecdsa_sign(hash=call:avsDirectory.calculateOperatorAVSRegistrationDigestHash,salt=last:salt,expiry=last:expiry)",
				},
			},
			"opt-in": {
				Name: "subscribe",
				Parameters: []string{
					"config:chain_id",
				},
			},
			"opt-out": {
				Name: "unsubscribe",
				Parameters: []string{
					"config:chain_id",
				},
			},
			"deregister": {
				Name:       "deregister",
				Parameters: []string{},
			},
			"status": {
				Name: "avsDirectory.avsOperatorStatus",
			},
		},
		Delegates: []Delegate{
			{
				Name: "avsDirectory",
				ABI:  "avs_directory.json",
				Functions: []Function{
					{
						Name: "calculateOperatorAVSRegistrationDigestHash",
						Parameters: []string{
							"config:operator.address",
							"spec:contract_address",
							"last:salt",
							"last:expiry",
						},
					},
					{
						Name: "avsOperatorStatus",
						Parameters: []string{
							"spec:contract_address",
							"config:operator.address",
						},
					},
				},
			},
			{
				Name: "committee",
				ABI:  "committee.json",
				Functions: []Function{
					{
						Name: "calculateKeyWithProofHash",
						Parameters: []string{
							"config:operator.address",
							"func:salt(seed=const:lagrange-sc)",
							"func:expiry(timeout=const:300)",
						},
					},
				},
			},
		},
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

func TestSpecificationWithMissingABI(t *testing.T) {
	spec := NewTestSpecification()
	spec.ABI = ""
	assert.ErrorContains(t, spec.Validate(), "abi is required")
}

func TestSpecificationWithNilFunctions(t *testing.T) {
	spec := NewTestSpecification()
	spec.Functions = nil
	assert.ErrorContains(t, spec.Validate(), "functions are required")
}

func TestSpecificationWithEmptyFunctions(t *testing.T) {
	spec := NewTestSpecification()
	spec.Functions = map[string]Function{}
	assert.ErrorContains(t, spec.Validate(), "functions are required")
}

func TestSpecificationWithMissingFunctionNameAndMissingFunctionMessage(t *testing.T) {
	spec := NewTestSpecification()
	spec.Functions = map[string]Function{"register": {"", nil, "", ""}}
	assert.ErrorContains(t, spec.Validate(), "message is required when name is not specified")
}

func TestSpecificationWithFunctionParametersWithoutFunctionName(t *testing.T) {
	spec := NewTestSpecification()
	spec.Functions = map[string]Function{"register": {"", []string{"operatorAddress"}, "", ""}}
	assert.ErrorContains(t, spec.Validate(), "name is required when parameters are specified")
}

func TestSpecificationWithFunctionNameAndFunctionMessage(t *testing.T) {
	spec := NewTestSpecification()
	spec.Functions = map[string]Function{"register": {"register", nil, "use opt-in", ""}}
	assert.ErrorContains(t, spec.Validate(), "message cannot be specified when name is specified")
}

func TestSpecificationWithMissingDelegateName(t *testing.T) {
	spec := NewTestSpecification()
	spec.Delegates = []Delegate{
		{"", "registry.json", "", []Function{{"calcX", nil, "", ""}}},
	}
	assert.ErrorContains(t, spec.Validate(), "name is required")
}

func TestSpecificationWithMissingDelegateABI(t *testing.T) {
	spec := NewTestSpecification()
	spec.Delegates = []Delegate{{"registry", "", "", []Function{{"calcY", nil, "", ""}}}}
	assert.ErrorContains(t, spec.Validate(), "abi is required")
}
