package kotsclient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/stretchr/testify/require"
)

func TestEnterprisePortalClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "test-token", r.Header.Get("Authorization"))

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v3/app/test-app/enterprise-portal/status":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "active",
			})
		case r.Method == http.MethodPut && r.URL.Path == "/v3/app/test-app/enterprise-portal/status":
			var body UpdateEnterprisePortalStatusRequest
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			require.Equal(t, "inactive", body.Status)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "inactive",
			})
		case r.Method == http.MethodPost && r.URL.Path == "/v3/app/test-app/enterprise-portal/customer-user":
			var body InviteEnterprisePortalRequest
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			require.Equal(t, "cus-123", body.CustomerID)
			require.Equal(t, "user@example.com", body.EmailAddress)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"url": "https://enterprise.example.com/invite/abc123",
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v3/app/test-app/enterprise-portal/customer-users":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"users": []map[string]interface{}{
					{"email": "user1@example.com"},
					{"email": "user2@example.com"},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	api := platformclient.NewHTTPClient(server.URL, "test-token")
	client := VendorV3Client{HTTPClient: *api}

	// Test GetEnterprisePortalStatus
	status, err := client.GetEnterprisePortalStatus("test-app")
	require.NoError(t, err)
	require.Equal(t, "active", status)

	// Test UpdateEnterprisePortalStatus
	updatedStatus, err := client.UpdateEnterprisePortalStatus("test-app", "inactive")
	require.NoError(t, err)
	require.Equal(t, "inactive", updatedStatus)

	// Test SendEnterprisePortalInvite
	inviteURL, err := client.SendEnterprisePortalInvite("test-app", "cus-123", "user@example.com")
	require.NoError(t, err)
	require.Equal(t, "https://enterprise.example.com/invite/abc123", inviteURL)

	// Test ListEnterprisePortalUsers
	users, err := client.ListEnterprisePortalUsers("test-app", false)
	require.NoError(t, err)
	require.Len(t, users, 2)
	require.Equal(t, "user1@example.com", users[0].Email)
	require.Equal(t, "user2@example.com", users[1].Email)
}
