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
		Use:   "shell [ID]",
		Short: "Open a new shell with kubeconfig configured.",
		Long: `The 'shell' command opens a new shell session with the kubeconfig configured for the specified test cluster. This allows you to have immediate kubectl access to the cluster within the shell environment.

You can either specify the cluster ID directly or provide the cluster name to resolve the corresponding cluster ID. The shell will inherit your existing environment and add the necessary kubeconfig context for interacting with the Kubernetes cluster.

Once inside the shell, you can use 'kubectl' to interact with the cluster. To exit the shell, press Ctrl-D or type 'exit'. When the shell closes, the kubeconfig will be reset back to your default configuration.`,
		Example: `  # Open a shell for a cluster by ID
  replicated cluster shell CLUSTER_ID

  # Open a shell for a cluster by name
  replicated cluster shell --name "My Cluster"`,
		RunE:              r.shellCluster,
		ValidArgsFunction: r.completeClusterIDs,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.shellClusterName, "name", "", "name of the cluster to have kubectl access to.")
	cmd.RegisterFlagCompletionFunc("name", r.completeClusterNames)

	cmd.Flags().StringVar(&r.args.shellClusterID, "id", "", "id of the cluster to have kubectl access to (when name is not provided)")
	cmd.RegisterFlagCompletionFunc("id", r.completeClusterIDs)

	return cmd
}

func (r *runners) shellCluster(_ *cobra.Command, args []string) error {
	shellCmd := os.Getenv("SHELL")
	if shellCmd == "" {
		return errors.New("SHELL environment is required for shell command")
	}

	// by default, we look at args[0] as the id
	// but if it's not provided, we look for a viper flag named "name" and use it
	// as the name of the cluster, not the id
	clusterID := ""
	if len(args) > 0 {
		clusterID = args[0]
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
		return errors.Errorf("cluster %s is not running, please check the cluster status", clusterID)
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
	if err := tmpFile.Chmod(0600); err != nil {
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
