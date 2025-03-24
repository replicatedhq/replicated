package kotsclient

import (
	"fmt"
	"os"
	"os/exec"
)

// SSHIntoVM connects to a VM via SSH using the provided ID and optional user flag
func (c *VendorV3Client) SSHIntoVM(vmID string, sshUserFlag string, identityFile string) error {
	connInfo, err := c.GetSSHConnectionInfo(vmID, sshUserFlag)
	if err != nil {
		return err
	}

	sshEndpoint := connInfo.Endpoint
	if connInfo.User != "" {
		sshEndpoint = fmt.Sprintf("%s@%s", connInfo.User, connInfo.Endpoint)
	}

	var cmd *exec.Cmd
	if identityFile != "" {
		cmd = exec.Command("ssh", sshEndpoint, "-p", fmt.Sprintf("%d", connInfo.Port), "-i", identityFile)
	} else {
		cmd = exec.Command("ssh", sshEndpoint, "-p", fmt.Sprintf("%d", connInfo.Port))
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
