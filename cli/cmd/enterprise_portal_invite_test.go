package cmd

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"text/tabwriter"

	"github.com/replicatedhq/replicated/client"
	"github.com/stretchr/testify/require"
)

func TestEnterprisePortalInviteByNameFallsBackAfterIDNotFound(t *testing.T) {
	var invited bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v3/customer/Acme":
			http.Error(w, "not found", http.StatusNotFound)
		case r.Method == http.MethodPost && r.URL.Path == "/v3/customers/search":
			_, _ = w.Write([]byte(`{"customers":[{"id":"cus-1","name":"Acme"}],"total_hits":1}`))
		case r.Method == http.MethodGet && r.URL.Path == "/v3/customer/cus-1":
			_, _ = w.Write([]byte(`{"customer":{"id":"cus-1","name":"Acme"}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v3/app/app-id/enterprise-portal/customer-user":
			invited = true
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"url":"https://portal.example.com/invite"}`))
		default:
			http.Error(w, "unexpected request", http.StatusNotFound)
		}
	}))
	defer server.Close()

	api := client.NewClient(server.URL, "fake-api-key", "")
	r := &runners{
		appID:   "app-id",
		appType: "kots",
		api:     api,
		kotsAPI: api.KotsClient,
		w:       tabwriter.NewWriter(io.Discard, 0, 0, 0, ' ', 0),
	}

	require.NoError(t, r.enterprisePortalInvite(nil, "app-id", "Acme", []string{"user@example.com"}))
	require.True(t, invited)
}
