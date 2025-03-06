package adapters

import (
	"encoding/json"
	"errors"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigensdk-go/utils"
)

type Specification interface {
	Type() string
	Validate() error
}

type BaseSpecification struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	Network         string `json:"network"`
	ContractAddress string `json:"contract_address"`
	Coordinator     string `json:"coordinator"`
	RemoteSigning   bool   `json:"remote_signing"`
}

func NewBaseSpecification(data []byte) (*BaseSpecification, error) {
	spec := BaseSpecification{}
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, err
	}

	return &spec, nil
}

func (spec BaseSpecification) Type() string {
	return spec.Coordinator
}

func (spec BaseSpecification) Validate() error {
	if common.IsEmptyString(spec.Name) {
		return errors.New("specification: name is required")
	}

	if common.IsEmptyString(spec.Network) {
		return errors.New("specification: network is required")
	}

	if common.IsEmptyString(spec.ContractAddress) {
		return errors.New("specification: contract address is required")
	}

	if !utils.IsValidEthereumAddress(spec.ContractAddress) {
		return errors.New("specification: contract address is invalid")
	}

	if common.IsEmptyString(spec.Coordinator) {
		return errors.New("specification: coordinator is required")
	}

	return nil
}
