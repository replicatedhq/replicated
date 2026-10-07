package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomerArchiveUnarchive(t *testing.T) {
	acme := map[string]interface{}{"id": "cus-1", "name": "Acme", "isArchived": false}
	acmeArchived := map[string]interface{}{"id": "cus-old", "name": "Acme", "isArchived": true}
	acmeArchivedOlder := map[string]interface{}{"id": "cus-older", "name": "Acme", "isArchived": true}

	tests := []struct {
		name                string
		cliArgs             []string
		customers           []map[string]interface{} // returned by ID and by search
		wantErr             string                   // when set, the command must fail with this message
		wantIncludeArchived *bool                    // asserted on the search request
		wantWrites          []string                 // archive, unarchive and invite requests
		wantContain         []string
	}{
		{
			name:                "archive by name falls back to search after ID lookup 404s",
			cliArgs:             []string{"customer", "archive", "Acme"},
			customers:           []map[string]interface{}{acme},
			wantIncludeArchived: boolPtr(false),
			wantWrites:          []string{"POST /v3/customer/cus-1/archive"},
		},
		{
			name:       "unarchive by ID",
			cliArgs:    []string{"customer", "unarchive", "cus-old"},
			customers:  []map[string]interface{}{acmeArchived},
			wantWrites: []string{"POST /v3/customer/cus-old/unarchive"},
		},
		{
			name:      "unarchive by ID of an active customer errors",
			cliArgs:   []string{"customer", "unarchive", "cus-1"},
			customers: []map[string]interface{}{acme},
			wantErr:   `customer "Acme" is not archived`,
		},
		{
			name:                "unarchive by name skips active customers with the same name",
			cliArgs:             []string{"customer", "unarchive", "Acme"},
			customers:           []map[string]interface{}{acme, acmeArchived},
			wantIncludeArchived: boolPtr(true),
			wantWrites:          []string{"POST /v3/customer/cus-old/unarchive"},
		},
		{
			name:      "unarchive by name with several archived matches is ambiguous",
			cliArgs:   []string{"customer", "unarchive", "Acme"},
			customers: []map[string]interface{}{acmeArchived, acmeArchivedOlder},
			wantErr:   `customer "Acme" is ambiguous, please use customer ID`,
		},
		{
			name:      "unarchive by name with only an active match errors",
			cliArgs:   []string{"customer", "unarchive", "Acme"},
			customers: []map[string]interface{}{acme},
			wantErr:   `no archived customer named "Acme" found; it may not be archived`,
		},
		{
			name:    "unarchive by name with no search results errors",
			cliArgs: []string{"customer", "unarchive", "Acme"},
			wantErr: `no archived customer named "Acme" found; it may not be archived`,
		},
		{
			name:                "enterprise portal invite by customer name",
			cliArgs:             []string{"enterprise-portal", "invite", "--customer", "Acme", "user@example.com"},
			customers:           []map[string]interface{}{acme},
			wantIncludeArchived: boolPtr(false),
			wantWrites:          []string{"POST /v3/app/app-123/enterprise-portal/customer-user"},
			wantContain:         []string{"user@example.com: https://portal.example.com/invite"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &fakeCustomerAPI{customers: tt.customers}
			server := api.server(t)
			defer server.Close()

			cmd := getCommand(tt.cliArgs, server)
			// Per-test HOME so the app cache from one case doesn't bleed into the next.
			cmd.Env = append(cmd.Env, "REPLICATED_APP=test-app", "HOME="+t.TempDir())

			out, err := cmd.CombinedOutput()
			if tt.wantErr != "" {
				assert.Error(t, err, "expected non-zero exit; output:\n%s", out)
				assert.Contains(t, string(out), tt.wantErr)
			} else {
				assert.NoError(t, err, "cli failed: %s", out)
			}

			for _, want := range tt.wantContain {
				assert.Contains(t, string(out), want)
			}

			if tt.wantIncludeArchived != nil {
				require.NotNil(t, api.searchBody, "search was not called")
				assert.Equal(t, *tt.wantIncludeArchived, api.searchBody["include_archived"])
			}

			assert.Equal(t, tt.wantWrites, api.writes)
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}

// fakeCustomerAPI serves the vendor API endpoints used to look up, archive,
// unarchive and invite customers of app test-app.
type fakeCustomerAPI struct {
	customers []map[string]interface{}

	mu         sync.Mutex
	searchBody map[string]interface{}
	writes     []string
}

func (f *fakeCustomerAPI) server(t *testing.T) *httptest.Server {
	r := mux.NewRouter()

	// /v1/apps returns no platform apps so the app resolves as kots.
	r.Methods(http.MethodGet).Path("/v1/apps").HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`[]`))
	})
	r.Methods(http.MethodGet).Path("/v3/apps").HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`{"apps":[{"id":"app-123","name":"Test App","slug":"test-app"}]}`))
	})

	// A name passed as a customer ID 404s, which makes the CLI fall back to search.
	r.Methods(http.MethodGet).Path("/v3/customer/{id}").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		for _, c := range f.customers {
			if c["id"] == mux.Vars(req)["id"] {
				json.NewEncoder(w).Encode(map[string]interface{}{"customer": c})
				return
			}
		}
		http.NotFound(w, req)
	})

	r.Methods(http.MethodPost).Path("/v3/customers/search").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var body map[string]interface{}
		assert.NoError(t, json.NewDecoder(req.Body).Decode(&body))
		f.mu.Lock()
		f.searchBody = body
		f.mu.Unlock()

		customers := f.customers
		if customers == nil {
			customers = []map[string]interface{}{}
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"customers": customers, "total_hits": len(customers)})
	})

	record := func(_ http.ResponseWriter, req *http.Request) {
		f.mu.Lock()
		f.writes = append(f.writes, req.Method+" "+req.URL.Path)
		f.mu.Unlock()
	}
	r.Methods(http.MethodPost).Path("/v3/customer/{id}/archive").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		record(w, req)
		w.WriteHeader(http.StatusNoContent)
	})
	r.Methods(http.MethodPost).Path("/v3/customer/{id}/unarchive").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		record(w, req)
		w.WriteHeader(http.StatusNoContent)
	})
	r.Methods(http.MethodPost).Path("/v3/app/{appID}/enterprise-portal/customer-user").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		record(w, req)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"url":"https://portal.example.com/invite"}`))
	})

	// Catch-all to surface unexpected calls in test output.
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		t.Logf("unexpected %s %s?%s", req.Method, req.URL.Path, req.URL.RawQuery)
		http.NotFound(w, req)
	})

	return httptest.NewServer(r)
}
