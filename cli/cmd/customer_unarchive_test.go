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

func TestCustomerUnarchiveByNameUnarchivesAllArchivedMatches(t *testing.T) {
	var searchBody map[string]interface{}
	var unarchived []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v3/customer/Acme":
			http.Error(w, "not found", http.StatusNotFound)
		case r.Method == http.MethodPost && r.URL.Path == "/v3/customers/search":
			assert.NoError(t, json.NewDecoder(r.Body).Decode(&searchBody))
			_, _ = w.Write([]byte(`{"customers":[
				{"id":"cus-active","name":"Acme","isArchived":false},
				{"id":"cus-old-1","name":"Acme","isArchived":true},
				{"id":"cus-old-2","name":"Acme","isArchived":true},
				{"id":"cus-other","name":"Acme Corp","isArchived":true}
			],"total_hits":4}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v3/customer/cus-old-1/unarchive",
			r.Method == http.MethodPost && r.URL.Path == "/v3/customer/cus-old-2/unarchive":
			unarchived = append(unarchived, r.URL.Path)
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
	require.NoError(t, cmd.RunE(cmd, []string{"Acme"}))
	require.Equal(t, true, searchBody["include_archived"])
	require.Equal(t, []string{"/v3/customer/cus-old-1/unarchive", "/v3/customer/cus-old-2/unarchive"}, unarchived)
}

func TestCustomerUnarchiveByNameNoArchivedMatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v3/customer/Acme":
			http.Error(w, "not found", http.StatusNotFound)
		case r.Method == http.MethodPost && r.URL.Path == "/v3/customers/search":
			_, _ = w.Write([]byte(`{"customers":[{"id":"cus-active","name":"Acme","isArchived":false}],"total_hits":1}`))
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
	require.ErrorContains(t, cmd.RunE(cmd, []string{"Acme"}), `find customer "Acme"`)
}

func TestCustomerUnarchiveMissingArgs(t *testing.T) {
	r := &runners{appID: "app-id", appType: "kots"}
	parent := r.InitCustomersCommand(&cobra.Command{Use: "replicated"})
	cmd := r.InitCustomersUnarchiveCommand(parent)
	require.EqualError(t, cmd.Args(cmd, nil), "requires at least 1 arg(s), only received 0")
}
