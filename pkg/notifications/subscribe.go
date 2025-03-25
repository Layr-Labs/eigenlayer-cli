package notifications

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/urfave/cli/v2"
)

// SubscriptionParams holds parameters for subscription creation
type SubscriptionParams struct {
	AvsName         string
	EventName       string
	OperatorID      string
	DeliveryMethod  string
	DeliveryDetails string
}

// validateSubscribeParams validates all required parameters for subscription
func validateSubscribeParams(cliCtx *cli.Context) (*SubscriptionParams, error) {
	params := &SubscriptionParams{}

	params.AvsName = cliCtx.String("avs-name")
	if params.AvsName == "" {
		return nil, fmt.Errorf("avs-name is required")
	}

	params.EventName = cliCtx.String("event-name")
	if params.EventName == "" {
		return nil, fmt.Errorf("event-name is required")
	}

	params.OperatorID = cliCtx.String("operator-id")
	if params.OperatorID == "" {
		return nil, fmt.Errorf("operator-id is required")
	}

	params.DeliveryMethod = cliCtx.String("delivery-method")
	if err := validateDeliveryMethod(params.DeliveryMethod); err != nil {
		return nil, err
	}

	params.DeliveryDetails = cliCtx.String("delivery-details")
	if params.DeliveryDetails == "" {
		return nil, fmt.Errorf("delivery-details is required")
	}

	return params, nil
}

// displaySubscriptionResponse formats and displays the subscription response
func displaySubscriptionResponse(resp SubscriptionResponseDto, params *SubscriptionParams) {
	fmt.Printf("Successfully subscribed to %s events for AVS '%s'\n", params.EventName, params.AvsName)
	fmt.Printf("Subscription ID: %s\n", resp.SubscriptionID)

	if resp.Message != "" {
		fmt.Printf("Message: %s\n", resp.Message)
	}

	if resp.WorkflowID != "" {
		fmt.Printf("Workflow ID: %s (for webhook delivery)\n", resp.WorkflowID)
	}
}

// createSubscription creates a subscription with the given parameters
func createSubscription(ctx context.Context, params *SubscriptionParams) (*SubscriptionResponseDto, error) {
	// Create request body with the new DTO type
	reqBody := SubscribeDto{
		DeliveryMethod:  params.DeliveryMethod,
		DeliveryDetails: params.DeliveryDetails,
		EventType:       params.EventName,
		AvsName:         params.AvsName,
		OperatorID:      params.OperatorID,
	}

	// Make the request to the subscriptions endpoint
	requestURL := fmt.Sprintf("%s%s", getAPIBaseURL(), SubscriptionsEndpoint)
	resp, body, err := makePostRequest(ctx, requestURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to send subscription request: %w", err)
	}

	// Check for success status code (expect 201 Created)
	if !isSuccessStatusCode(resp.StatusCode) {
		return nil, handleErrorResponse(resp.StatusCode, body)
	}

	// Parse success response
	var subscriptionResp SubscriptionResponseDto
	if err := json.Unmarshal(body, &subscriptionResp); err != nil {
		return nil, fmt.Errorf("failed to parse subscription response: %w", err)
	}

	return &subscriptionResp, nil
}

// SubscribeEventsCmd returns a command to subscribe to events via the notification service
func SubscribeEventsCmd() *cli.Command {
	return &cli.Command{
		Name:      "subscribe",
		Usage:     "Subscribe to an event via the notification service",
		UsageText: "subscribe --avs-name <avs-name> --event-name <event-name> --operator-id <operator-id> --delivery-method <method> --delivery-details <details>",
		Description: `
This command can be used to subscribe to an event via the notification service.
Supported delivery methods are:
- email: use with an email address
- webhook: use with a webhook URL
- telegram: use with a telegram chat ID
		`,
		After: telemetry.AfterRunAction(),
		Action: func(cliCtx *cli.Context) error {
			// Validate all required parameters
			params, err := validateSubscribeParams(cliCtx)
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Create subscription
			subscriptionResp, err := createSubscription(ctx, params)
			if err != nil {
				return err
			}

			// Display results
			displaySubscriptionResponse(*subscriptionResp, params)
			return nil
		},
		Flags: []cli.Flag{
			&AvsNameFlag,
			&EventNameFlag,
			&OperatorIDFlag,
			&DeliveryMethodFlag,
			&DeliveryDetailsFlag,
		},
	}
}
