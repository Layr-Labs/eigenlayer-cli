package avs

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils/codec"
	"github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

const (
	AVSConfigTemplateFile = "config.yaml"
	OperatorConfigFile    = "operator.yaml"
	OperatorMetaFile      = "metadata.json"
)

type Configuration struct {
	logger   logging.Logger
	prompter utils.Prompter
	registry map[string]interface{}
	codec    *codec.JSONCodec
}

func NewConfiguration(
	ctx *cli.Context,
	prompter utils.Prompter,
	operatorConfig string,
	avsConfig string,
	overrides []string,
) (*Configuration, error) {
	config := Configuration{
		codec:    codec.NewJSONCodec(),
		logger:   common.GetLogger(ctx),
		prompter: prompter,
	}

	if err := config.loadFromFile(operatorConfig); err != nil {
		return nil, err
	}

	if err := config.loadFromFile(avsConfig); err != nil {
		return nil, err
	}

	if err := config.loadFromOverrides(overrides); err != nil {
		return nil, err
	}

	return &config, nil
}

func (provider *Configuration) Unmarshal(cfg any) error {
	data, err := json.Marshal(provider.registry)
	if err != nil {
		return err
	}

	if err = provider.codec.Unmarshal(data, cfg); err != nil {
		return err
	}

	return nil
}

func (provider *Configuration) UnmarshalYaml(cfg any) error {
	data, err := yaml.Marshal(provider.registry)
	if err != nil {
		return err
	}

	if err = provider.codec.Unmarshal(data, cfg); err != nil {
		return err
	}

	return nil
}

func (provider *Configuration) loadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("%s [path=%s]", err.Error(), path)
	}

	if err := yaml.Unmarshal(data, &provider.registry); err != nil {
		return fmt.Errorf("%w [path=%s]", ErrFailedToLoadFile, path)
	}

	return nil
}

func (provider *Configuration) loadFromOverrides(overrides []string) error {
	for _, entry := range overrides {
		tokens := strings.Split(entry, "=")
		if len(tokens) == 2 {
			provider.registry[strings.TrimSpace(tokens[0])] = strings.TrimSpace(tokens[1])
		}

	}

	return nil
}

func (provider *Configuration) Prompt(key string, required bool, hidden bool) (interface{}, error) {
	prompt := "Enter " + strings.ReplaceAll(key, "_", " ") + ":"
	validator := func(v string) error {
		if required && len(v) == 0 {
			return fmt.Errorf("valid value is required")
		}

		return nil
	}

	if hidden {
		return provider.prompter.InputHiddenString(prompt, "", validator)
	} else {
		return provider.prompter.InputString(prompt, "", "", validator)
	}
}

func (provider *Configuration) Get(configName string) (interface{}, error) {
	config, exists := provider.registry[configName]
	if !exists {
		return "", ErrInvalidConfig
	}

	switch v := config.(type) {
	case string:
		return v, nil
	case int:
		return v, nil
	case float64:
		return v, nil
	case map[string]interface{}:
		return v, nil
	default:
		return "", ErrInvalidConfigType
	}
}

func (provider *Configuration) GetAll() map[string]interface{} {
	return provider.registry
}

func ConfigCmd(p utils.Prompter) *cli.Command {
	configCmd := &cli.Command{
		Name:  "config",
		Usage: "Manage AVS specific configuration",
		Subcommands: []*cli.Command{
			CreateCmd(p),
		},
	}

	return configCmd
}

