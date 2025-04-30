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
				ID:                "vm-123",
				DirectSSHEndpoint: "test-vm.example.com",
				DirectSSHPort:     22,
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
				ID:                "vm-456",
				DirectSSHEndpoint: "test-vm.example.com",
				DirectSSHPort:     22,
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
				ID:                "vm-789",
				DirectSSHEndpoint: "",
				DirectSSHPort:     0,
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
				ID:                "vm-123",
				DirectSSHEndpoint: "test-vm.example.com",
				DirectSSHPort:     22,
			},
			mockGithubUsername: "",
			expectedOutput:     "",
			expectedError:      "no github account associated with vendor portal user",
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
