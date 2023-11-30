package main

import (
	"fmt"
	"os"

	"github.com/Layr-Labs/eigenlayer-cli/pkg"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/urfave/cli/v2"
)

func main() {
	var app = cli.NewApp()

	app.Name = "eigenlayer"
	app.Usage = "EigenLayer CLI"
	app.Version = "0.1.0"

	// Initialize the dependencies
	prompter := utils.NewPrompter()
	app.Commands = append(app.Commands, pkg.OperatorCmd(prompter))

	if err := app.Run(os.Args); err != nil {
		_, err := fmt.Fprintln(os.Stderr, err)
		if err != nil {
			return
		}
		os.Exit(1)
	}
}
