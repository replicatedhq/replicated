package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"text/tabwriter"

	"github.com/replicatedhq/replicated/client"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestCustomerUpdateChannelOnlyUsesPatchWithoutOptionalFields(t *testing.T) {
	var patchBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v3/customer/customer-id":
			_, _ = w.Write([]byte(`{
				"customer": {
					"id": "customer-id",
					"name": "EC Customer",
					"email": "customer@example.com",
					"channels": [{"id": "unstable-channel-id", "name": "Unstable"}],
					"airgap": true,
					"isEmbeddedClusterDownloadEnabled": true,
					"isSupportBundleUploadEnabled": true
				}
			}`))
		case r.Method == http.MethodGet && r.URL.Path == "/v3/app/app-id/channel/stable":
			_, _ = w.Write([]byte(`{"channel":{"id":"stable-channel-id","name":"Stable"}}`))
		case r.Method == http.MethodPatch && r.URL.Path == "/v3/customer/customer-id":
			require.NoError(t, json.NewDecoder(r.Body).Decode(&patchBody))
			_, _ = w.Write([]byte(`{"customer":{"id":"customer-id","name":"EC Customer"}}`))
		default:
			http.Error(w, "unexpected request", http.StatusNotFound)
		}
	}))
	defer server.Close()

	api := client.NewClient(server.URL, "fake-api-key", "")
	r := &runners{
		appID:        "app-id",
		appType:      "kots",
		api:          api,
		outputFormat: "json",
		w:            tabwriter.NewWriter(io.Discard, 0, 0, 0, ' ', 0),
	}

	parent := r.InitCustomersCommand(&cobra.Command{Use: "replicated"})
	updateCmd := r.InitCustomerUpdateCommand(parent)
	require.NoError(t, updateCmd.Flags().Set("customer", "customer-id"))
	require.NoError(t, updateCmd.Flags().Set("channel", "stable"))
	require.NoError(t, updateCmd.RunE(updateCmd, nil))

	require.ElementsMatch(t, []string{"add_channels", "remove_channels"}, customerUpdateMapKeys(patchBody))
	require.Equal(t, []interface{}{"unstable-channel-id"}, patchBody["remove_channels"])
}

func TestCustomerUpdateNameOnlyDoesNotRequireChannel(t *testing.T) {
	var patchBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)
		require.Equal(t, "/v3/customer/customer-id", r.URL.Path)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&patchBody))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"customer":{"id":"customer-id","name":"New Name"}}`))
	}))
	defer server.Close()

	api := client.NewClient(server.URL, "fake-api-key", "")
	r := &runners{
		appID:        "app-id",
		appType:      "kots",
		api:          api,
		outputFormat: "json",
		w:            tabwriter.NewWriter(io.Discard, 0, 0, 0, ' ', 0),
	}

	parent := r.InitCustomersCommand(&cobra.Command{Use: "replicated"})
	updateCmd := r.InitCustomerUpdateCommand(parent)
	require.NoError(t, updateCmd.Flags().Set("customer", "customer-id"))
	require.NoError(t, updateCmd.Flags().Set("name", "New Name"))
	require.NoError(t, updateCmd.RunE(updateCmd, nil))

	require.Equal(t, map[string]interface{}{"name": "New Name"}, patchBody)
}

func TestCustomerUpdateRequiresAChangedField(t *testing.T) {
	r := &runners{
		appID:   "app-id",
		appType: "kots",
	}

	parent := r.InitCustomersCommand(&cobra.Command{Use: "replicated"})
	updateCmd := r.InitCustomerUpdateCommand(parent)
	require.NoError(t, updateCmd.Flags().Set("customer", "customer-id"))

	err := updateCmd.RunE(updateCmd, nil)
	require.EqualError(t, err, "at least one customer field must be specified")
}

func customerUpdateMapKeys(values map[string]interface{}) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}
