package avs

import (
	"errors"
	"fmt"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/avs/adapters"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/avs/adapters/contract"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/avs/adapters/middleware"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/avs/adapters/plugin"
	"github.com/Layr-Labs/eigensdk-go/logging"
)

func NewSpecification(repo *Repository, name string) (adapters.Specification, error) {
	data, err := repo.LoadResource(name, "avs.json")
	if err != nil {
		return nil, err
	}

	base, err := adapters.NewBaseSpecification(data)
	if err != nil {
		return nil, err
	}

	if name != base.Name {
		return nil, fmt.Errorf("specification name does not match repository location: %s", name)
	}

	switch {
	case base.Type() == "contract":
		return contract.NewSpecification(repo, data)
	case base.Type() == "middleware":
		return middleware.NewSpecification(repo, data)
	case base.Type() == "plugin":
		return plugin.NewSpecification(repo, data)
	}

	return nil, errors.New("specification kind is invalid")
}

func NewCoordinator(
	repository *Repository,
	logger logging.Logger,
	spec adapters.Specification,
	config Configuration,
	dryRun bool,
) (adapters.Coordinator, error) {
	switch {
	case spec.Type() == "contract":
		return contract.NewCoordinator(repository, logger, spec.(contract.Specification), &config, dryRun)
	case spec.Type() == "middleware":
		return middleware.NewCoordinator(repository, logger, spec.(middleware.Specification), &config, dryRun), nil
	case spec.Type() == "plugin":
		return plugin.NewCoordinator(repository, logger, spec.(plugin.Specification), &config, dryRun)
	}

	return nil, errors.New("coordinator kind is invalid")
}
