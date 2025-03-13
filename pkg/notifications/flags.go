package notifications

import "github.com/urfave/cli/v2"

var (
	AvsNameFlag = cli.StringFlag{
		Name:     "avs-name",
		Aliases:  []string{"k"},
		Required: true,
		Usage:    "Use this flag to specify the AVS name'",
		EnvVars:  []string{"AVS_NAME"},
	}

	EventNameFlag = cli.StringFlag{
		Name:    "event-name",
		Aliases: []string{"i"},
		Usage:   "Use this flag to specify the event name",
		EnvVars: []string{"EVENT_NAME"},
	}

	OperatorIdFlag = cli.StringFlag{
		Name:    "operator-id",
		Aliases: []string{"p"},
		Usage:   "Uset his flag to specify the operator ID to filter supported events by",
		EnvVars: []string{"OPERATOR_ID"},
	}

	SubscriptionIdFlag = cli.StringFlag{
		Name:    "subscription-id",
		Aliases: []string{"s"},
		Usage:   "Use this flag to specify the subscription ID for unsubscribing from an existing subscription",
		EnvVars: []string{"SUBSCRIPTION_ID"},
	}
)
