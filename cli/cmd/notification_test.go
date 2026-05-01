package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadNotificationCreateRequest(t *testing.T) {
	t.Run("valid webhook request", func(t *testing.T) {
		path := writeNotificationJSON(t, `{
  "eventConfigs": [{"eventType":"customer.created","filters":{}}],
  "isEnabled": true,
  "webhookUrl": "https://example.com/webhook"
}`)

		req, err := readNotificationCreateRequest(path)
		require.NoError(t, err)
		require.Equal(t, "https://example.com/webhook", req.WebhookURL)
		require.Len(t, req.EventConfigs, 1)
	})

	t.Run("missing delivery target", func(t *testing.T) {
		path := writeNotificationJSON(t, `{
  "eventConfigs": [{"eventType":"customer.created","filters":{}}],
  "isEnabled": true
}`)

		_, err := readNotificationCreateRequest(path)
		require.EqualError(t, err, "one of emailAddress or webhookUrl must be set")
	})

	t.Run("missing event configs", func(t *testing.T) {
		path := writeNotificationJSON(t, `{
  "isEnabled": true,
  "emailAddress": "alerts@example.com"
}`)

		_, err := readNotificationCreateRequest(path)
		require.EqualError(t, err, "eventConfigs must contain at least one event config")
	})
}

func TestReadNotificationUpdateRequest(t *testing.T) {
	t.Run("reject empty patch", func(t *testing.T) {
		path := writeNotificationJSON(t, `{}`)

		_, err := readNotificationUpdateRequest(path)
		require.EqualError(t, err, "subscription update file must include at least one field")
	})

	t.Run("accept explicit false enable flag", func(t *testing.T) {
		path := writeNotificationJSON(t, `{"isEnabled":false}`)

		req, err := readNotificationUpdateRequest(path)
		require.NoError(t, err)
		require.NotNil(t, req.IsEnabled)
		require.False(t, *req.IsEnabled)
	})
}

func TestReadNotificationWebhookTestRequest(t *testing.T) {
	t.Run("reject missing event type", func(t *testing.T) {
		path := writeNotificationJSON(t, `{"webhookUrl":"https://example.com/webhook"}`)

		_, err := readNotificationWebhookTestRequest(path)
		require.EqualError(t, err, "eventType is required")
	})

	t.Run("reject invalid json", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "bad.json")
		require.NoError(t, os.WriteFile(path, []byte(`{"webhookUrl":`), 0644))

		_, err := readNotificationWebhookTestRequest(path)
		require.EqualError(t, err, "file is not valid JSON")
	})
}

func writeNotificationJSON(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "notification.json")
	require.NoError(t, os.WriteFile(path, []byte(body), 0644))
	return path
}