func CreateCmd(p utils.Prompter) *cli.Command {
	createCmd := &cli.Command{
		Name:      "create",
		Usage:     "Create an AVS specific configuration file",
		UsageText: "create <avs-id> [avs-config-file]",
		Description: `
		This command will create an AVS specific empty configuration file based
		on the corresponding specifications.
		`,
		After: telemetry.AfterRunAction(),
		Action: func(context *cli.Context) error {
			logger := common.GetLogger(context)

			args := context.Args().Slice()
			if err := validateCreateCmdInput(args); err != nil {
				return err
			}

			avsId, err := getNextArg(&args, "avsId", "")
			if err != nil {
				return err
			}

			if err := validateAVSId(avsId); err != nil {
				return err
			}

			avsConfigDefault := strings.Replace(avsId, "/", "-", -1) + "-" + AVSConfigTemplateFile
			avsConfig, err := getNextArg(&args, "avsConfig", avsConfigDefault)
			if err != nil {
				return err
			}

			if err := validateAVSConfig(avsConfig); err != nil {
				logger.Errorf("Failed to validate configuration template: %s", err.Error())
				return failed(err)
			}

			if err := createAVSConfigFile(p, avsId, avsConfig); err != nil {
				logger.Errorf("Failed to create configuration file: %s", err.Error())
				return failed(err)
			}

			logger.Infof("%s Configuration file %s for %s AVS created", utils.EmojiCheckMark, avsConfig, avsId)

			return nil
		},
	}

	return createCmd
}

func getNextArg(args *[]string, argName string, defaultValue string) (string, error) {
	argsLength := len(*args)
	argValue := ""
	if argsLength == 0 {
		if len(defaultValue) == 0 {
			return argValue, fmt.Errorf("%w: argument %s", ErrEmptyArgValue, argName)
		} else {
			argValue = defaultValue
		}
	} else {
		argValue = (*args)[0]
		*args = (*args)[1:]
	}

	return argValue, nil
}

func fileExists(filePath string) bool {
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		return true
	}
}

func createAVSConfigFile(p utils.Prompter, avsId string, avsConfig string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	avsSpecDir := filepath.Join(RepositorySubFolder, avsId)
	avsConfigTemplatePath := filepath.Join(home, avsSpecDir, AVSConfigTemplateFile)
	yamlData, err := os.ReadFile(avsConfigTemplatePath)
	if err != nil {
		return fmt.Errorf("%w: config template file expected in directory %s", err, avsSpecDir)
	}

	values := make(map[string]string)
	if err := yaml.Unmarshal(yamlData, &values); err != nil {
		return fmt.Errorf("%w: invalid yaml in config template %s", err, avsConfigTemplatePath)
	}

	id, exists := values["avs_id"]
	if !exists {
		return fmt.Errorf("missing avs_id in config template %s", avsConfigTemplatePath)
	}

	if id != avsId {
		return fmt.Errorf("avs_id is invalid in config template %s", avsConfigTemplatePath)
	}

	isConfigExist := fileExists(avsConfig)

	confirm := false
	if isConfigExist {
		confirm, err = p.Confirm("This will overwrite existing avs config file. Are you sure you want to continue?")
		if err != nil {
			return err
		}
	}

	if !isConfigExist || isConfigExist && confirm {
		err = common.WriteToFile(yamlData, avsConfig)
		if err != nil {
			return err
		}
	}
	return nil
}

func validateCreateCmdInput(args []string) error {
	argsLength := len(args)
	if argsLength < 1 || argsLength > 2 {
		return fmt.Errorf("%w: accepts 1 arg, received %d", ErrInvalidNumberOfArgs, argsLength)
	}

	return nil
}

func validateAVSId(avsId string) error {
	if len(avsId) == 0 {
		return fmt.Errorf("%w: provided avs id is empty", ErrEmptyArgValue)
	}

	if match, _ := regexp.MatchString("\\s", avsId); match {
		return fmt.Errorf("%w: provided avs id is %s", ErrArgValueContainsWhitespaces, avsId)
	}

	return nil
}

func validateAVSConfig(avsConfig string) error {
	if len(avsConfig) == 0 {
		return fmt.Errorf("%w: provided avs config is empty", ErrEmptyArgValue)
	}

	if match, _ := regexp.MatchString("\\s", avsConfig); match {
		return fmt.Errorf("%w: provided avs config is %s", ErrArgValueContainsWhitespaces, avsConfig)
	}

	if !strings.HasSuffix(avsConfig, ".yaml") && !strings.HasSuffix(avsConfig, ".yml") {
		return fmt.Errorf("%w: provided avs config is %s", ErrInvalidConfigFile, avsConfig)
	}

	return nil
}
