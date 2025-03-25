package notifications

import "github.com/urfave/cli/v2"

// Command flags for notification commands
var (
	// AvsNameFlag specifies the AVS name to filter events or subscriptions
	AvsNameFlag = cli.StringFlag{
		Name:     "avs-name",
		Aliases:  []string{"avs"},
		Required: false,
		Usage:    "Specify the AVS name to filter events or create subscriptions",
		EnvVars:  []string{"AVS_NAME"},
	}

	// EventNameFlag specifies the event name for subscriptions
	EventNameFlag = cli.StringFlag{
		Name:     "event-name",
		Aliases:  []string{"event"},
		Required: true,
		Usage:    "Specify the event name to subscribe to",
		EnvVars:  []string{"EVENT_NAME"},
	}

	// OperatorIDFlag specifies the operator ID for subscriptions
	OperatorIDFlag = cli.StringFlag{
		Name:     "operator-id",
		Aliases:  []string{"operator"},
		Required: true,
		Usage:    "Specify the operator ID to filter events or create subscriptions",
		EnvVars:  []string{"OPERATOR_ID"},
	}

	// SubscriptionIDFlag specifies the subscription ID for unsubscribing
	SubscriptionIDFlag = cli.StringFlag{
		Name:     "subscription-id",
		Aliases:  []string{"id"},
		Required: false,
		Usage:    "Specify the subscription ID for unsubscribing from an existing subscription",
		EnvVars:  []string{"SUBSCRIPTION_ID"},
	}

	// DeliveryMethodFlag specifies the delivery method for notifications
	DeliveryMethodFlag = cli.StringFlag{
		Name:     "delivery-method",
		Aliases:  []string{"method"},
		Required: true,
		Usage:    "Specify the delivery method for notifications (email, webhook, or telegram)",
		EnvVars:  []string{"DELIVERY_METHOD"},
	}

	// DeliveryDetailsFlag specifies the delivery details for notifications
	DeliveryDetailsFlag = cli.StringFlag{
		Name:     "delivery-details",
		Aliases:  []string{"details"},
		Required: true,
		Usage:    "Specify the delivery details for notifications (email address, webhook URL, or telegram chat ID)",
		EnvVars:  []string{"DELIVERY_DETAILS"},
	}
)
