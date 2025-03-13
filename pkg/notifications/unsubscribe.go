package notifications

import (
	"fmt"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"

	"github.com/urfave/cli/v2"
)

func UnsubscribeEventsCmd() *cli.Command {
	unsubscribeCmd := &cli.Command{
		Name:      "unsubscribe",
		Usage:     "Unsubscribe from an existging subscription via the notification service",
		UsageText: "unsubscribe",
		Description: `
This command can be used to unsubscribe via the notification service.
		`,
		After: telemetry.AfterRunAction(),
		Action: func(context *cli.Context) error {
			fmt.Println("Unsubscribing from ", context.String("subscription-id"))
			return nil
		},
		Flags: []cli.Flag{
			&SubscriptionIdFlag,
		},
	}
	return unsubscribeCmd
}
