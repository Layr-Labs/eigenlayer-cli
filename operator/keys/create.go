package keys

import (
	"fmt"
	"github.com/Layr-Labs/eigenlayer-cli/utils/prompts"
	"github.com/urfave/cli/v2"
)

var CreateCmd = &cli.Command{
	Name:   "create",
	Action: Create,
}

func Create(ctx *cli.Context) error {
	password := prompts.GetPasswordModel()
	fmt.Println(password.View())
	return nil
}
