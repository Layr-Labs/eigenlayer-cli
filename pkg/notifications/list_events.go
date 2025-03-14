package notifications

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"

	"github.com/urfave/cli/v2"
)

type Event struct {
	Name            string `json:"name"`
	ContractAddress string `json:"contractAddress"`
	EthereumTopic   string `json:"ethereumTopic"`
}

type EventsResponse struct {
	Events []Event `json:"events"`
}

type ErrorResponse struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

const baseURL = "http://localhost:3000/api"

func ListEventsCmd() *cli.Command {
	listCmd := &cli.Command{
		Name:      "list-events",
		Usage:     "List all the events available to be subscribed to via notification service for a given AVS",
		UsageText: "list-events --avs-name <avs-name>",
		Description: `
This command provides a listing of all events available to be subscribed via the notification service.
		`,
		After: telemetry.AfterRunAction(),
		Action: func(context *cli.Context) error {
			avsName := context.String("avs-name")
			if avsName == "" {
				return fmt.Errorf("avs-name is required")
			}

			// Construct URL with query parameter
			reqURL := fmt.Sprintf("%s/available-events?avsName=%s", baseURL, url.QueryEscape(avsName))

			// Create request with headers
			req, err := http.NewRequest("GET", reqURL, nil)
			if err != nil {
				return fmt.Errorf("failed to create request: %v", err)
			}
			req.Header.Set("Accept", "application/json")

			// Send request
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("failed to fetch events data: %v", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response body: %v", err)
			}

			if resp.StatusCode != http.StatusOK {
				var errResp ErrorResponse
				if err := json.Unmarshal(body, &errResp); err != nil {
					return fmt.Errorf("server error: %d - %s", resp.StatusCode, string(body))
				}
				return fmt.Errorf("server error: %s", errResp.Message)
			}

			var eventsResponse EventsResponse
			if err := json.Unmarshal(body, &eventsResponse); err != nil {
				return fmt.Errorf("failed to parse events data: %v", err)
			}

			fmt.Printf("Available Events for AVS '%s':\n", avsName)
			fmt.Println("----------------")
			for _, event := range eventsResponse.Events {
				fmt.Printf("Event Name: %s\n", event.Name)
				fmt.Printf("Contract Address: %s\n", event.ContractAddress)
				fmt.Printf("Ethereum Topic: %s\n", event.EthereumTopic)
				fmt.Println("----------------")
			}

			return nil
		},
		Flags: []cli.Flag{
			&AvsNameFlag,
		},
	}
	return listCmd
}
