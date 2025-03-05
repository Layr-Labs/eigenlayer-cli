package plugin

import (
	"encoding/json"
	"errors"
	"plugin"
	"runtime"
	"strings"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/avs/adapters"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigensdk-go/utils"
)

type Specification struct {
	adapters.BaseSpecification
	LibraryURL string `json:"library_url"`
	Plugin     *plugin.Plugin
}

func NewSpecification(repo adapters.Repository, data []byte) (adapters.Specification, error) {
	spec := Specification{}
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, err
	}

	if err := spec.Validate(); err != nil {
		return nil, err
	}

	url := spec.LibraryURL
	url = strings.ReplaceAll(url, "${ARCH}", runtime.GOARCH)
	url = strings.ReplaceAll(url, "${OS}", runtime.GOOS)
	if runtime.GOOS == "windows" {
		url = strings.ReplaceAll(url, "${EXT}", "dll")
	} else {
		url = strings.ReplaceAll(url, "${EXT}", "so")
	}

	plugin, err := repo.LoadPlugin(spec.Name, url)
	if err != nil {
		return nil, err
	}

	spec.Plugin = plugin

	if err := spec.Initialize(data); err != nil {
		return nil, err
	}

	return spec, nil
}

func (spec Specification) Validate() error {
	err := spec.BaseSpecification.Validate()
	if err != nil {
		return err
	}

	if spec.Coordinator != "plugin" {
		return errors.New("specification: coordinator is invalid")
	}

	if common.IsEmptyString(spec.LibraryURL) {
		return errors.New("specification: library url is required")
	}

	if err := utils.CheckBasicURLValidation(spec.LibraryURL); err != nil {
		return errors.New("specification: library url is invalid")
	}

	return nil
}

func (spec *Specification) Initialize(data []byte) error {
	symbol, err := spec.Plugin.Lookup("PluginSpecification")
	if err != nil {
		return nil
	}

	delegate, ok := symbol.(PluginSpecification)
	if !ok {
		return errors.New("library does not contain a valid plugin specification")
	}

	if err := json.Unmarshal(data, &delegate); err != nil {
		return err
	}

	return nil
}
