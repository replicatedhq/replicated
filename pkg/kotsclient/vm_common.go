package kotsclient

import (
	"fmt"
	"os"
)

// SSHConnectionInfo contains the information needed to connect to a VM via SSH or SCP
type SSHConnectionInfo struct {
	Endpoint string
	Port     int64
	User     string
}

// GetSSHConnectionInfo returns the SSH connection information for a VM
func (c *VendorV3Client) GetSSHConnectionInfo(vmID string, sshUserFlag string) (*SSHConnectionInfo, error) {
	vm, err := c.GetVM(vmID)
	if err != nil {
		return nil, err
	}

	if vm.DirectSSHEndpoint == "" || vm.DirectSSHPort == 0 {
		return nil, fmt.Errorf("VM %s does not have SSH access configured", vmID)
	}

	// Try to get the SSH user in order of precedence
	sshUser := getSSHUser(sshUserFlag)

	return &SSHConnectionInfo{
		Endpoint: vm.DirectSSHEndpoint,
		Port:     vm.DirectSSHPort,
		User:     sshUser,
	}, nil
}

// getSSHUser returns the SSH user based on the provided flag and environment variables
func getSSHUser(sshUserFlag string) string {
	// Try to get the SSH user in order of precedence
	if sshUserFlag != "" {
		return sshUserFlag
	}
	if user := os.Getenv("REPLICATED_SSH_USER"); user != "" {
		return user
	}
	if user := os.Getenv("GITHUB_ACTOR"); user != "" {
		return user
	}
	if user := os.Getenv("GITHUB_USER"); user != "" {
		return user
	}
	return ""
}
