package kotsclient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/replicatedhq/replicated/pkg/types"
)

type CreateNotificationSubscriptionRequest struct {
	Name          string                          `json:"name,omitempty"`
	EventConfigs  []types.NotificationEventConfig `json:"eventConfigs"`
	IsEnabled     bool                            `json:"isEnabled"`
	EmailAddress  string                          `json:"emailAddress,omitempty"`
	WebhookURL    string                          `json:"webhookUrl,omitempty"`
	WebhookSecret string                          `json:"webhookSecret,omitempty"`
	CustomHeaders []types.CustomHeaderInput       `json:"customHeaders,omitempty"`
}

type UpdateNotificationSubscriptionRequest struct {
	Name          *string                         `json:"name,omitempty"`
	EventConfigs  []types.NotificationEventConfig `json:"eventConfigs,omitempty"`
	IsEnabled     *bool                           `json:"isEnabled,omitempty"`
	EmailAddress  string                          `json:"emailAddress,omitempty"`
	WebhookURL    string                          `json:"webhookUrl,omitempty"`
	WebhookSecret string                          `json:"webhookSecret,omitempty"`
	CustomHeaders *[]types.CustomHeaderInput      `json:"customHeaders,omitempty"`
}

type VerifyNotificationEmailRequest struct {
	EmailAddress     string `json:"emailAddress"`
	VerificationCode string `json:"verificationCode"`
}

type ResendNotificationVerificationRequest struct {
	EmailAddress string `json:"emailAddress"`
}

type TestNotificationWebhookRequest struct {
	WebhookURL    string                    `json:"webhookUrl"`
	WebhookSecret string                    `json:"webhookSecret,omitempty"`
	EventType     string                    `json:"eventType"`
	CustomHeaders []types.CustomHeaderInput `json:"customHeaders,omitempty"`
}

type ListNotificationEventsOpts struct {
	Type             string
	EventTypes       []string
	SubscriptionID   string
	SubscriptionType string
	StartTime        string
	EndTime          string
	Search           string
	Status           string
	CurrentPage      int
	PageSize         int
}

func (c *VendorV3Client) ListNotificationSubscriptions(search string, subscriptionType string) (*types.NotificationListSubscriptionsResponse, error) {
	v := url.Values{}
	if strings.TrimSpace(search) != "" {
		v.Set("search", search)
	}
	if strings.TrimSpace(subscriptionType) != "" {
		v.Set("type", subscriptionType)
	}

	path := "/v3/notification_subscriptions"
	if encoded := v.Encode(); encoded != "" {
		path = path + "?" + encoded
	}

	resp := &types.NotificationListSubscriptionsResponse{}
	if err := c.DoJSON(context.TODO(), "GET", path, http.StatusOK, nil, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *VendorV3Client) GetNotificationSubscription(id string) (*types.NotificationSubscription, error) {
	resp := &types.NotificationSubscription{}
	if err := c.DoJSON(context.TODO(), "GET", fmt.Sprintf("/v3/notification_subscription/%s", url.PathEscape(id)), http.StatusOK, nil, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *VendorV3Client) CreateNotificationSubscription(req CreateNotificationSubscriptionRequest) (*types.NotificationSubscription, error) {
	resp := &types.NotificationSubscription{}
	if err := c.DoJSON(context.TODO(), "POST", "/v3/notification_subscription", http.StatusCreated, req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *VendorV3Client) UpdateNotificationSubscription(id string, req UpdateNotificationSubscriptionRequest) (*types.NotificationSubscription, error) {
	resp := &types.NotificationSubscription{}
	if err := c.DoJSON(context.TODO(), "PUT", fmt.Sprintf("/v3/notification_subscription/%s", url.PathEscape(id)), http.StatusOK, req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *VendorV3Client) DeleteNotificationSubscription(id string) error {
	return c.DoJSON(context.TODO(), "DELETE", fmt.Sprintf("/v3/notification_subscription/%s", url.PathEscape(id)), http.StatusNoContent, nil, nil)
}

func (c *VendorV3Client) ListNotificationSubscriptionEvents(id string) (*types.NotificationSubscriptionEventsResponse, error) {
	resp := &types.NotificationSubscriptionEventsResponse{}
	if err := c.DoJSON(context.TODO(), "GET", fmt.Sprintf("/v3/notification_subscription/%s/events", url.PathEscape(id)), http.StatusOK, nil, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *VendorV3Client) ListNotificationEvents(opts ListNotificationEventsOpts) (*types.NotificationEventsResponse, error) {
	v := url.Values{}
	if opts.Type != "" {
		v.Set("type", opts.Type)
	}
	for _, eventType := range opts.EventTypes {
		v.Add("event_types", eventType)
	}
	if opts.SubscriptionID != "" {
		v.Set("subscription_id", opts.SubscriptionID)
	}
	if opts.SubscriptionType != "" {
		v.Set("subscription_type", opts.SubscriptionType)
	}
	if opts.StartTime != "" {
		v.Set("start_time", opts.StartTime)
	}
	if opts.EndTime != "" {
		v.Set("end_time", opts.EndTime)
	}
	if opts.Search != "" {
		v.Set("search", opts.Search)
	}
	if opts.Status != "" {
		v.Set("status", opts.Status)
	}
	if opts.CurrentPage > 0 {
		v.Set("currentPage", fmt.Sprintf("%d", opts.CurrentPage))
	}
	if opts.PageSize > 0 {
		v.Set("pageSize", fmt.Sprintf("%d", opts.PageSize))
	}

	path := "/v3/notification_events"
	if encoded := v.Encode(); encoded != "" {
		path = path + "?" + encoded
	}

	resp := &types.NotificationEventsResponse{}
	if err := c.DoJSON(context.TODO(), "GET", path, http.StatusOK, nil, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *VendorV3Client) RetryNotificationEvent(id string) (*types.NotificationRetryResponse, error) {
	resp := &types.NotificationRetryResponse{}
	if err := c.DoJSON(context.TODO(), "POST", fmt.Sprintf("/v3/notification_event/%s/retry", url.PathEscape(id)), http.StatusOK, nil, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *VendorV3Client) ListNotificationEventTypes(query string, limit int) (*types.NotificationEventTypesResponse, error) {
	v := url.Values{}
	if strings.TrimSpace(query) != "" {
		v.Set("q", query)
	}
	if limit > 0 {
		v.Set("limit", fmt.Sprintf("%d", limit))
	}

	path := "/v3/notification_event_types"
	if encoded := v.Encode(); encoded != "" {
		path = path + "?" + encoded
	}

	resp := &types.NotificationEventTypesResponse{}
	if err := c.DoJSON(context.TODO(), "GET", path, http.StatusOK, nil, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *VendorV3Client) ResendNotificationVerification(req ResendNotificationVerificationRequest) (*types.NotificationEmailResponse, error) {
	resp := &types.NotificationEmailResponse{}
	if err := c.DoJSON(context.TODO(), "POST", "/v3/notification_email/resend_verification", http.StatusOK, req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *VendorV3Client) VerifyNotificationEmail(req VerifyNotificationEmailRequest) (*types.NotificationEmailResponse, error) {
	resp := &types.NotificationEmailResponse{}
	if err := c.DoJSON(context.TODO(), "POST", "/v3/notification_email/verify", http.StatusOK, req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *VendorV3Client) TestNotificationWebhook(req TestNotificationWebhookRequest) (*types.NotificationWebhookTestResponse, error) {
	resp := &types.NotificationWebhookTestResponse{}
	if err := c.DoJSON(context.TODO(), "POST", "/v3/notification_webhook/test", http.StatusOK, req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
