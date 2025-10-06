package integration

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReleaseCreate(t *testing.T) {
	tests := []struct {
		name       string
		cliArgs    []string
		wantFormat format
		wantLines  int
		wantOutput string
		setup      func(t *testing.T) *httptest.Server
	}{
		{
			name:       "release create table",
			cliArgs:    []string{"release", "create"},
			wantFormat: FormatTable,
			wantLines:  2, // Creating Release ✓ + SEQUENCE line
			setup: func(t *testing.T) *httptest.Server {
				r := mux.NewRouter()

				// List apps used by GetAppType("test-app") path; our CLI resolves app via API
				r.Methods(http.MethodGet).Path("/v1/apps").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`[{"app":{"id":"app-123","name":"test-app","slug":"test-app","scheduler":"native"}}]`))
				})

				// Create release
				r.Methods(http.MethodPost).Path("/v1/app/app-123/release").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte(`{"appId":"app-123","sequence":42}`))
				})

				// Update release YAML after creation
				r.Methods(http.MethodPut).Path("/v1/app/app-123/42/raw").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					body, err := io.ReadAll(r.Body)
					assert.NoError(t, err)
					assert.Contains(t, string(body), "apiVersion:")
					w.WriteHeader(http.StatusOK)
				})

				return httptest.NewServer(r)
			},
		},
		{
			name:       "release create json",
			cliArgs:    []string{"release", "create", "--output", "json"},
			wantFormat: FormatJSON,
			wantLines:  0,
			wantOutput: `{
  "sequence": 43,
  "appId": "app-123"
}
`,
			setup: func(t *testing.T) *httptest.Server {
				r := mux.NewRouter()

				r.Methods(http.MethodGet).Path("/v1/apps").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`[{"app":{"id":"app-123","name":"test-app","slug":"test-app","scheduler":"native"}}]`))
				})

				r.Methods(http.MethodPost).Path("/v1/app/app-123/release").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte(`{"appId":"app-123","sequence":43}`))
				})

				r.Methods(http.MethodPut).Path("/v1/app/app-123/43/raw").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				})

				return httptest.NewServer(r)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.setup(t)
			defer server.Close()

			// Create a temporary yaml-dir with a simple manifest and pass it
			tempDir := t.TempDir()
			manifest := "apiVersion: kots.io/v1beta1\nkind: Application\nmetadata:\n  name: test-app\nspec:\n  title: Test App\n"
			require.NoError(t, os.WriteFile(filepath.Join(tempDir, "app.yaml"), []byte(manifest), 0644))
			// Append flag at the end to avoid changing per-test cliArgs above
			tt.cliArgs = append(tt.cliArgs, "--yaml-file", filepath.Join(tempDir, "app.yaml"))

			// Set REPLICATED_APP to allow resolving the app
			cmd := getCommand(tt.cliArgs, server)
			cmd.Env = append(cmd.Env, "REPLICATED_APP=test-app")
			cmd.Env = append(cmd.Env, "HOME="+tempDir)

			out, err := cmd.CombinedOutput()
			assert.NoError(t, err)

			if tt.wantOutput != "" {
				require.Equal(t, tt.wantOutput, string(out))
				return
			}

			AssertCLIOutput(t, string(out), tt.wantFormat, tt.wantLines)
		})
	}
}
