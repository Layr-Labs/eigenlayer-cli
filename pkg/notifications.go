package pkg

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/notifications"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	"github.com/urfave/cli/v2"
)

func NotificationsCmd(p utils.Prompter) *cli.Command {
	var notificationsCmd = &cli.Command{
		Name:  "notifications",
		Usage: "Subscribe and unsubscribe to EigenLayer events via the notification service",
		Subcommands: []*cli.Command{
			notifications.ListEventsCmd(),
			notifications.ListAvsCmd(),
			notifications.SubscribeEventsCmd(),
			notifications.UnsubscribeEventsCmd(),
		},
	}

	return notificationsCmd
}
