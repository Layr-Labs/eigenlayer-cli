package main

import (
	"fmt"
	"github.com/Layr-Labs/eigenlayer-cli/operator"
	"github.com/urfave/cli/v2"
	"os"
)

func main() {
	var app = cli.NewApp()

	app.Name = "eigenlayer"
	app.Usage = "EigenLayer CLI"
	app.Version = "0.1.0"
	app.Commands = append(app.Commands, operator.KeysCmd())

	if err := app.Run(os.Args); err != nil {
		_, err := fmt.Fprintln(os.Stderr, err)
		if err != nil {
			return
		}
		os.Exit(1)
	}
}
