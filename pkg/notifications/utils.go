package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// API Endpoints based on the RESTful API
const (
	// APIBaseURL is the base URL for the notification API
	APIBaseURL = "http://localhost:3000/api"

	// AvsListEndpoint is the endpoint for listing AVS services
	AvsListEndpoint = "/avs"

	// AvsEventsEndpointFormat is the format string for the AVS events endpoint
	AvsEventsEndpointFormat = "/avs/%s/events" // needs formatting with avsName

	// SubscriptionsEndpoint is the endpoint for managing subscriptions
	SubscriptionsEndpoint = "/subscriptions"

	// SubscriptionByIDEndpointFormat is the format string for specific subscription endpoints
	SubscriptionByIDEndpointFormat = "/subscriptions/%s" // needs formatting with subscriptionID
)

// defaultTimeout is the default timeout for HTTP requests
const defaultTimeout = 10 * time.Second

// httpClientOnce ensures the HTTP client is created only once
var httpClientOnce sync.Once

// httpClient is the shared HTTP client for all requests
var httpClient *http.Client

// RequestConfig holds configuration for HTTP requests
type RequestConfig struct {
	Method      string
	URL         string
	Body        interface{}
	ContentType string
	Accept      string
}

// getHTTPClient returns the shared HTTP client instance
func getHTTPClient() *http.Client {
	httpClientOnce.Do(func() {
		httpClient = &http.Client{
			Timeout: defaultTimeout,
		}
	})
	return httpClient
}

// getAPIBaseURL returns the API base URL, with potential environment override
func getAPIBaseURL() string {
	// TODO: Make this configurable via environment variable or config file
	return APIBaseURL
}

// doHTTPRequest performs the HTTP request and processes the response
func doHTTPRequest(ctx context.Context, req *http.Request) (*http.Response, []byte, error) {
	resp, err := getHTTPClient().Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return resp, body, nil
}

// createRequest creates an HTTP request with proper headers
func createRequest(ctx context.Context, config RequestConfig) (*http.Request, error) {
	var bodyReader io.Reader

	if config.Body != nil {
		jsonBody, err := json.Marshal(config.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, config.Method, config.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	accept := config.Accept
	if accept == "" {
		accept = "application/json"
	}
	req.Header.Set("Accept", accept)

	if bodyReader != nil {
		contentType := config.ContentType
		if contentType == "" {
			contentType = "application/json"
		}
		req.Header.Set("Content-Type", contentType)
	}

	return req, nil
}

// makeRequest makes an HTTP request using the provided configuration
func makeRequest(ctx context.Context, config RequestConfig) (*http.Response, []byte, error) {
	req, err := createRequest(ctx, config)
	if err != nil {
		return nil, nil, err
	}

	return doHTTPRequest(ctx, req)
}

// makeGetRequest makes a GET request to the specified URL with proper headers
func makeGetRequest(ctx context.Context, url string) (*http.Response, []byte, error) {
	config := RequestConfig{
		Method: http.MethodGet,
		URL:    url,
	}

	return makeRequest(ctx, config)
}

// makePostRequest makes a POST request to the specified URL with the given body and proper headers
func makePostRequest(ctx context.Context, url string, requestBody interface{}) (*http.Response, []byte, error) {
	config := RequestConfig{
		Method: http.MethodPost,
		URL:    url,
		Body:   requestBody,
	}

	return makeRequest(ctx, config)
}

// makeDeleteRequest makes a DELETE request to the specified URL with proper headers
func makeDeleteRequest(ctx context.Context, url string) (*http.Response, []byte, error) {
	config := RequestConfig{
		Method: http.MethodDelete,
		URL:    url,
	}

	return makeRequest(ctx, config)
}

// makeDeleteRequestWithBody makes a DELETE request with a JSON body to the specified URL
func makeDeleteRequestWithBody(
	ctx context.Context,
	url string,
	requestBody interface{},
) (*http.Response, []byte, error) {
	config := RequestConfig{
		Method: http.MethodDelete,
		URL:    url,
		Body:   requestBody,
	}

	return makeRequest(ctx, config)
}

// handleErrorResponse parses and returns a formatted error from an API error response
func handleErrorResponse(statusCode int, body []byte) error {
	var errResp ErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return fmt.Errorf("server error: %d - %s", statusCode, string(body))
	}
	return fmt.Errorf("server error: %s", errResp.Message)
}

// isSuccessStatusCode checks if the given status code indicates success
func isSuccessStatusCode(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}

// validateDeliveryMethod checks if the provided delivery method is valid
func validateDeliveryMethod(method string) error {
	validMethods := map[string]bool{
		string(DeliveryMethodEmail):    true,
		string(DeliveryMethodWebhook):  true,
		string(DeliveryMethodTelegram): true,
	}
	if !validMethods[method] {
		return fmt.Errorf("invalid delivery method: %s. Must be one of: email, webhook, telegram", method)
	}
	return nil
}
