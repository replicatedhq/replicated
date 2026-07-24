package kotsclient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/stretchr/testify/require"
)

func TestUpdateCustomerUsesPatchAndOmitsUnchangedFields(t *testing.T) {
	var requestBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)
		require.Equal(t, "/v3/customer/customer-id", r.URL.Path)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&requestBody))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"customer":{"id":"customer-id"}}`))
	}))
	defer server.Close()

	httpClient := platformclient.NewHTTPClient(server.URL, "fake-api-key")
	client := &VendorV3Client{HTTPClient: *httpClient}

	customer, err := client.UpdateCustomer("customer-id", UpdateCustomerOpts{
		AddChannels: []CustomerChannel{{
			ID:        "stable-channel-id",
			IsDefault: true,
		}},
		RemoveChannels: []string{"unstable-channel-id"},
	})
	require.NoError(t, err)
	require.Equal(t, "customer-id", customer.ID)

	require.ElementsMatch(t, []string{"add_channels", "remove_channels"}, mapKeys(requestBody))
	require.Equal(t, []interface{}{"unstable-channel-id"}, requestBody["remove_channels"])

	addChannels, ok := requestBody["add_channels"].([]interface{})
	require.True(t, ok)
	require.Len(t, addChannels, 1)
	require.Equal(t, map[string]interface{}{
		"channel_id":              "stable-channel-id",
		"is_default_for_customer": true,
		"pinned_channel_sequence": nil,
	}, addChannels[0])
}

func TestUpdateCustomerIncludesExplicitFalse(t *testing.T) {
	var requestBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&requestBody))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"customer":{"id":"customer-id"}}`))
	}))
	defer server.Close()

	httpClient := platformclient.NewHTTPClient(server.URL, "fake-api-key")
	client := &VendorV3Client{HTTPClient: *httpClient}

	disabled := false
	_, err := client.UpdateCustomer("customer-id", UpdateCustomerOpts{
		IsEmbeddedClusterDownloadEnabled: &disabled,
	})
	require.NoError(t, err)

	require.Equal(t, map[string]interface{}{
		"is_embedded_cluster_download_enabled": false,
	}, requestBody)
}

func mapKeys(values map[string]interface{}) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}
