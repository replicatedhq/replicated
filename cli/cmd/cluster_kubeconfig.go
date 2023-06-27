package cmd

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	homeEnvVar       = "HOME"
	KubeConfigEnvVar = "KUBECONFIG"
)

func (r *runners) InitClusterKubeconfig(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "kubeconfig ID",
		Short:        "Download credentials for a test cluster",
		Long:         `Download credentials for a test cluster`,
		RunE:         r.kubeconfigCluster,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.kubeconfigClusterName, "name", "", "name of the cluster to download credentials for (when id is not provided)")
	cmd.Flags().StringVar(&r.args.kubeconfigClusterID, "id", "", "id of the cluster to download credentials for (when name is not provided)")
	cmd.Flags().StringVar(&r.args.kubeconfigPath, "output", "", "path to kubeconfig file to write to, if not provided, it will be merged into your existing kubeconfig")

	return cmd
}

func (r *runners) kubeconfigCluster(_ *cobra.Command, args []string) error {
	kotsRestClient := kotsclient.VendorV3Client{HTTPClient: *r.platformAPI}

	// by default, we look at args[0] as the id
	// but if it's not provided, we look for a viper flag named "name" and use it
	// as the name of the cluster, not the id
	clusterID := ""
	if len(args) > 0 {
		clusterID = args[0]
	} else if r.args.kubeconfigClusterName != "" {
		clusters, err := kotsRestClient.ListClusters(false, nil, nil)
		if err != nil {
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

	kubeconfig, err := kotsRestClient.GetClusterKubeconfig(clusterID)
	if err != nil {
		return errors.Wrap(err, "get cluster kubeconfig")
	}

	if r.args.kubeconfigPath != "" {
		dir := filepath.Dir(r.args.kubeconfigPath)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return errors.Wrap(err, "create kubeconfig dir")
			}
		}
		if err := ioutil.WriteFile(r.args.kubeconfigPath, kubeconfig, 0644); err != nil {
			return errors.Wrap(err, "write kubeconfig")
		}

		fmt.Printf("kubeconfig written to %s\n", r.args.kubeconfigPath)
		return nil
	}

	tmpFile, err := ioutil.TempFile("", "replicated-kubeconfig")
	if err != nil {
		return errors.Wrap(err, "create temp file")
	}
	defer os.Remove(tmpFile.Name())
	if err := ioutil.WriteFile(tmpFile.Name(), kubeconfig, 0644); err != nil {
		return errors.Wrap(err, "write temp file")
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
		data, err := ioutil.ReadFile(kubeconfigPath)
		if err != nil {
			return errors.Wrap(err, "read kubeconfig")
		}

		if err := ioutil.WriteFile(backupPath, data, fi.Mode()); err != nil {
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
