package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	homeEnvVar       = "HOME"
	KubeConfigEnvVar = "KUBECONFIG"
)

func (r *runners) InitClusterKubeconfig(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kubeconfig [ID_OR_NAME]",
		Short: "Download credentials for a test cluster.",
		Long: `The 'cluster kubeconfig' command downloads the credentials (kubeconfig) required to access a test cluster. You can either merge these credentials into your existing kubeconfig file or save them as a new file.

This command ensures that the kubeconfig is correctly configured for use with your Kubernetes tools. You can specify the cluster by ID or name directly as an argument, or by using the '--id' or '--name' flags. Additionally, the kubeconfig can be written to a specific file path or printed to stdout.

You can also use this command to automatically update your current Kubernetes context with the downloaded credentials.`,
		Example: `# Download and merge kubeconfig into your existing configuration
replicated cluster kubeconfig CLUSTER_ID_OR_NAME

# Save the kubeconfig to a specific file
replicated cluster kubeconfig CLUSTER_ID_OR_NAME --output-path ./kubeconfig

# Print the kubeconfig to stdout
replicated cluster kubeconfig CLUSTER_ID_OR_NAME --stdout

# Download kubeconfig for a cluster by name using a flag
replicated cluster kubeconfig --name "My Cluster"

# Download kubeconfig for a cluster by ID using a flag
replicated cluster kubeconfig --id CLUSTER_ID`,
		RunE:              r.kubeconfigCluster,
		ValidArgsFunction: r.completeClusterIDsAndNames,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.kubeconfigClusterName, "name", "", "name of the cluster to download credentials for (when id is not provided)")
	cmd.RegisterFlagCompletionFunc("name", r.completeClusterNames)
	cmd.Flag("name").Deprecated = "use ID_OR_NAME arguments instead"

	cmd.Flags().StringVar(&r.args.kubeconfigClusterID, "id", "", "id of the cluster to download credentials for (when name is not provided)")
	cmd.RegisterFlagCompletionFunc("id", r.completeClusterIDs)

	cmd.Flags().StringVar(&r.args.kubeconfigPath, "output-path", "", "path to kubeconfig file to write to, if not provided, it will be merged into your existing kubeconfig")
	cmd.Flags().BoolVar(&r.args.kubeconfigStdout, "stdout", false, "write kubeconfig to stdout")

	return cmd
}

func (r *runners) kubeconfigCluster(_ *cobra.Command, args []string) error {
	// by default, we look at args[0] as the id or name
	// but if it's not provided, we look for a flag named "name" or "id"
	clusterID := ""
	if len(args) > 0 {
		var err error
		clusterID, err = r.getClusterIDFromArg(args[0])
		if err != nil {
			return errors.Wrap(err, "get cluster id from arg")
		}
	} else if r.args.kubeconfigClusterName != "" {
		clusters, err := r.kotsAPI.ListClusters(false, nil, nil)
		if errors.Cause(err) == platformclient.ErrForbidden {
			return ErrCompatibilityMatrixTermsNotAccepted
		} else if err != nil {
			return errors.Wrap(err, "list clusters")
		}
		for _, cluster := range clusters {
			if cluster.Name == r.args.kubeconfigClusterName {
				clusterID = cluster.ID
				break
			}
		}
	} else if r.args.kubeconfigClusterID != "" {
		clusterID = r.args.kubeconfigClusterID
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

	if r.args.kubeconfigStdout && r.args.kubeconfigPath != "" {
		return errors.New("cannot use both --stdout and --output-path")
	}

	if r.args.kubeconfigStdout {
		fmt.Println(string(kubeconfig))
		return nil
	}

	if r.args.kubeconfigPath != "" {
		dir := filepath.Dir(r.args.kubeconfigPath)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return errors.Wrap(err, "create kubeconfig dir")
			}
		}
		if err := os.WriteFile(r.args.kubeconfigPath, kubeconfig, 0644); err != nil {
			return errors.Wrap(err, "write kubeconfig")
		}

		fmt.Printf("kubeconfig written to %s\n", r.args.kubeconfigPath)
		return nil
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

	replicatedLoadingRules := clientcmd.ClientConfigLoadingRules{
		ExplicitPath: tmpFile.Name(),
	}
	replicatedConfig, err := replicatedLoadingRules.Load()
	if err != nil {
		return errors.Wrap(err, "load kubeconfig")
	}

	kubeconfigPaths := getKubeconfigPaths()
	backupPaths := []string{}

	// back up the curent kubeconfigs
	for _, kubeconfigPath := range kubeconfigPaths {
		backupPath := kubeconfigPath + ".replicated_backup"

		fi, err := os.Stat(kubeconfigPath)
		var pathError *fs.PathError
		if errors.As(err, &pathError) {
			// file doesn't exist, nothing to backup
			continue
		} else if err != nil {
			return errors.Wrap(err, "stat kubeconfig")
		}
		data, err := os.ReadFile(kubeconfigPath)
		if err != nil {
			return errors.Wrap(err, "read kubeconfig")
		}

		if err := os.WriteFile(backupPath, data, fi.Mode()); err != nil {
			return errors.Wrap(err, "write backup kubeconfig")
		}

		backupPaths = append(backupPaths, backupPath)
	}
	defer func() {
		for _, backupPath := range backupPaths {
			err := os.Remove(backupPath)
			if err != nil {
				fmt.Printf("failed to remove backup kubeconfig: %s\n", err.Error())
			}
		}
	}()

	// parse the current kubeconfig
	loadingRules := clientcmd.ClientConfigLoadingRules{
		Precedence: kubeconfigPaths,
	}
	mergedConfig, err := loadingRules.Load()
	if err != nil {
		return errors.Wrap(err, "load kubeconfig")
	}

	// add the replicated context
	for contextName, context := range replicatedConfig.Contexts {
		mergedConfig.Contexts[contextName] = context
	}
	// add the replicated credentials
	for credentialName, credential := range replicatedConfig.AuthInfos {
		mergedConfig.AuthInfos[credentialName] = credential
	}
	// add the replicated cluster
	for clusterName, cluster := range replicatedConfig.Clusters {
		mergedConfig.Clusters[clusterName] = cluster
	}

	mergedConfig.CurrentContext = replicatedConfig.CurrentContext

	// write the merged kubeconfig
	err = clientcmd.WriteToFile(*mergedConfig, kubeconfigPaths[0])
	if err != nil {
		return errors.Wrap(err, "write kubeconfig")
	}

	fmt.Printf(" ✓  Updated kubernetes context '%s' in '%s'\n", mergedConfig.CurrentContext, kubeconfigPaths[0])

	return nil
}

func getKubeconfigPaths() []string {
	home := getHomeDir()
	kubeconfig := []string{filepath.Join(home, ".kube", "config")}
	kubeconfigEnv := os.Getenv(KubeConfigEnvVar)
	if len(kubeconfigEnv) > 0 {
		kubeconfig = splitKubeConfigEnv(kubeconfigEnv)
	}

	return kubeconfig
}

func getHomeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}

func splitKubeConfigEnv(value string) []string {
	if runtime.GOOS == "windows" {
		return strings.Split(value, ";")
	}
	return strings.Split(value, ":")
}
