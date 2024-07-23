package main

import (
	"fmt"
	"os"

	"github.com/Layr-Labs/eigenlayer-cli/internal/versionupdate"
	"github.com/Layr-Labs/eigenlayer-cli/pkg"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	
	"github.com/urfave/cli/v2"
)

var (
	version = "development"
)

func main() {
	cli.AppHelpTemplate = fmt.Sprintf(`        
     _______ _                   _                              
    (_______|_)                 | |                             
     _____   _  ____  ____ ____ | |      ____ _   _  ____  ____ 
    |  ___) | |/ _  |/ _  )  _ \| |     / _  | | | |/ _  )/ ___)
    | |_____| ( ( | ( (/ /| | | | |____( ( | | |_| ( (/ /| |    
    |_______)_|\_|| |\____)_| |_|_______)_||_|\__  |\____)_|    
              (_____|                        (____/             
    %s`, cli.AppHelpTemplate)
	app := cli.NewApp()

	app.Name = "eigenlayer"
	app.Usage = "EigenLayer CLI"
	app.Version = version
	app.Copyright = "(c) 2024 EigenLabs"

	// Initialize the dependencies
	prompter := utils.NewPrompter()
	app.After = func(c *cli.Context) error {
		versionupdate.Check(app.Version)
		return nil
	}

	app.Commands = append(app.Commands, pkg.OperatorCmd(prompter))
	app.Commands = append(app.Commands, pkg.RewardsCmd(prompter))
	app.Commands = append(app.Commands, pkg.KeysCmd(prompter))

	if err := app.Run(os.Args); err != nil {
		_, err := fmt.Fprintln(os.Stderr, err)
		if err != nil {
			return
		}
		os.Exit(1)
	}
}
