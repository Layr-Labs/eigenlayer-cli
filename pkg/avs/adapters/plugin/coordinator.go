package plugin

import (
	"errors"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/avs/adapters"
	"github.com/Layr-Labs/eigensdk-go/logging"
)

type Coordinator struct {
	repository    adapters.Repository
	logger        logging.Logger
	specification Specification
	config        adapters.Configuration
	delegate      PluginCoordinator
	dryRun        bool
}

func NewCoordinator(
	repository adapters.Repository,
	logger logging.Logger,
	specification Specification,
	configuration adapters.Configuration,
	dryRun bool,
) (*Coordinator, error) {
	coordinator := Coordinator{
		repository, logger, specification, configuration, nil, dryRun,
	}

	if err := coordinator.initialize(); err != nil {
		return nil, err
	}

	return &coordinator, nil
}

func (coordinator *Coordinator) initialize() error {
	symbol, err := coordinator.specification.Plugin.Lookup("PluginConfiguration")
	if symbol != nil && err == nil {
		delegate, ok := symbol.(PluginConfiguration)
		if !ok {
			return errors.New("library does not contain a valid plugin configuration")
		}

		for key, value := range coordinator.config.GetAll() {
			delegate.Set(key, value)
		}
	}

	symbol, err = coordinator.specification.Plugin.Lookup("PluginCoordinator")
	if err != nil {
		return err
	}

	delegate, ok := symbol.(PluginCoordinator)
	if !ok {
		return errors.New("library does not contain a valid plugin coordinator")
	}

	coordinator.delegate = delegate

	return nil
}

func (coordinator Coordinator) Type() string {
	return "plugin"
}

func (coordinator Coordinator) Register() error {
	return coordinator.delegate.Register()
}

func (coordinator Coordinator) OptIn() error {
	return coordinator.delegate.OptIn()
}

func (coordinator Coordinator) OptOut() error {
	return coordinator.delegate.OptOut()
}

func (coordinator Coordinator) Deregister() error {
	return coordinator.delegate.Deregister()
}

func (coordinator Coordinator) Status() (int, error) {
	return coordinator.delegate.Status()
}
