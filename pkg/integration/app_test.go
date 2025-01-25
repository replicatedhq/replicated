package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestAppLs(t *testing.T) {
	tests := []struct {
		name            string
		cliArgs         []string
		getServer       func() *httptest.Server
		wantFormat      format
		wantLines       int
		wantAPIRequests []string
		ignoreCLIOutput bool
	}{
		{
			name:    "app ls table",
			cliArgs: []string{"app", "ls"},
			getServer: func() *httptest.Server {
				r := mux.NewRouter()

				r.Methods(http.MethodGet).Path("/v3/apps").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"apps": [{"id": "123", "name": "test-app"}]}`))
				})

				return httptest.NewServer(r)
			},
			wantFormat: FormatTable,
			wantLines:  2,
		},
		{
			name:    "app ls json",
			cliArgs: []string{"app", "ls", "--output", "json"},
			getServer: func() *httptest.Server {
				r := mux.NewRouter()

				r.Methods(http.MethodGet).Path("/v3/apps").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"apps": [{"id": "123", "name": "test-app"}]}`))
				})

				return httptest.NewServer(r)
			},
			wantFormat: FormatJSON,
			wantLines:  0,
		},
		{
			name:    "app rm",
			cliArgs: []string{"app", "rm", "app-slug", "--force"},
			getServer: func() *httptest.Server {
				r := mux.NewRouter()

				r.Methods(http.MethodGet).Path("/v3/apps").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"apps": [{"id": "123", "name": "test-app", "slug": "app-slug"}]}`))
				})

				r.Methods(http.MethodDelete).Path("/v3/app/123").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				})

				return httptest.NewServer(r)
			},
			wantFormat: FormatTable,
			wantLines:  4, // fetching and deleting progress plus 2 lines for the table
		},
		{
			name:    "app create",
			cliArgs: []string{"app", "create", "App Name"},
			getServer: func() *httptest.Server {
				r := mux.NewRouter()

				r.Methods(http.MethodPost).Path("/v3/app").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Validate the request body
					body, err := io.ReadAll(r.Body)
					assert.NoError(t, err)
					requestPayload := make(map[string]string)
					err = json.Unmarshal(body, &requestPayload)
					assert.NoError(t, err)
					assert.Equal(t, "App Name", requestPayload["name"])

					// Return a response
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte(`{"app": {"id": "123", "name": "App Name", "slug": "app-slug"}}`))
				})

				return httptest.NewServer(r)
			},
			wantFormat: FormatTable,
			wantLines:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.getServer()
			defer server.Close()

			cmd := getCommand(tt.cliArgs, server)
			out, err := cmd.CombinedOutput()
			assert.NoError(t, err)

			AssertCLIOutput(t, string(out), tt.wantFormat, tt.wantLines)
		})
	}
}
