package kotsclient

import (
	"fmt"
	"os"
	"os/exec"
)

// SCPToVM copies a file to a VM using SCP
func (c *VendorV3Client) SCPToVM(vmID string, sshUserFlag string, localPath string, remotePath string) error {
	vm, err := c.GetVM(vmID)
	if err != nil {
		return err
	}

	if vm.DirectSSHEndpoint == "" || vm.DirectSSHPort == 0 {
		return fmt.Errorf("VM %s does not have SSH access configured", vmID)
	}

	// Try to get the SSH user in order of precedence
	sshUser := firstNonEmpty(
		sshUserFlag,
		os.Getenv("REPLICATED_SSH_USER"),
		os.Getenv("GITHUB_ACTOR"),
		os.Getenv("GITHUB_USER"),
	)

	// Format the remote destination
	remoteDestination := remotePath
	if sshUser != "" {
		remoteDestination = fmt.Sprintf("%s@%s:%s", sshUser, vm.DirectSSHEndpoint, remotePath)
	} else {
		remoteDestination = fmt.Sprintf("%s:%s", vm.DirectSSHEndpoint, remotePath)
	}

	// Build the scp command
	cmd := exec.Command("scp", "-P", fmt.Sprintf("%d", vm.DirectSSHPort), localPath, remoteDestination)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// SCPFromVM copies a file from a VM using SCP
func (c *VendorV3Client) SCPFromVM(vmID string, sshUserFlag string, remotePath string, localPath string) error {
	vm, err := c.GetVM(vmID)
	if err != nil {
		return err
	}

	if vm.DirectSSHEndpoint == "" || vm.DirectSSHPort == 0 {
		return fmt.Errorf("VM %s does not have SSH access configured", vmID)
	}

	// Try to get the SSH user in order of precedence
	sshUser := firstNonEmpty(
		sshUserFlag,
		os.Getenv("REPLICATED_SSH_USER"),
		os.Getenv("GITHUB_ACTOR"),
		os.Getenv("GITHUB_USER"),
	)

	// Format the remote source
	remoteSource := remotePath
	if sshUser != "" {
		remoteSource = fmt.Sprintf("%s@%s:%s", sshUser, vm.DirectSSHEndpoint, remotePath)
	} else {
		remoteSource = fmt.Sprintf("%s:%s", vm.DirectSSHEndpoint, remotePath)
	}

	// Build the scp command
	cmd := exec.Command("scp", "-P", fmt.Sprintf("%d", vm.DirectSSHPort), remoteSource, localPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
