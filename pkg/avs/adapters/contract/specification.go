package contract

import (
	"encoding/json"
	"errors"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/avs/adapters"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
)

type Specification struct {
	adapters.BaseSpecification
	ABI       string              `json:"abi"`
	Functions map[string]Function `json:"functions"`
	Delegates []Delegate          `json:"delegates"`
}

type Function struct {
	Name       string   `json:"name"`
	Parameters []string `json:"parameters"`
	Message    string   `json:"message"`
	Transform  string   `json:"transform"`
}

type Delegate struct {
	Name            string     `json:"name"`
	ABI             string     `json:"abi"`
	ContractAddress string     `json:"contract_address"`
	Functions       []Function `json:"functions"`
}

func NewSpecification(repo adapters.Repository, data []byte) (adapters.Specification, error) {
	spec := Specification{}
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, err
	}

	if err := spec.Validate(); err != nil {
		return nil, err
	}

	return spec, nil
}

func (spec Specification) Validate() error {
	err := spec.BaseSpecification.Validate()
	if err != nil {
		return err
	}

	if spec.Coordinator != "contract" {
		return errors.New("specification: coordinator is invalid")
	}

	if common.IsEmptyString(spec.ABI) {
		return errors.New("specification: abi is required")
	}

	if len(spec.Functions) == 0 {
		return errors.New("specification: functions are required")
	}

	for _, function := range spec.Functions {
		err := function.Validate()
		if err != nil {
			return err
		}
	}

	if len(spec.Delegates) > 0 {
		for _, delegate := range spec.Delegates {
			err := delegate.Validate()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (function Function) Validate() error {
	if common.IsEmptyString(function.Name) {
		if len(function.Parameters) > 0 {
			return errors.New("function: name is required when parameters are specified")
		}

		if common.IsEmptyString(function.Message) {
			return errors.New("function: message is required when name is not specified")
		}
	} else {
		if !common.IsEmptyString(function.Message) {
			return errors.New("function: message cannot be specified when name is specified")
		}
	}

	return nil
}

func (delegate Delegate) Validate() error {
	if common.IsEmptyString(delegate.Name) {
		return errors.New("delegate: name is required")
	}

	if common.IsEmptyString(delegate.ABI) {
		return errors.New("delegate: abi is required")
	}

	return nil
}
