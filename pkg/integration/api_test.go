package integration

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestAPI(t *testing.T) {
	tests := []struct {
		name       string
		cliArgs    []string
		getCommand func([]string, *httptest.Server) *exec.Cmd
		getServer  func(t *testing.T) *httptest.Server
		wantFormat format
		wantLines  int
		wantError  string
	}{
		{
			name:       "api get",
			cliArgs:    []string{"api", "get", "/v3/some/path"},
			getCommand: getCommand,
			getServer: func(t *testing.T) *httptest.Server {
				r := mux.NewRouter()

				r.Methods(http.MethodGet).Path("/v3/some/path").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					auth := r.Header.Get("Authorization")
					assert.Equal(t, auth, "test-token")

					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"test-key": "test-value"}`))
				})

				return httptest.NewServer(r)
			},
			wantFormat: FormatJSON,
		},
		{
			name:       "api patch",
			cliArgs:    []string{"api", "patch", "/v3/some/path", "-b", `{"test-key": "test-value"}`},
			getCommand: getCommand,
			getServer: func(t *testing.T) *httptest.Server {
				r := mux.NewRouter()
				r.Methods(http.MethodPatch).Path("/v3/some/path").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					auth := r.Header.Get("Authorization")
					assert.Equal(t, auth, "test-token")

					body, err := io.ReadAll(r.Body)
					assert.NoError(t, err)
					assert.Equal(t, string(body), `{"test-key": "test-value"}`)

					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"test-key": "test-value"}`))
				})
				return httptest.NewServer(r)
			},
			wantFormat: FormatJSON,
		},
		{
			name:       "api post",
			cliArgs:    []string{"api", "post", "/v3/some/path", "-b", `{"test-key": "test-value"}`},
			getCommand: getCommand,
			getServer: func(t *testing.T) *httptest.Server {
				r := mux.NewRouter()
				r.Methods(http.MethodPost).Path("/v3/some/path").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					auth := r.Header.Get("Authorization")
					assert.Equal(t, auth, "test-token")

					body, err := io.ReadAll(r.Body)
					assert.NoError(t, err)
					assert.Equal(t, string(body), `{"test-key": "test-value"}`)

					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"test-key": "test-value"}`))
				})
				return httptest.NewServer(r)
			},
			wantFormat: FormatJSON,
		},
		{
			name:       "api put",
			cliArgs:    []string{"api", "put", "/v3/some/path", "-b", `{"test-key": "test-value"}`},
			getCommand: getCommand,
			getServer: func(t *testing.T) *httptest.Server {
				r := mux.NewRouter()
				r.Methods(http.MethodPut).Path("/v3/some/path").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					auth := r.Header.Get("Authorization")
					assert.Equal(t, auth, "test-token")

					body, err := io.ReadAll(r.Body)
					assert.NoError(t, err)
					assert.Equal(t, string(body), `{"test-key": "test-value"}`)

					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"test-key": "test-value"}`))
				})
				return httptest.NewServer(r)
			},
			wantFormat: FormatJSON,
		},
		{
			name:       "api get missing api token",
			cliArgs:    []string{"api", "get", "/v3/some/path"},
			getCommand: getCommandWithoutToken,
			getServer: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusUnauthorized)
				}))
			},
			wantFormat: FormatTable,
			wantLines:  1,
			wantError:  "token or log in",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.getServer(t)
			defer server.Close()

			cmd := tt.getCommand(tt.cliArgs, server)
			out, err := cmd.CombinedOutput()

			if tt.wantError != "" {
				assert.Regexp(t, `^Error:`, string(out))
				assert.Contains(t, string(out), tt.wantError)
				return
			} else {
				assert.NoError(t, err)
			}

			AssertCLIOutput(t, string(out), tt.wantFormat, tt.wantLines)
		})
	}
}
