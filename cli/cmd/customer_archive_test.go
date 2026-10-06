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

func TestCustomerArchiveByNameFallsBackAfterIDNotFound(t *testing.T) {
	var searchBody map[string]interface{}
	var archived []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v3/customer/Acme":
			http.Error(w, "not found", http.StatusNotFound)
		case r.Method == http.MethodPost && r.URL.Path == "/v3/customers/search":
			require.NoError(t, json.NewDecoder(r.Body).Decode(&searchBody))
			_, _ = w.Write([]byte(`{"customers":[{"id":"cus-1","name":"Acme"}],"total_hits":1}`))
		case r.Method == http.MethodGet && r.URL.Path == "/v3/customer/cus-1":
			_, _ = w.Write([]byte(`{"customer":{"id":"cus-1","name":"Acme"}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v3/customer/cus-1/archive":
			archived = append(archived, "cus-1")
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "unexpected request", http.StatusNotFound)
		}
	}))
	defer server.Close()

	r := &runners{
		appID:   "app-id",
		appType: "kots",
		api:     client.NewClient(server.URL, "fake-api-key", ""),
		w:       tabwriter.NewWriter(io.Discard, 0, 0, 0, ' ', 0),
	}

	parent := r.InitCustomersCommand(&cobra.Command{Use: "replicated"})
	cmd := r.InitCustomersArchiveCommand(parent)
	require.NoError(t, cmd.RunE(cmd, []string{"Acme"}))
	require.Equal(t, false, searchBody["include_archived"])
	require.Equal(t, []string{"cus-1"}, archived)
}

func TestCustomerArchiveArgs(t *testing.T) {
	r := &runners{appID: "app-id", appType: "kots"}
	parent := r.InitCustomersCommand(&cobra.Command{Use: "replicated"})
	cmd := r.InitCustomersArchiveCommand(parent)

	require.EqualError(t, cmd.Args(cmd, nil), "requires at least 1 arg(s), only received 0")
	require.NoError(t, cmd.Args(cmd, []string{"Acme"}))

	require.NoError(t, cmd.Flags().Set("customer", "Acme"))
	require.NoError(t, cmd.Args(cmd, nil))
}
