package types

import (
	"encoding/json"
	"time"
)

type Actor struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Link        string    `json:"link"`
	Timestamp   time.Time `json:"timestamp"`
}

type NotificationEventConfig struct {
	EventType string                 `json:"eventType"`
	Filters   map[string]interface{} `json:"filters"`
}

type CustomHeaderInput struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type CustomHeaderMasked struct {
	Name        string `json:"name"`
	MaskedValue string `json:"maskedValue"`
}

type NotificationSubscription struct {
	ID                   string                    `json:"id"`
	UserID               string                    `json:"userId"`
	TeamID               string                    `json:"teamId"`
	Name                 string                    `json:"name,omitempty"`
	CreatedAt            time.Time                 `json:"createdAt"`
	UpdatedAt            time.Time                 `json:"updatedAt,omitempty"`
	DeletedAt            *time.Time                `json:"deletedAt,omitempty"`
	CreatedBy            *Actor                    `json:"createdBy,omitempty"`
	EventConfigs         []NotificationEventConfig `json:"eventConfigs"`
	IsEnabled            bool                      `json:"isEnabled"`
	EmailAddress         string                    `json:"emailAddress"`
	WebhookURL           string                    `json:"webhookUrl,omitempty"`
	CustomHeaders        []CustomHeaderMasked      `json:"customHeaders,omitempty"`
	EmailVerified        bool                      `json:"emailVerified,omitempty"`
	RequiresVerification bool                      `json:"requiresVerification,omitempty"`
	Readonly             bool                      `json:"readonly,omitempty"`
}

type NotificationListSubscriptionsResponse struct {
	Subscriptions []NotificationSubscription `json:"subscriptions"`
	TotalCount    int                        `json:"totalCount"`
}

type NotificationEventTypeFilterOption struct {
	Value      string `json:"value"`
	Label      string `json:"label"`
	AppID      string `json:"appId,omitempty"`
	MetricType string `json:"metricType,omitempty"`
}

type NotificationEventTypeFilterField struct {
	Key         string                              `json:"key"`
	Label       string                              `json:"label"`
	Type        string                              `json:"type"`
	Placeholder string                              `json:"placeholder,omitempty"`
	Options     []NotificationEventTypeFilterOption `json:"options,omitempty"`
	Required    bool                                `json:"required"`
}

type NotificationPayloadField struct {
	Key         string `json:"key"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

type NotificationEventType struct {
	Key                  string                             `json:"key"`
	DisplayName          string                             `json:"displayName"`
	Description          string                             `json:"description"`
	Category             string                             `json:"category"`
	Filters              []NotificationEventTypeFilterField `json:"filters,omitempty"`
	WebhookPayloadFormat []NotificationPayloadField         `json:"webhookPayloadFormat,omitempty"`
}

type NotificationEventTypesResponse struct {
	EventTypes []NotificationEventType `json:"eventTypes"`
	Total      int                     `json:"total"`
}

type NotificationEventAttempt struct {
	ID           string          `json:"id"`
	EventID      string          `json:"eventId"`
	AttemptedAt  time.Time       `json:"attemptedAt"`
	Success      bool            `json:"success"`
	StatusCode   *int            `json:"statusCode,omitempty"`
	DurationMs   *int            `json:"durationMs,omitempty"`
	ErrorMessage *string         `json:"errorMessage,omitempty"`
	ResponseBody *string         `json:"responseBody,omitempty"`
	RequestBody  json.RawMessage `json:"requestBody,omitempty"`
}

type NotificationEvent struct {
	ID                    string                     `json:"id"`
	SubscriptionID        string                     `json:"subscriptionId"`
	EventType             string                     `json:"eventType"`
	EventData             json.RawMessage            `json:"eventData"`
	PayloadType           string                     `json:"payloadType"`
	Payload               json.RawMessage            `json:"payload"`
	CreatedAt             time.Time                  `json:"createdAt"`
	SentAt                *time.Time                 `json:"sentAt,omitempty"`
	RetryCount            int                        `json:"retryCount"`
	LastError             *string                    `json:"lastError,omitempty"`
	Status                string                     `json:"status,omitempty"`
	NextRetryAt           *time.Time                 `json:"nextRetryAt,omitempty"`
	Attempts              []NotificationEventAttempt `json:"attempts,omitempty"`
	CreatedByUserID       string                     `json:"createdByUserId,omitempty"`
	SubscriptionDeletedAt *time.Time                 `json:"subscriptionDeletedAt,omitempty"`
}

type NotificationEventsResponse struct {
	Events     []NotificationEvent `json:"events"`
	TotalCount int                 `json:"totalCount"`
}

type NotificationSubscriptionEventsResponse struct {
	Events []NotificationEvent `json:"events"`
}

type NotificationRetryResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type NotificationEmailResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type NotificationWebhookTestResponse struct {
	Success    bool   `json:"success"`
	StatusCode *int   `json:"statusCode,omitempty"`
	DurationMs int64  `json:"durationMs"`
	Error      string `json:"error,omitempty"`
}
