package command

import (
	"github.com/urfave/cli/v2"
)

type BaseCommand interface {
	Execute(c *cli.Context) error
}
