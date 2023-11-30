package keys

import (
	"fmt"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/urfave/cli/v2"
)

func CreateCmd(p utils.Prompter) *cli.Command {
	createCmd := &cli.Command{
		Name: "create",
		Action: func(context *cli.Context) error {
			confirm, err := p.Confirm("Would you like to populate the operator config file?")
			if err != nil {
				return err
			}
			fmt.Println(confirm)
			return nil
		},
	}
	return createCmd
}
