package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/urfave/cli/v2"
)

// fetchAvailableEvents fetches available events for a given AVS
func fetchAvailableEvents(ctx context.Context, avsName string) ([]AvailableEventItemDto, error) {
	// Use the avs/{avsName}/events endpoint
	reqURL := fmt.Sprintf("%s%s", getAPIBaseURL(),
		fmt.Sprintf(AvsEventsEndpointFormat, url.PathEscape(avsName)))

	resp, body, err := makeGetRequest(ctx, reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch events data: %w", err)
	}

	if !isSuccessStatusCode(resp.StatusCode) {
		return nil, handleErrorResponse(resp.StatusCode, body)
	}

	var eventsResponse AvailableEventsResponseDto
	if err := json.Unmarshal(body, &eventsResponse); err != nil {
		return nil, fmt.Errorf("failed to parse events data: %w", err)
	}

	return eventsResponse.Events, nil
}

// displayAvailableEvents displays a formatted list of available events
func displayAvailableEvents(events []AvailableEventItemDto, avsName string) {
	fmt.Printf("Available Events for AVS '%s':\n", avsName)
	fmt.Println("----------------")

	if len(events) == 0 {
		fmt.Println("No events found for this AVS.")
		return
	}

	for _, event := range events {
		fmt.Printf("Event Name: %s\n", event.Name)
		fmt.Printf("Contract Address: %s\n", event.ContractAddress)
		fmt.Printf("Ethereum Topic: %s\n", event.EthereumTopic)
		fmt.Println("----------------")
	}
}

// ListEventsCmd returns a command to list available events for a given AVS
func ListEventsCmd() *cli.Command {
	return &cli.Command{
		Name:      "list-events",
		Usage:     "List all the events available to be subscribed to via notification service for a given AVS",
		UsageText: "list-events --avs-name <avs-name>",
		Description: `
This command provides a listing of all events available to be subscribed via the notification service.
The AVS name is required to filter events by a specific AVS.
		`,
		After: telemetry.AfterRunAction(),
		Action: func(cliCtx *cli.Context) error {
			avsName := cliCtx.String("avs-name")
			if avsName == "" {
				return fmt.Errorf("avs-name is required")
			}

			ctx := context.Background()

			// Fetch events
			events, err := fetchAvailableEvents(ctx, avsName)
			if err != nil {
				return err
			}

			// Display results
			displayAvailableEvents(events, avsName)
			return nil
		},
		Flags: []cli.Flag{
			&AvsNameFlag,
		},
	}
}
