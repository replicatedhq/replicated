package kotsclient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestNotificationSubscriptionClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "test-token", r.Header.Get("Authorization"))

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v3/notification_subscriptions":
			require.Equal(t, "release", r.URL.Query().Get("search"))
			require.Equal(t, "team", r.URL.Query().Get("type"))
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"subscriptions": []map[string]interface{}{
					{
						"id":         "sub-1",
						"userId":     "user-1",
						"teamId":     "team-1",
						"name":       "Release webhook",
						"isEnabled":  true,
						"webhookUrl": "https://example.com/hook",
						"eventConfigs": []map[string]interface{}{
							{"eventType": "release.promoted", "filters": map[string]interface{}{}},
						},
					},
				},
				"totalCount": 1,
			})
		case r.Method == http.MethodPost && r.URL.Path == "/v3/notification_subscription":
			var body CreateNotificationSubscriptionRequest
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			require.Equal(t, "Release webhook", body.Name)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":           "sub-1",
				"userId":       "user-1",
				"teamId":       "team-1",
				"name":         body.Name,
				"isEnabled":    body.IsEnabled,
				"webhookUrl":   body.WebhookURL,
				"eventConfigs": body.EventConfigs,
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v3/notification_subscription/sub-1":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":         "sub-1",
				"userId":     "user-1",
				"teamId":     "team-1",
				"name":       "Release webhook",
				"isEnabled":  true,
				"webhookUrl": "https://example.com/hook",
				"eventConfigs": []map[string]interface{}{
					{"eventType": "release.promoted", "filters": map[string]interface{}{}},
				},
			})
		case r.Method == http.MethodPut && r.URL.Path == "/v3/notification_subscription/sub-1":
			var body UpdateNotificationSubscriptionRequest
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			require.NotNil(t, body.IsEnabled)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":         "sub-1",
				"userId":     "user-1",
				"teamId":     "team-1",
				"name":       "Release webhook",
				"isEnabled":  *body.IsEnabled,
				"webhookUrl": "https://example.com/hook",
				"eventConfigs": []map[string]interface{}{
					{"eventType": "release.promoted", "filters": map[string]interface{}{}},
				},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/v3/notification_subscription/sub-1":
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	api := platformclient.NewHTTPClient(server.URL, "test-token")
	client := VendorV3Client{HTTPClient: *api}

	listResp, err := client.ListNotificationSubscriptions("release", "team")
	require.NoError(t, err)
	require.Equal(t, 1, listResp.TotalCount)

	created, err := client.CreateNotificationSubscription(CreateNotificationSubscriptionRequest{
		Name:       "Release webhook",
		IsEnabled:  true,
		WebhookURL: "https://example.com/hook",
		EventConfigs: []types.NotificationEventConfig{
			{EventType: "release.promoted", Filters: map[string]interface{}{}},
		},
	})
	require.NoError(t, err)
	require.Equal(t, "sub-1", created.ID)

	got, err := client.GetNotificationSubscription("sub-1")
	require.NoError(t, err)
	require.Equal(t, "Release webhook", got.Name)

	enabled := false
	updated, err := client.UpdateNotificationSubscription("sub-1", UpdateNotificationSubscriptionRequest{
		IsEnabled: &enabled,
	})
	require.NoError(t, err)
	require.False(t, updated.IsEnabled)

	require.NoError(t, client.DeleteNotificationSubscription("sub-1"))
}

func TestNotificationActionClients(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v3/notification_events":
			require.Equal(t, "failed", r.URL.Query().Get("status"))
			require.Equal(t, "sub-1", r.URL.Query().Get("subscription_id"))
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"events": []map[string]interface{}{
					{
						"id":             "evt-1",
						"subscriptionId": "sub-1",
						"eventType":      "release.promoted",
						"eventData":      map[string]interface{}{"releaseId": "rel-1"},
						"payloadType":    "webhook",
						"payload":        map[string]interface{}{"event": "release.promoted"},
						"retryCount":     8,
						"status":         "failed",
						"attempts":       []map[string]interface{}{},
					},
				},
				"totalCount": 1,
			})
		case "/v3/notification_event_types":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"eventTypes": []map[string]interface{}{
					{
						"key":         "release.promoted",
						"displayName": "Release Promoted",
						"description": "A release was promoted.",
						"category":    "Release",
					},
				},
				"total": 1,
			})
		case "/v3/notification_event/evt-1/retry":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "queued"})
		case "/v3/notification_email/resend_verification":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "resent"})
		case "/v3/notification_email/verify":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "verified"})
		case "/v3/notification_webhook/test":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "statusCode": 200, "durationMs": 12})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	api := platformclient.NewHTTPClient(server.URL, "test-token")
	client := VendorV3Client{HTTPClient: *api}

	eventsResp, err := client.ListNotificationEvents(ListNotificationEventsOpts{
		Status:         "failed",
		SubscriptionID: "sub-1",
	})
	require.NoError(t, err)
	require.Equal(t, 1, eventsResp.TotalCount)

	eventTypesResp, err := client.ListNotificationEventTypes("release", 20)
	require.NoError(t, err)
	require.Equal(t, 1, eventTypesResp.Total)

	retryResp, err := client.RetryNotificationEvent("evt-1")
	require.NoError(t, err)
	require.True(t, retryResp.Success)

	resendResp, err := client.ResendNotificationVerification(ResendNotificationVerificationRequest{EmailAddress: "alerts@example.com"})
	require.NoError(t, err)
	require.True(t, resendResp.Success)

	verifyResp, err := client.VerifyNotificationEmail(VerifyNotificationEmailRequest{
		EmailAddress:     "alerts@example.com",
		VerificationCode: "123456",
	})
	require.NoError(t, err)
	require.True(t, verifyResp.Success)

	webhookResp, err := client.TestNotificationWebhook(TestNotificationWebhookRequest{
		WebhookURL: "https://example.com/hook",
		EventType:  "release.promoted",
	})
	require.NoError(t, err)
	require.True(t, webhookResp.Success)
	require.NotNil(t, webhookResp.StatusCode)
}
