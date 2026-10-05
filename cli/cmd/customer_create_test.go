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

func TestCustomerCreateExpiresAt(t *testing.T) {
	var createBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v3/app/app-id/channel/stable":
			_, _ = w.Write([]byte(`{"channel":{"id":"stable-channel-id","name":"Stable"}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v3/customer":
			require.NoError(t, json.NewDecoder(r.Body).Decode(&createBody))
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"customer":{"id":"customer-id","name":"Acme"}}`))
		default:
			http.Error(w, "unexpected request", http.StatusNotFound)
		}
	}))
	defer server.Close()

	r := &runners{
		appID:        "app-id",
		appType:      "kots",
		api:          client.NewClient(server.URL, "fake-api-key", ""),
		outputFormat: "json",
		w:            tabwriter.NewWriter(io.Discard, 0, 0, 0, ' ', 0),
	}

	parent := r.InitCustomersCommand(&cobra.Command{Use: "replicated"})
	createCmd := r.InitCustomersCreateCommand(parent)
	require.NoError(t, createCmd.Flags().Set("name", "Acme"))
	require.NoError(t, createCmd.Flags().Set("channel", "stable"))
	require.NoError(t, createCmd.Flags().Set("expires-at", "2027-01-31"))
	require.NoError(t, createCmd.RunE(createCmd, nil))

	require.Equal(t, "2027-01-31T00:00:00Z", createBody["expires_at"])
}

func TestCustomerCreateExpiresAtInvalid(t *testing.T) {
	r := &runners{
		appID:   "app-id",
		appType: "kots",
	}

	parent := r.InitCustomersCommand(&cobra.Command{Use: "replicated"})
	createCmd := r.InitCustomersCreateCommand(parent)
	require.NoError(t, createCmd.Flags().Set("channel", "stable"))
	require.NoError(t, createCmd.Flags().Set("expires-at", "31/01/2027"))

	err := createCmd.RunE(createCmd, nil)
	require.ErrorContains(t, err, `invalid --expires-at value "31/01/2027"`)
}
