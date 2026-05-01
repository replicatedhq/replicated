package kotsclient

import (
	"fmt"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	realkotsclient "github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/stretchr/testify/assert"
)

func Test_ListNotificationSubscriptions(t *testing.T) {
	test := func() error {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)
		api := platformclient.NewHTTPClient(u, "replicated-cli-list-notifications-token")
		client := realkotsclient.VendorV3Client{HTTPClient: *api}

		resp, err := client.ListNotificationSubscriptions("release", "team")
		assert.NoError(t, err)
		assert.Equal(t, 1, resp.TotalCount)
		assert.Len(t, resp.Subscriptions, 1)
		assert.Equal(t, "Release webhook", resp.Subscriptions[0].Name)

		return nil
	}

	pact.AddInteraction().
		Given("List notification subscriptions").
		UponReceiving("A request to list notification subscriptions").
		WithRequest(dsl.Request{
			Method: "GET",
			Path:   dsl.String("/v3/notification_subscriptions"),
			Query: dsl.MapMatcher{
				"search": dsl.String("release"),
				"type":   dsl.String("team"),
			},
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-list-notifications-token"),
				"Content-Type":  dsl.String("application/json"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"subscriptions": []map[string]interface{}{
					{
						"id":         dsl.Like("notif-sub-1"),
						"userId":     dsl.Like("user-1"),
						"teamId":     dsl.Like("team-1"),
						"name":       dsl.String("Release webhook"),
						"isEnabled":  true,
						"webhookUrl": dsl.String("https://example.com/hook"),
						"eventConfigs": []map[string]interface{}{
							{
								"eventType": dsl.String("release.promoted"),
								"filters":   map[string]interface{}{},
							},
						},
					},
				},
				"totalCount": dsl.Like(1),
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func Test_CreateUpdateGetDeleteNotificationSubscription(t *testing.T) {
	test := func() error {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)
		api := platformclient.NewHTTPClient(u, "replicated-cli-subscription-token")
		client := realkotsclient.VendorV3Client{HTTPClient: *api}

		created, err := client.CreateNotificationSubscription(realkotsclient.CreateNotificationSubscriptionRequest{
			Name:       "Release webhook",
			IsEnabled:  true,
			WebhookURL: "https://example.com/hook",
			EventConfigs: []types.NotificationEventConfig{
				{
					EventType: "release.promoted",
					Filters:   map[string]interface{}{},
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, "notif-sub-1", created.ID)

		got, err := client.GetNotificationSubscription("notif-sub-1")
		assert.NoError(t, err)
		assert.Equal(t, "Release webhook", got.Name)

		enabled := false
		updated, err := client.UpdateNotificationSubscription("notif-sub-1", realkotsclient.UpdateNotificationSubscriptionRequest{
			IsEnabled: &enabled,
		})
		assert.NoError(t, err)
		assert.False(t, updated.IsEnabled)

		err = client.DeleteNotificationSubscription("notif-sub-1")
		assert.NoError(t, err)

		return nil
	}

	eventConfigs := []map[string]interface{}{
		{
			"eventType": "release.promoted",
			"filters":   map[string]interface{}{},
		},
	}

	pact.AddInteraction().
		Given("Create notification subscription").
		UponReceiving("A request to create a notification subscription").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/v3/notification_subscription"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-subscription-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"name":         "Release webhook",
				"isEnabled":    true,
				"webhookUrl":   "https://example.com/hook",
				"eventConfigs": eventConfigs,
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 201,
			Body: map[string]interface{}{
				"id":           dsl.Like("notif-sub-1"),
				"userId":       dsl.Like("user-1"),
				"teamId":       dsl.Like("team-1"),
				"name":         dsl.String("Release webhook"),
				"isEnabled":    true,
				"webhookUrl":   dsl.String("https://example.com/hook"),
				"eventConfigs": eventConfigs,
			},
		})

	pact.AddInteraction().
		Given("Get notification subscription").
		UponReceiving("A request to get a notification subscription").
		WithRequest(dsl.Request{
			Method: "GET",
			Path:   dsl.String("/v3/notification_subscription/notif-sub-1"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-subscription-token"),
				"Content-Type":  dsl.String("application/json"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"id":           dsl.Like("notif-sub-1"),
				"userId":       dsl.Like("user-1"),
				"teamId":       dsl.Like("team-1"),
				"name":         dsl.String("Release webhook"),
				"isEnabled":    true,
				"webhookUrl":   dsl.String("https://example.com/hook"),
				"eventConfigs": eventConfigs,
			},
		})

	pact.AddInteraction().
		Given("Update notification subscription").
		UponReceiving("A request to update a notification subscription").
		WithRequest(dsl.Request{
			Method: "PUT",
			Path:   dsl.String("/v3/notification_subscription/notif-sub-1"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-subscription-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"isEnabled": false,
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"id":           dsl.Like("notif-sub-1"),
				"userId":       dsl.Like("user-1"),
				"teamId":       dsl.Like("team-1"),
				"name":         dsl.String("Release webhook"),
				"isEnabled":    false,
				"webhookUrl":   dsl.String("https://example.com/hook"),
				"eventConfigs": eventConfigs,
			},
		})

	pact.AddInteraction().
		Given("Delete notification subscription").
		UponReceiving("A request to delete a notification subscription").
		WithRequest(dsl.Request{
			Method: "DELETE",
			Path:   dsl.String("/v3/notification_subscription/notif-sub-1"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-subscription-token"),
				"Content-Type":  dsl.String("application/json"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 204,
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func Test_ListNotificationEventsAndTypes(t *testing.T) {
	test := func() error {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)
		api := platformclient.NewHTTPClient(u, "replicated-cli-events-token")
		client := realkotsclient.VendorV3Client{HTTPClient: *api}

		eventsResp, err := client.ListNotificationEvents(realkotsclient.ListNotificationEventsOpts{
			Status:         "failed",
			SubscriptionID: "notif-sub-1",
			PageSize:       20,
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, eventsResp.TotalCount)
		assert.Len(t, eventsResp.Events, 1)

		typesResp, err := client.ListNotificationEventTypes("release", 20)
		assert.NoError(t, err)
		assert.Equal(t, 1, typesResp.Total)
		assert.Equal(t, "release.promoted", typesResp.EventTypes[0].Key)

		return nil
	}

	pact.AddInteraction().
		Given("List notification events").
		UponReceiving("A request to list notification events").
		WithRequest(dsl.Request{
			Method: "GET",
			Path:   dsl.String("/v3/notification_events"),
			Query: dsl.MapMatcher{
				"status":          dsl.String("failed"),
				"subscription_id": dsl.String("notif-sub-1"),
				"pageSize":        dsl.String("20"),
			},
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-events-token"),
				"Content-Type":  dsl.String("application/json"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"events": []map[string]interface{}{
					{
						"id":             dsl.Like("evt-1"),
						"subscriptionId": dsl.Like("notif-sub-1"),
						"eventType":      dsl.String("release.promoted"),
						"eventData":      map[string]interface{}{"releaseId": "rel-1"},
						"payloadType":    dsl.String("webhook"),
						"payload":        map[string]interface{}{"event": "release.promoted"},
						"retryCount":     dsl.Like(8),
						"status":         dsl.String("failed"),
						"attempts":       []map[string]interface{}{},
					},
				},
				"totalCount": dsl.Like(1),
			},
		})

	pact.AddInteraction().
		Given("List notification event types").
		UponReceiving("A request to list notification event types").
		WithRequest(dsl.Request{
			Method: "GET",
			Path:   dsl.String("/v3/notification_event_types"),
			Query: dsl.MapMatcher{
				"q":     dsl.String("release"),
				"limit": dsl.String("20"),
			},
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-events-token"),
				"Content-Type":  dsl.String("application/json"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"eventTypes": []map[string]interface{}{
					{
						"key":         dsl.String("release.promoted"),
						"displayName": dsl.String("Release Promoted"),
						"description": dsl.String("A release was promoted."),
						"category":    dsl.String("Release"),
						"filters":     []map[string]interface{}{},
					},
				},
				"total": dsl.Like(1),
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func Test_NotificationRetryEmailAndWebhook(t *testing.T) {
	test := func() error {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)
		api := platformclient.NewHTTPClient(u, "replicated-cli-notification-actions-token")
		client := realkotsclient.VendorV3Client{HTTPClient: *api}

		retryResp, err := client.RetryNotificationEvent("evt-1")
		assert.NoError(t, err)
		assert.True(t, retryResp.Success)

		resendResp, err := client.ResendNotificationVerification(realkotsclient.ResendNotificationVerificationRequest{
			EmailAddress: "alerts@example.com",
		})
		assert.NoError(t, err)
		assert.True(t, resendResp.Success)

		verifyResp, err := client.VerifyNotificationEmail(realkotsclient.VerifyNotificationEmailRequest{
			EmailAddress:     "alerts@example.com",
			VerificationCode: "123456",
		})
		assert.NoError(t, err)
		assert.True(t, verifyResp.Success)

		webhookResp, err := client.TestNotificationWebhook(realkotsclient.TestNotificationWebhookRequest{
			WebhookURL: "https://example.com/hook",
			EventType:  "release.promoted",
		})
		assert.NoError(t, err)
		assert.True(t, webhookResp.Success)
		assert.NotNil(t, webhookResp.StatusCode)

		return nil
	}

	pact.AddInteraction().
		Given("Retry notification event").
		UponReceiving("A request to retry a notification event").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/v3/notification_event/evt-1/retry"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-notification-actions-token"),
				"Content-Type":  dsl.String("application/json"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"success": dsl.Like(true),
				"message": dsl.String("Notification event queued for retry"),
			},
		})

	pact.AddInteraction().
		Given("Resend notification verification email").
		UponReceiving("A request to resend notification verification email").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/v3/notification_email/resend_verification"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-notification-actions-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"emailAddress": "alerts@example.com",
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"success": dsl.Like(true),
				"message": dsl.String("Verification email sent successfully"),
			},
		})

	pact.AddInteraction().
		Given("Verify notification email").
		UponReceiving("A request to verify notification email").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/v3/notification_email/verify"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-notification-actions-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"emailAddress":     "alerts@example.com",
				"verificationCode": "123456",
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"success": dsl.Like(true),
				"message": dsl.String("Email address verified successfully"),
			},
		})

	pact.AddInteraction().
		Given("Test notification webhook").
		UponReceiving("A request to test notification webhook").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/v3/notification_webhook/test"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-notification-actions-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"webhookUrl": "https://example.com/hook",
				"eventType":  "release.promoted",
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"success":    dsl.Like(true),
				"statusCode": dsl.Like(200),
				"durationMs": dsl.Like(15),
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}
