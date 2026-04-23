package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/stretchr/testify/require"
)

func TestListClusters_ForbiddenErrors(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		contentType   string
		expectedError string
	}{
		{
			name:          "rbac denial returns server message",
			body:          `access to "kots/cluster/list" is denied`,
			contentType:   "text/plain",
			// ListClusters wraps the error with "list clusters page %d", so the RBAC
			// message arrives wrapped at the cmd layer.
			expectedError: `list clusters page 0: access to "kots/cluster/list" is denied`,
		},
		{
			name:          "non-rbac forbidden returns compatibility matrix message",
			body:          `{"error":{"message":"You must read and accept the Compatibility Matrix Terms of Service"}}`,
			contentType:   "application/json",
			expectedError: ErrCompatibilityMatrixTermsNotAccepted.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !strings.HasPrefix(r.URL.Path, "/v3/clusters") || r.Method != http.MethodGet {
					http.NotFound(w, r)
					return
				}

				w.Header().Set("Content-Type", tt.contentType)
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			httpClient := platformclient.NewHTTPClient(server.URL, "fake-api-key")
			runner := &runners{
				kotsAPI: &kotsclient.VendorV3Client{HTTPClient: *httpClient},
			}

			err := runner.listClusters(nil, nil)
			require.Error(t, err)
			require.Equal(t, tt.expectedError, err.Error())
		})
	}
}
