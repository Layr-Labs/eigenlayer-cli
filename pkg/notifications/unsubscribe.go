package notifications

import (
	"context"
	"fmt"
	"net/url"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/urfave/cli/v2"
)

// validateSubscriptionID validates that a subscription ID is provided
func validateSubscriptionID(cliCtx *cli.Context) (string, error) {
	subscriptionID := cliCtx.String("subscription-id")
	if subscriptionID == "" {
		return "", fmt.Errorf("subscription-id is required")
	}
	return subscriptionID, nil
}

// UnsubscribeEventsCmd returns a command to unsubscribe from notifications
func UnsubscribeEventsCmd() *cli.Command {
	return &cli.Command{
		Name:      "unsubscribe",
		Usage:     "Unsubscribe from notifications via the notification service",
		UsageText: "unsubscribe --subscription-id <id>",
		Description: `
This command can be used to unsubscribe from a specific notification subscription via the notification service.
You must provide the subscription ID of the subscription you wish to unsubscribe from.
		`,
		After: telemetry.AfterRunAction(),
		Action: func(cliCtx *cli.Context) error {
			ctx := context.Background()

			// Validate subscription ID
			subscriptionID, err := validateSubscriptionID(cliCtx)
			if err != nil {
				return err
			}

			// Unsubscribe by ID
			return unsubscribeByID(ctx, subscriptionID)
		},
		Flags: []cli.Flag{
			&SubscriptionIDFlag,
		},
	}
}

// unsubscribeByID unsubscribes from a specific subscription by ID
func unsubscribeByID(ctx context.Context, subscriptionID string) error {
	// Construct URL with the subscription ID
	reqURL := fmt.Sprintf("%s%s", getAPIBaseURL(),
		fmt.Sprintf(SubscriptionByIDEndpointFormat, url.PathEscape(subscriptionID)))

	resp, body, err := makeDeleteRequest(ctx, reqURL)
	if err != nil {
		return fmt.Errorf("failed to send unsubscribe request: %w", err)
	}

	if !isSuccessStatusCode(resp.StatusCode) {
		return handleErrorResponse(resp.StatusCode, body)
	}

	fmt.Printf("successfully unsubscribed from subscription with ID: %s\n", subscriptionID)
	return nil
}
