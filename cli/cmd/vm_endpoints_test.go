package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetVMEndpoint(t *testing.T) {
	tests := []struct {
		name               string
		vmID               string
		endpointType       string
		mockVM             *VM
		mockGithubUsername string
		expectedOutput     string
		expectedError      string
	}{
		{
			name:         "Valid SSH endpoint",
			vmID:         "vm-123",
			endpointType: "ssh",
			mockVM: &VM{
				ID:             "vm-123",
				DirectEndpoint: "test-vm.example.com",
				DirectPort:     22,
				Status:         "running",
			},
			mockGithubUsername: "testuser",
			expectedOutput:     "ssh://testuser@test-vm.example.com:22\n",
			expectedError:      "",
		},
		{
			name:         "Valid SCP endpoint",
			vmID:         "vm-456",
			endpointType: "scp",
			mockVM: &VM{
				ID:             "vm-456",
				DirectEndpoint: "test-vm.example.com",
				DirectPort:     22,
				Status:         "running",
			},
			mockGithubUsername: "testuser",
			expectedOutput:     "scp://testuser@test-vm.example.com:22\n",
			expectedError:      "",
		},
		{
			name:         "Missing SSH endpoint",
			vmID:         "vm-789",
			endpointType: "ssh",
			mockVM: &VM{
				ID:             "vm-789",
				DirectEndpoint: "",
				DirectPort:     0,
				Status:         "running",
			},
			mockGithubUsername: "testuser",
			expectedOutput:     "",
			expectedError:      "VM vm-789 does not have ssh endpoint configured",
		},
		{
			name:         "Missing GitHub username",
			vmID:         "vm-123",
			endpointType: "ssh",
			mockVM: &VM{
				ID:             "vm-123",
				DirectEndpoint: "test-vm.example.com",
				DirectPort:     22,
				Status:         "running",
			},
			mockGithubUsername: "",
			expectedOutput:     "",
			expectedError:      "no github account associated with vendor portal user",
		},
		{
			name:         "Invalid endpoint type",
			vmID:         "vm-123",
			endpointType: "invalid",
			mockVM: &VM{
				ID:             "vm-123",
				DirectEndpoint: "test-vm.example.com",
				DirectPort:     22,
				Status:         "running",
			},
			mockGithubUsername: "testuser",
			expectedOutput:     "",
			expectedError:      "invalid endpoint type: invalid",
		},
		{
			name:         "VM not in running state",
			vmID:         "vm-123",
			endpointType: "ssh",
			mockVM: &VM{
				ID:             "vm-123",
				DirectEndpoint: "test-vm.example.com",
				DirectPort:     22,
				Status:         "provisioning",
			},
			mockGithubUsername: "testuser",
			expectedOutput:     "",
			expectedError:      "VM vm-123 is not in running state (current state: provisioning). SSH is only available for running VMs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Create test runner
			runner := &runners{}

			// Run function
			err := runner.getVMEndpoint(tt.vmID, tt.endpointType, tt.mockVM, tt.mockGithubUsername)

			// Restore stdout immediately to prevent nil pointer issues
			w.Close()
			os.Stdout = oldStdout
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// Assert expectations
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, output)
			}
		})
	}
}
