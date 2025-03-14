package notifications

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"

	"github.com/urfave/cli/v2"
)

type AvsInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type AvsListResponse struct {
	AvsList []AvsInfo `json:"avsList"`
}

func ListAvsCmd() *cli.Command {
	listAvsCmd := &cli.Command{
		Name:      "list-avs",
		Usage:     "List all the AVSs available to be subscribed to via notification service",
		UsageText: "list-avs",
		Description: `
This command provides a listing of all AVS with events available to be subscribed via the notification service.
		`,
		After: telemetry.AfterRunAction(),
		Action: func(context *cli.Context) error {
			resp, err := http.Get(baseURL + "/available-avs")
			if err != nil {
				return fmt.Errorf("failed to fetch AVS data: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("server returned non-200 status code: %d", resp.StatusCode)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response body: %v", err)
			}

			var avsResponse AvsListResponse
			if err := json.Unmarshal(body, &avsResponse); err != nil {
				return fmt.Errorf("failed to parse AVS data: %v", err)
			}

			fmt.Println("Available AVS Services:")
			fmt.Println("----------------------")
			for _, avs := range avsResponse.AvsList {
				fmt.Printf("Name: %s\n", avs.Name)
				fmt.Printf("Display Name: %s\n", avs.DisplayName)
				fmt.Printf("Description: %s\n", avs.Description)
				fmt.Printf("Status: %s\n", avs.Status)
				fmt.Println("----------------------")
			}

			return nil
		},
	}
	return listAvsCmd
}
