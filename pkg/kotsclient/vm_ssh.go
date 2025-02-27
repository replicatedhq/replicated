package kotsclient

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/replicatedhq/replicated/pkg/types"
)

// SSHIntoVM connects to a VM via SSH using the provided ID and optional user flag
func (c *VendorV3Client) SSHIntoVM(id string, sshUserFlag string) error {
	vm, err := c.GetVM(id)
	if err != nil {
		return err
	}

	if vm.DirectSSHEndpoint == "" || vm.DirectSSHPort == 0 {
		return fmt.Errorf("VM %s does not have SSH access configured", id)
	}

	// Check if VM is running
	if vm.Status != types.VMStatusRunning {
		return fmt.Errorf("VM: %s is not running (current status: %q): SSH connection requires a running VM", vmID, vm.Status)
	}

	// Try to get the SSH user in order of precedence
	sshUser := firstNonEmpty(
		sshUserFlag,
		os.Getenv("REPLICATED_SSH_USER"),
		os.Getenv("GITHUB_ACTOR"),
		os.Getenv("GITHUB_USER"),
	)

	sshEndpoint := vm.DirectSSHEndpoint
	if sshUser != "" {
		sshEndpoint = fmt.Sprintf("%s@%s", sshUser, vm.DirectSSHEndpoint)
	}

	cmd := exec.Command("ssh", sshEndpoint, "-p", fmt.Sprintf("%d", vm.DirectSSHPort))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// firstNonEmpty returns the first non-empty string from the provided values
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
