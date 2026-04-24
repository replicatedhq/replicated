package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/stretchr/testify/require"
)

func TestCreateAndWaitForVM_ForbiddenErrors(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		contentType   string
		expectedError string
	}{
		{
			name:          "rbac denial returns server message",
			body:          `access to "kots/vm/create" is denied`,
			contentType:   "text/plain",
			expectedError: `access to "kots/vm/create" is denied`,
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
				if r.URL.Path != "/v3/vm" || r.Method != http.MethodPost {
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

			_, err := runner.createAndWaitForVM(kotsclient.CreateVMOpts{
				Name:         "test-vm",
				Distribution: "ubuntu",
				Version:      "22.04",
				Count:        1,
			})
			require.Error(t, err)
			require.Equal(t, tt.expectedError, err.Error())
		})
	}
}

func TestCreateVM_OverlayFSRequestBody(t *testing.T) {
	tests := []struct {
		name              string
		overlayFS         bool
		expectFieldPresent bool
		expectFieldValue   bool
	}{
		{name: "overlayfs true sends overlayfs:true", overlayFS: true, expectFieldPresent: true, expectFieldValue: true},
		{name: "overlayfs false omits the field", overlayFS: false, expectFieldPresent: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedBody map[string]any

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v3/vm" || r.Method != http.MethodPost {
					http.NotFound(w, r)
					return
				}

				bodyBytes, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				require.NoError(t, json.Unmarshal(bodyBytes, &capturedBody))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"vms":[{"id":"vm-1","status":"assigned"}]}`))
			}))
			defer server.Close()

			httpClient := platformclient.NewHTTPClient(server.URL, "fake-api-key")
			runner := &runners{
				kotsAPI: &kotsclient.VendorV3Client{HTTPClient: *httpClient},
			}

			_, err := runner.createAndWaitForVM(kotsclient.CreateVMOpts{
				Name:         "test-vm",
				Distribution: "ubuntu",
				Version:      "24.04",
				Count:        1,
				OverlayFS:    tt.overlayFS,
			})
			require.NoError(t, err)

			val, ok := capturedBody["overlayfs"]
			if tt.expectFieldPresent {
				require.True(t, ok, "expected overlayfs field in request body")
				require.Equal(t, tt.expectFieldValue, val)
			} else {
				require.False(t, ok, "expected overlayfs field to be omitted from request body")
			}
		})
	}
}
