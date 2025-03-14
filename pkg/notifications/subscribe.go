package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"

	"github.com/urfave/cli/v2"
)

type SubscribeRequest struct {
	DeliveryMethod  string `json:"deliveryMethod"`
	DeliveryDetails string `json:"deliveryDetails"`
	EventType       string `json:"eventType"`
	AvsName         string `json:"avsName"`
	OperatorId      string `json:"operatorId"`
}

var DeliveryMethodFlag = cli.StringFlag{
	Name:     "delivery-method",
	Usage:    "Method of delivery for notifications (email, webhook, or telegram)",
	Required: true,
}

var DeliveryDetailsFlag = cli.StringFlag{
	Name:     "delivery-details",
	Usage:    "Details for the delivery method (email address, webhook URL, or telegram chat ID)",
	Required: true,
}

func validateDeliveryMethod(method string) error {
	validMethods := map[string]bool{
		"email":    true,
		"webhook":  true,
		"telegram": true,
	}
	if !validMethods[method] {
		return fmt.Errorf("invalid delivery method: %s. Must be one of: email, webhook, telegram", method)
	}
	return nil
}

func SubscribeEventsCmd() *cli.Command {
	subscribeCmd := &cli.Command{
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
		Action: func(context *cli.Context) error {
			// Validate all required parameters
			avsName := context.String("avs-name")
			if avsName == "" {
				return fmt.Errorf("avs-name is required")
			}

			eventName := context.String("event-name")
			if eventName == "" {
				return fmt.Errorf("event-name is required")
			}

			operatorId := context.String("operator-id")
			if operatorId == "" {
				return fmt.Errorf("operator-id is required")
			}

			deliveryMethod := context.String("delivery-method")
			if err := validateDeliveryMethod(deliveryMethod); err != nil {
				return err
			}

			deliveryDetails := context.String("delivery-details")
			if deliveryDetails == "" {
				return fmt.Errorf("delivery-details is required")
			}

			// Create request body
			reqBody := SubscribeRequest{
				DeliveryMethod:  deliveryMethod,
				DeliveryDetails: deliveryDetails,
				EventType:       eventName,
				AvsName:         avsName,
				OperatorId:      operatorId,
			}

			jsonBody, err := json.Marshal(reqBody)
			if err != nil {
				return fmt.Errorf("failed to create request body: %v", err)
			}

			// Create request
			req, err := http.NewRequest("POST", baseURL+"/subscribe", bytes.NewBuffer(jsonBody))
			if err != nil {
				return fmt.Errorf("failed to create request: %v", err)
			}

			req.Header.Set("Accept", "application/json")
			req.Header.Set("Content-Type", "application/json")

			// Send request
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("failed to send subscription request: %v", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response body: %v", err)
			}

			// Accept both 200 and 201 as success codes
			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
				var errResp ErrorResponse
				if err := json.Unmarshal(body, &errResp); err != nil {
					return fmt.Errorf("server error: %d - %s", resp.StatusCode, string(body))
				}
				return fmt.Errorf("server error: %s", errResp.Message)
			}

			fmt.Printf("Successfully subscribed to %s events for AVS '%s'\n", eventName, avsName)
			return nil
		},
		Flags: []cli.Flag{
			&AvsNameFlag,
			&EventNameFlag,
			&OperatorIdFlag,
			&DeliveryMethodFlag,
			&DeliveryDetailsFlag,
		},
	}
	return subscribeCmd
}
