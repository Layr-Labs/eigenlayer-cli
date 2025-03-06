package middleware

import (
	"encoding/json"
	"errors"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/avs/adapters"
)

type Specification struct {
	adapters.BaseSpecification
	ChurnerURL string `json:"churner_url"`
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

	if spec.Coordinator != "middleware" {
		return errors.New("specification: coordinator is invalid")
	}

	return nil
}
