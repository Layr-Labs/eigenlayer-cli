package operator

import (
	"fmt"
	"strings"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/operator/keys"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	"github.com/urfave/cli/v2"
)

func KeysCmd(p utils.Prompter) *cli.Command {
	var keysCmd = &cli.Command{
		Name:  "keys",
		Usage: "Manage the operator's keys",
		Before: func(context *cli.Context) error {
			deprecationMessage()
			return nil
		},
		After: func(context *cli.Context) error {
			deprecationMessage()
			return nil
		},
		Subcommands: []*cli.Command{
			keys.CreateCmd(p),
			keys.ListCmd(),
			keys.ImportCmd(p),
			keys.ExportCmd(p),
		},
	}

	return keysCmd
}

func deprecationMessage() {
	line1 := "# The keys commands have been moved to the 'keys' subcommand. #"
	line2 := "# Please see 'eigenlayer keys --help' for more information.   #"
	line3 := "# This command will be deprecated in the future.              #"
	fmt.Println(strings.Repeat("#", len(line1)))
	fmt.Println(line1)
	fmt.Println(line2)
	fmt.Println(line3)
	fmt.Println(strings.Repeat("#", len(line1)))
	fmt.Println()
}
