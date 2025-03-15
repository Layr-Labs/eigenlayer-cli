package avs

import (
	"fmt"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/avs/adapters"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/urfave/cli/v2"
)

func NewFlow(context *cli.Context, prompter utils.Prompter) (adapters.Coordinator, error) {
	logger := common.GetLogger(context)
	logger.Debug("Starting new AVS workflow")

	logger.Debug(fmt.Sprintf("Validating arguments [Arguments=%+v]", context.Args()))
	args := context.Args()
	if args.Len() != 2 {
		return nil, fmt.Errorf("%w: accepts 2 args, received %d", ErrInvalidNumberOfArgs, args.Len())
	}

	logger.Debug(fmt.Sprintf("Loading configuration: %s", args.Get(0)))
	logger.Debug(fmt.Sprintf("Loading configuration: %s", args.Get(1)))
	configuration, err := NewConfiguration(
		context,
		prompter,
		args.Get(0),
		args.Get(1),
		context.StringSlice(ArgFlag.Name),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	logger.Debug("Initializing specification repository")
	repository, err := NewRepository(context)
	if err != nil {
		return nil, err
	}

	avs, err := configuration.Get("avs_id")
	if err != nil {
		return nil, err
	}

	logger.Info(fmt.Sprintf("AVS: %s", avs))
	logger.Debug(fmt.Sprintf("Loading specification [AVS=%s]", avs.(string)))
	specification, err := NewSpecification(repository, avs.(string))
	if err != nil {
		return nil, err
	}

	logger.Debug(fmt.Sprintf("Specification: %+v", specification))
	logger.Debug("Loading coordinator")

	dryRun := context.Bool(flags.DryRunFlag.Name)
	if dryRun {
		logger.Debug("Dry run mode enabled")
	}

	return NewCoordinator(repository, logger, specification, *configuration, dryRun)
}
