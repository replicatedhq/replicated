package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func (r *runners) InitClusterShell(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shell [ID_OR_NAME]",
		Short: "Open a new shell with kubeconfig configured.",
		Long: `The 'shell' command opens a new shell session with the kubeconfig configured for the specified test cluster. This allows you to have immediate kubectl access to the cluster within the shell environment.

You can either specify the cluster ID or name directly as an argument, or provide the cluster name or ID using flags. The shell will inherit your existing environment and add the necessary kubeconfig context for interacting with the Kubernetes cluster.

Once inside the shell, you can use 'kubectl' to interact with the cluster. To exit the shell, press Ctrl-D or type 'exit'. When the shell closes, the kubeconfig will be reset back to your default configuration.`,
		Example: `# Open a shell for a cluster by ID or name
replicated cluster shell CLUSTER_ID_OR_NAME

# Open a shell for a cluster by name using a flag
replicated cluster shell --name "My Cluster"

# Open a shell for a cluster by ID using a flag
replicated cluster shell --id CLUSTER_ID`,
		RunE:              r.shellCluster,
		ValidArgsFunction: r.completeClusterIDsAndNames,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.shellClusterName, "name", "", "name of the cluster to have kubectl access to.")
	cmd.RegisterFlagCompletionFunc("name", r.completeClusterNames)
	cmd.Flag("name").Deprecated = "use ID_OR_NAME arguments instead"

	cmd.Flags().StringVar(&r.args.shellClusterID, "id", "", "id of the cluster to have kubectl access to (when name is not provided)")
	cmd.RegisterFlagCompletionFunc("id", r.completeClusterIDs)
	cmd.Flag("id").Deprecated = "use ID_OR_NAME arguments instead"

	return cmd
}

func (r *runners) shellCluster(_ *cobra.Command, args []string) error {
	shellCmd := os.Getenv("SHELL")
	if shellCmd == "" {
		return errors.New("SHELL environment is required for shell command")
	}

	// by default, we look at args[0] as the id or name
	// but if it's not provided, we look for a flag named "name" or "id"
	clusterID := ""
	if len(args) > 0 {
		var err error
		clusterID, err = r.getClusterIDFromArg(args[0])
		if err != nil {
			return errors.Wrap(err, "get cluster id from arg")
		}
	} else if r.args.shellClusterName != "" {
		clusters, err := r.kotsAPI.ListClusters(false, nil, nil)
		if errors.Cause(err) == platformclient.ErrForbidden {
			return ErrCompatibilityMatrixTermsNotAccepted
		} else if err != nil {
			return errors.Wrap(err, "list clusters")
		}
		for _, cluster := range clusters {
			if cluster.Name == r.args.shellClusterName {
				clusterID = cluster.ID
				break
			}
		}
	} else if r.args.shellClusterID != "" {
		clusterID = r.args.shellClusterID
	} else {
		return errors.New("must provide cluster id or name")
	}

	cluster, err := r.kotsAPI.GetCluster(clusterID)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return errors.Wrap(err, "get cluster")
	}

	if cluster.Status != types.ClusterStatusRunning {
		return errors.Errorf("cluster %s is not running, current status is %s", clusterID, cluster.Status)
	}

	kubeconfig, err := r.kotsAPI.GetClusterKubeconfig(clusterID)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return errors.Wrap(err, "get cluster kubeconfig")
	}

	tmpFile, err := os.CreateTemp("", "replicated-kubeconfig")
	if err != nil {
		return errors.Wrap(err, "create temp file")
	}
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()
	if _, err := tmpFile.Write(kubeconfig); err != nil {
		return errors.Wrap(err, "write kubeconfig file")
	}
	if err := tmpFile.Chmod(0o600); err != nil {
		return errors.Wrap(err, "chmod kubeconfig file")
	}

	shellExec := exec.Command(shellCmd)
	shellExec.Env = os.Environ()
	fmt.Println("Starting new shell with KUBECONFIG. Press Ctrl-D or type 'exit' when done to end the shell and the connection to the server")
	shellPty, err := pty.Start(shellExec)
	if err != nil {
		return errors.Wrap(err, "failed to start shell")
	}

	// Handle pty size.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if err := pty.InheritSize(os.Stdin, shellPty); err != nil {
				log.Printf("error resizing pty: %s", err)
			}
		}
	}()
	ch <- syscall.SIGWINCH // Initial resize.
	defer func() {
		signal.Stop(ch)
		close(ch)
	}()

	// Set stdin to raw mode.
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = term.Restore(int(os.Stdin.Fd()), oldState)
		fmt.Println("Exiting Compatibility Matrix Shell (this will reset your kubeconfig back to default)")
	}()

	// Setup the shell
	setupCmd := fmt.Sprintf("export KUBECONFIG=%s\n", tmpFile.Name())
	_, _ = io.WriteString(shellPty, setupCmd)
	_, _ = io.CopyN(io.Discard, shellPty, 2*int64(len(setupCmd))) // Don't print to screen, terminal will echo anyway

	// Copy stdin to the pty and the pty to stdout.
	go io.Copy(shellPty, os.Stdin)
	go io.Copy(os.Stdout, shellPty)

	_ = shellExec.Wait()
	return nil
}
