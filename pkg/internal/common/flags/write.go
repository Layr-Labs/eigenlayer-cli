package flags

import "github.com/urfave/cli/v2"

var WriteFlags = []cli.Flag{
	&OutputFileFlag,
	&OutputTypeFlag,
	&CallerAddressFlag,
	&BroadcastFlag,
}
