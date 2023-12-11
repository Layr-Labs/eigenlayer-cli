package main

import (
	"fmt"
	"os"

	"github.com/Layr-Labs/eigenlayer-cli/pkg"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/urfave/cli/v2"
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
	app.Version = "0.5.0"
	app.Copyright = "(c) 2023 EigenLabs"

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
