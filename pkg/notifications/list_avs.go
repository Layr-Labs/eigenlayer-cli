package notifications

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/urfave/cli/v2"
)

// fetchAvailableAvsList fetches the list of available AVS services
func fetchAvailableAvsList(ctx context.Context) ([]AvailableAvsItemDto, error) {
	requestURL := fmt.Sprintf("%s%s", getAPIBaseURL(), AvsListEndpoint)

	resp, body, err := makeGetRequest(ctx, requestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch AVS data: %w", err)
	}

	if !isSuccessStatusCode(resp.StatusCode) {
		return nil, handleErrorResponse(resp.StatusCode, body)
	}

	var avsResponse AvailableAvsResponseDto
	if err := json.Unmarshal(body, &avsResponse); err != nil {
		return nil, fmt.Errorf("failed to parse AVS data: %w", err)
	}

	return avsResponse.AvsList, nil
}

// displayAvsList displays a formatted list of available AVS services
func displayAvsList(avsList []AvailableAvsItemDto) {
	fmt.Println("Available AVS Services:")
	fmt.Println("----------------------")

	if len(avsList) == 0 {
		fmt.Println("No AVS services available.")
		return
	}

	for _, avs := range avsList {
		fmt.Printf("Name: %s\n", avs.Name)
		fmt.Printf("Display Name: %s\n", avs.DisplayName)
		fmt.Printf("Description: %s\n", avs.Description)
		fmt.Printf("Status: %s\n", avs.Status)
		fmt.Println("----------------------")
	}
}

// ListAvsCmd returns a command to list available AVS services
func ListAvsCmd() *cli.Command {
	return &cli.Command{
		Name:      "list-avs",
		Usage:     "List all the AVSs available to be subscribed to via notification service",
		UsageText: "list-avs",
		Description: `
This command provides a listing of all AVS services with events available to be subscribed via the notification service.
		`,
		After: telemetry.AfterRunAction(),
		Action: func(cliCtx *cli.Context) error {
			ctx := context.Background()

			// Fetch AVS list
			avsList, err := fetchAvailableAvsList(ctx)
			if err != nil {
				return err
			}

			// Display results
			displayAvsList(avsList)
			return nil
		},
	}
}
