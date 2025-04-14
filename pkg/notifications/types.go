// Package notifications provides functionality for interacting with the EigenLayer notification service.
package notifications

import "time"

// DeliveryMethod represents the available delivery methods for notifications.
type DeliveryMethod string

// Available delivery methods
const (
	DeliveryMethodEmail    DeliveryMethod = "email"
	DeliveryMethodWebhook  DeliveryMethod = "webhook"
	DeliveryMethodTelegram DeliveryMethod = "telegram"
)

// ErrorResponse represents an error response from the API.
type ErrorResponse struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

// AvailableEventItemDto represents an event available for subscription.
type AvailableEventItemDto struct {
	Name            string `json:"name"`
	ContractAddress string `json:"contractAddress"`
	EthereumTopic   string `json:"ethereumTopic"`
}

// AvailableEventsResponseDto represents the API response for available events.
type AvailableEventsResponseDto struct {
	Events []AvailableEventItemDto `json:"events"`
}

// AvailableAvsItemDto represents information about an AVS (Autonomous Validation Service).
type AvailableAvsItemDto struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

// AvailableAvsResponseDto represents the API response for available AVS services.
type AvailableAvsResponseDto struct {
	AvsList []AvailableAvsItemDto `json:"avsList"`
}

// SubscribeDto represents the request body for a subscription.
type SubscribeDto struct {
	DeliveryMethod  string `json:"deliveryMethod"`
	DeliveryDetails string `json:"deliveryDetails"`
	EventType       string `json:"eventType"`
	AvsName         string `json:"avsName"`
	OperatorID      string `json:"operatorId"`
}

// SubscriptionResponseDto represents the API response for a subscription creation.
type SubscriptionResponseDto struct {
	Status         string `json:"status"`
	Message        string `json:"message"`
	SubscriptionID string `json:"subscriptionId"`
	WorkflowID     string `json:"workflowId,omitempty"`
}

// SubscriptionItemDto represents a single subscription item.
type SubscriptionItemDto struct {
	ID              string    `json:"id"`
	DeliveryMethod  string    `json:"deliveryMethod"`
	DeliveryDetails string    `json:"deliveryDetails"`
	EventType       string    `json:"eventType"`
	AvsName         string    `json:"avsName"`
	OperatorID      string    `json:"operatorId"`
	CreatedAt       time.Time `json:"createdAt"`
}

// SubscriptionsResponseDto represents the API response for listing subscriptions.
type SubscriptionsResponseDto struct {
	Subscriptions []SubscriptionItemDto `json:"subscriptions"`
}
