package cmd

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
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

func TestCreateVMNetworkAndNetworkPolicyAreMutuallyExclusive(t *testing.T) {
	runner := &runners{}
	runner.args.createVMName = "test-vm"
	runner.args.createVMDistribution = "ubuntu"
	runner.args.createVMNetwork = "network-id"
	runner.args.createVMNetworkPolicy = "airgap"

	err := runner.createVM(&cobra.Command{}, nil)
	require.ErrorContains(t, err, "cannot specify both --network and --network-policy")
}
