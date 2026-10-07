package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"text/tabwriter"

	"github.com/replicatedhq/replicated/client"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomerUnarchiveByID(t *testing.T) {
	var unarchived []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v3/customer/cus-1":
			_, _ = w.Write([]byte(`{"customer":{"id":"cus-1","name":"Acme"}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v3/customer/cus-1/unarchive":
			unarchived = append(unarchived, "cus-1")
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
	cmd := r.InitCustomersUnarchiveCommand(parent)
	require.NoError(t, cmd.RunE(cmd, []string{"cus-1"}))
	require.Equal(t, []string{"cus-1"}, unarchived)
}

// unarchiveByNameServer fakes the vendor API for an unarchive by name of "Acme".
// The search returns searchResults, and every unarchive call is recorded in unarchived.
func unarchiveByNameServer(t *testing.T, searchResults string, searchBody *map[string]interface{}, unarchived *[]string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v3/customer/Acme":
			http.Error(w, "not found", http.StatusNotFound)
		case r.Method == http.MethodPost && r.URL.Path == "/v3/customers/search":
			assert.NoError(t, json.NewDecoder(r.Body).Decode(searchBody))
			_, _ = w.Write([]byte(searchResults))
		case r.Method == http.MethodGet && r.URL.Path == "/v3/customer/cus-old":
			_, _ = w.Write([]byte(`{"customer":{"id":"cus-old","name":"Acme","isArchived":true}}`))
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/unarchive"):
			*unarchived = append(*unarchived, r.URL.Path)
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "unexpected request", http.StatusNotFound)
		}
	}))
}

func runUnarchive(t *testing.T, serverURL string, nameOrID string) error {
	r := &runners{
		appID:   "app-id",
		appType: "kots",
		api:     client.NewClient(serverURL, "fake-api-key", ""),
		w:       tabwriter.NewWriter(io.Discard, 0, 0, 0, ' ', 0),
	}

	parent := r.InitCustomersCommand(&cobra.Command{Use: "replicated"})
	cmd := r.InitCustomersUnarchiveCommand(parent)
	return cmd.RunE(cmd, []string{nameOrID})
}

func TestCustomerUnarchiveByNameSkipsActiveCustomers(t *testing.T) {
	var searchBody map[string]interface{}
	var unarchived []string
	server := unarchiveByNameServer(t, `{"customers":[
		{"id":"cus-active","name":"Acme","isArchived":false},
		{"id":"cus-old","name":"Acme","isArchived":true}
	],"total_hits":2}`, &searchBody, &unarchived)
	defer server.Close()

	require.NoError(t, runUnarchive(t, server.URL, "Acme"))
	require.Equal(t, true, searchBody["include_archived"])
	require.Equal(t, []string{"/v3/customer/cus-old/unarchive"}, unarchived)
}

func TestCustomerUnarchiveByNameAmbiguous(t *testing.T) {
	var searchBody map[string]interface{}
	var unarchived []string
	server := unarchiveByNameServer(t, `{"customers":[
		{"id":"cus-old","name":"Acme","isArchived":true},
		{"id":"cus-older","name":"Acme","isArchived":true}
	],"total_hits":2}`, &searchBody, &unarchived)
	defer server.Close()

	require.ErrorContains(t, runUnarchive(t, server.URL, "Acme"), `customer "Acme" is ambiguous, please use customer ID`)
	require.Empty(t, unarchived)
}

func TestCustomerUnarchiveByNameNoArchivedMatch(t *testing.T) {
	var searchBody map[string]interface{}
	var unarchived []string
	server := unarchiveByNameServer(t, `{"customers":[
		{"id":"cus-active","name":"Acme","isArchived":false}
	],"total_hits":1}`, &searchBody, &unarchived)
	defer server.Close()

	require.ErrorContains(t, runUnarchive(t, server.URL, "Acme"), `customer "Acme" not found`)
	require.Empty(t, unarchived)
}

func TestCustomerUnarchiveMissingArgs(t *testing.T) {
	r := &runners{appID: "app-id", appType: "kots"}
	parent := r.InitCustomersCommand(&cobra.Command{Use: "replicated"})
	cmd := r.InitCustomersUnarchiveCommand(parent)
	require.EqualError(t, cmd.Args(cmd, nil), "requires at least 1 arg(s), only received 0")
}
