package kotsclient

import (
	"fmt"
	"os"
	"os/exec"
)

// SCPToVM copies a file to a VM using SCP
func (c *VendorV3Client) SCPToVM(vmID string, sshUserFlag string, localPath string, remotePath string) error {
	connInfo, err := c.GetSSHConnectionInfo(vmID, sshUserFlag)
	if err != nil {
		return err
	}

	// Format the remote destination
	remoteDestination := formatRemotePath(connInfo, remotePath)

	// Build the scp command
	cmd := exec.Command("scp", "-P", fmt.Sprintf("%d", connInfo.Port), localPath, remoteDestination)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// SCPFromVM copies a file from a VM using SCP
func (c *VendorV3Client) SCPFromVM(vmID string, sshUserFlag string, remotePath string, localPath string) error {
	connInfo, err := c.GetSSHConnectionInfo(vmID, sshUserFlag)
	if err != nil {
		return err
	}

	// Format the remote source
	remoteSource := formatRemotePath(connInfo, remotePath)

	// Build the scp command
	cmd := exec.Command("scp", "-P", fmt.Sprintf("%d", connInfo.Port), remoteSource, localPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// formatRemotePath formats a remote path for SCP using the connection info
func formatRemotePath(connInfo *SSHConnectionInfo, path string) string {
	if connInfo.User != "" {
		return fmt.Sprintf("%s@%s:%s", connInfo.User, connInfo.Endpoint, path)
	}
	return fmt.Sprintf("%s:%s", connInfo.Endpoint, path)
}
