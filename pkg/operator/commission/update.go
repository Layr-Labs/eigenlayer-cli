package commission

import (
	"fmt"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/urfave/cli/v2"
)

func UpdateCmd(p utils.Prompter) *cli.Command {
	var updateCmd = &cli.Command{
		Name:  "update",
		Usage: "Update the operator's commission",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "commission",
				Usage: "The commission rate to set",
			},
		},
		Action: func(c *cli.Context) error {
			commission := c.String("commission")
			if commission == "" {
				return fmt.Errorf("commission rate is required")
			}

			return nil
		},
	}

	return updateCmd
}
