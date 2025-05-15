package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/stretchr/testify/assert"
)

// TestGetVMEndpoint tests the getVMEndpoint function with various scenarios
func TestGetVMEndpoint(t *testing.T) {
	// Define test cases with clear descriptions
	tests := []struct {
		name               string    // Test case name
		vmID               string    // Virtual machine ID to test
		endpointType       string    // Endpoint type (ssh, scp, or invalid)
		username           string    // Custom username (empty to use GitHub username)
		mockVM             *types.VM // Mock VM response from API
		mockGithubUsername string    // Mock GitHub username response
		githubAPIError     bool      // Should GitHub username API return an error
		expectedOutput     string    // Expected output to stdout
		expectedError      string    // Expected error string (empty for no error)
	}{
		{
			name:         "Success - Get SSH endpoint for running VM",
			vmID:         "vm-123",
			endpointType: "ssh",
			username:     "",
			mockVM: &types.VM{
				ID:                "vm-123",
				DirectSSHEndpoint: "test-vm.example.com",
				DirectSSHPort:     22,
				Status:            types.VMStatusRunning,
			},
			mockGithubUsername: "testuser",
			githubAPIError:     false,
			expectedOutput:     "ssh://testuser@test-vm.example.com:22\n",
			expectedError:      "",
		},
		{
			name:         "Success - Get SCP endpoint for running VM",
			vmID:         "vm-456",
			endpointType: "scp",
			username:     "",
			mockVM: &types.VM{
				ID:                "vm-456",
				DirectSSHEndpoint: "test-vm.example.com",
				DirectSSHPort:     22,
				Status:            types.VMStatusRunning,
			},
			mockGithubUsername: "testuser",
			githubAPIError:     false,
			expectedOutput:     "scp://testuser@test-vm.example.com:22\n",
			expectedError:      "",
		},
		{
			name:         "Success - Custom username provided",
			vmID:         "vm-123",
			endpointType: "ssh",
			username:     "customuser",
			mockVM: &types.VM{
				ID:                "vm-123",
				DirectSSHEndpoint: "test-vm.example.com",
				DirectSSHPort:     22,
				Status:            types.VMStatusRunning,
			},
			mockGithubUsername: "testuser", // Should be ignored
			githubAPIError:     false,      // Should be ignored
			expectedOutput:     "ssh://customuser@test-vm.example.com:22\n",
			expectedError:      "",
		},
		{
			name:         "Error - Missing SSH endpoint configuration",
			vmID:         "vm-789",
			endpointType: "ssh",
			username:     "",
			mockVM: &types.VM{
				ID:                "vm-789",
				DirectSSHEndpoint: "",
				DirectSSHPort:     0,
				Status:            types.VMStatusRunning,
			},
			mockGithubUsername: "testuser",
			githubAPIError:     false,
			expectedOutput:     "",
			expectedError:      "VM vm-789 does not have ssh endpoint configured",
		},
		{
			name:         "Error - Missing GitHub username",
			vmID:         "vm-123",
			endpointType: "ssh",
			username:     "",
			mockVM: &types.VM{
				ID:                "vm-123",
				DirectSSHEndpoint: "test-vm.example.com",
				DirectSSHPort:     22,
				Status:            types.VMStatusRunning,
			},
			mockGithubUsername: "",
			githubAPIError:     false,
			expectedOutput:     "",
			expectedError:      "no GitHub account associated with Vendor Portal user",
		},
		{
			name:         "Error - GitHub username API error",
			vmID:         "vm-123",
			endpointType: "ssh",
			username:     "",
			mockVM: &types.VM{
				ID:                "vm-123",
				DirectSSHEndpoint: "test-vm.example.com",
				DirectSSHPort:     22,
				Status:            types.VMStatusRunning,
			},
			mockGithubUsername: "", // Won't be used due to API error
			githubAPIError:     true,
			expectedOutput:     "",
			expectedError:      "--username flag to specify a custom username", // Check for new error message
		},
		{
			name:         "Error - Invalid endpoint type",
			vmID:         "vm-123",
			endpointType: "invalid",
			username:     "",
			mockVM: &types.VM{
				ID:                "vm-123",
				DirectSSHEndpoint: "test-vm.example.com",
				DirectSSHPort:     22,
				Status:            types.VMStatusRunning,
			},
			mockGithubUsername: "testuser",
			githubAPIError:     false,
			expectedOutput:     "",
			expectedError:      "invalid endpoint type: invalid",
		},
		{
			name:         "Error - VM not in running state",
			vmID:         "vm-123",
			endpointType: "ssh",
			username:     "",
			mockVM: &types.VM{
				ID:                "vm-123",
				DirectSSHEndpoint: "test-vm.example.com",
				DirectSSHPort:     22,
				Status:            types.VMStatusProvisioning,
			},
			mockGithubUsername: "testuser",
			githubAPIError:     false,
			expectedOutput:     "",
			expectedError:      "VM vm-123 is not in running state (current state: provisioning). SSH is only available for running VMs",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock server and client for API testing
			server, apiClient := createMockServerAndClient(tc.vmID, tc.mockVM, tc.mockGithubUsername, tc.githubAPIError)
			defer server.Close()

			// Capture stdout directly
			originalStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Create test runner with mock client
			runner := &runners{
				kotsAPI: apiClient,
			}

			// Execute the function under test
			err := runner.getVMEndpoint(tc.vmID, tc.endpointType, tc.username)

			// Get captured output
			w.Close()
			os.Stdout = originalStdout
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// Assert expectations for both error cases and success cases
			if tc.expectedError != "" {
				assert.Error(t, err, "Expected an error but got none")
				assert.Contains(t, err.Error(), tc.expectedError, "Error message did not match expected")
			} else {
				assert.NoError(t, err, "Unexpected error occurred")
				assert.Equal(t, tc.expectedOutput, output, "Output did not match expected")
			}
		})
	}
}

// createMockServerAndClient sets up a mock HTTP server and returns both kotsAPI.GetVM and kotsAPI.GetGitHubUsername
func createMockServerAndClient(vmID string, mockVM *types.VM, mockGithubUsername string, githubAPIError bool) (*httptest.Server, *kotsclient.VendorV3Client) {
	// Create a mock HTTP server to handle API requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		// Handle GetVM request
		case r.URL.Path == fmt.Sprintf("/v3/vm/%s", vmID) && r.Method == "GET":
			response := kotsclient.GetVMResponse{
				VM: mockVM,
			}
			json.NewEncoder(w).Encode(response)

		// Handle GetGitHubUsername request
		case r.URL.Path == "/v1/user" && r.Method == "GET":
			if githubAPIError {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "GitHub API error"})
				return
			}
			response := kotsclient.GetUserResponse{
				GitHubUsername: mockGithubUsername,
			}
			json.NewEncoder(w).Encode(response)

		// Handle unexpected requests
		default:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		}
	}))

	// Create an API client that uses the mock server
	httpClient := platformclient.NewHTTPClient(server.URL, "fake-api-key")
	apiClient := &kotsclient.VendorV3Client{
		HTTPClient: *httpClient,
	}

	return server, apiClient
}
