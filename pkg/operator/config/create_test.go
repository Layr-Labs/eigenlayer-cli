package config

import (
	"context"
	"testing"

	prompterMock "github.com/Layr-Labs/eigenlayer-cli/pkg/utils/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"go.uber.org/mock/gomock"
)

func TestCreateCmd_WithYesFlag(t *testing.T) {
	// Arrange
	controller := gomock.NewController(t)
	prompter := prompterMock.NewMockPrompter(controller)

	cmd := CreateCmd(prompter)
	app := cli.NewApp()
	flags := []string{"--yes"}

	// Expect that the prompter will not be called
	prompter.EXPECT().Confirm(gomock.Any()).Times(0)

	// // We do this because the in the parsing of arguments it ignores the first argument
	// // for commands, so we add a blank string as the first argument
	// // I suspect it does this because it is expecting the first argument to be the name of the command
	// // But when we are testing the command, we don't want to have to specify the name of the command
	// // since we are creating the command ourselves
	// // https://github.com/urfave/cli/blob/c023d9bc5a3122830c9355a0a8c17137e0c8556f/command.go#L323
	args := append([]string{""}, flags...)

	cCtx := cli.NewContext(app, nil, &cli.Context{Context: context.Background()})

	// Act
	err := cmd.Run(cCtx, args...)

	// Assert
	assert.NoError(t, err)
	assert.FileExists(t, "operator.yaml")
	assert.FileExists(t, "metadata.json")
}
