package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	golog "log"
	"os"
	"strings"
	"sync"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/logger"
	"github.com/replicatedhq/replicated/pkg/types"
	tsloader "github.com/replicatedhq/troubleshoot/pkg/loader"
	"github.com/replicatedhq/troubleshoot/pkg/preflight"
	// troubleshootv1beta2 "github.com/replicatedhq/troubleshoot/pkg/apis/troubleshoot/v1beta2"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/release"
)

func (r *runners) InitClusterPrepare(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prepare",
		Short: "prepare cluster for testing",
		Long: `The cluster prepare command will provision a cluster and install 
a local helm chart with a custom values.yaml and custom replicated sdk entitlements.

This is a higher level CLI command that is useful in CI when you have a Helm chart and
want it running in a variety of clusters.

For more control over the workflow, consider using the cluster create command and 
using kubectl and helm CLI tools to install your application.

Example:

replicated cluster prepare --distribution eks --version 1.27 --instance-type c6.xlarge --node-count 3 \
	  --entitlement seat_count=100 --entitlement license_type=enterprise \
	  --chart ./my-helm-chart --values ./values.yaml --set chart-key=value --set chart-key2=value2`,
		RunE:         r.prepareCluster,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.prepareClusterName, "name", "", "cluster name")
	cmd.Flags().StringVar(&r.args.prepareClusterKubernetesDistribution, "distribution", "kind", "Kubernetes distribution of the cluster to provision")
	cmd.Flags().StringVar(&r.args.prepareClusterKubernetesVersion, "version", "v1.25.3", "Kubernetes version to provision (format is distribution dependent)")
	cmd.Flags().IntVar(&r.args.prepareClusterNodeCount, "node-count", int(1), "Node count")
	cmd.Flags().Int64Var(&r.args.prepareClusterVCpus, "vcpu", int64(4), "Number of vCPUs to request per node")
	cmd.Flags().Int64Var(&r.args.prepareClusterMemoryGiB, "memory", int64(4), "Memory (GiB) to request per node")
	cmd.Flags().Int64Var(&r.args.prepareClusterDiskGiB, "disk", int64(50), "Disk Size (GiB) to request per node (Default: 50)")
	cmd.Flags().StringVar(&r.args.prepareClusterTTL, "ttl", "2h", "Cluster TTL (duration, max 48h)")
	cmd.Flags().StringVar(&r.args.prepareClusterInstanceType, "instance-type", "", "the type of instance to use for cloud-based clusters (e.g. x5.xlarge)")

	// todo maybe remove
	cmd.Flags().StringVar(&r.args.prepareClusterID, "cluster-id", "", "cluster id")

	cmd.Flags().StringArrayVar(&r.args.prepareClusterEntitlements, "entitlement", []string{}, "entitlements to add to the application when deploying")

	// for premium plans (kots etc)
	cmd.Flags().StringVar(&r.args.prepareClusterYaml, "yaml", "", "The YAML config for this release. Use '-' to read from stdin. Cannot be used with the --yaml-file flag.")
	cmd.Flags().StringVar(&r.args.prepareClusterYamlFile, "yaml-file", "", "The YAML config for this release. Cannot be used with the --yaml flag.")
	cmd.Flags().StringVar(&r.args.prepareClusterYamlDir, "yaml-dir", "", "The directory containing multiple yamls for a Kots release. Cannot be used with the --yaml flag.")

	// for builders plan (chart only)
	cmd.Flags().StringVar(&r.args.prepareClusterChart, "chart", "", "path to the helm chart to deploy")
	cmd.Flags().StringArrayVar(&r.args.prepareClusterValuesPath, "values", []string{}, "path to the values.yaml file to use when deploying the chart")
	cmd.Flags().StringArrayVar(&r.args.prepareClusterValueItems, "set", []string{}, "set a helm value (e.g. --set foo=bar)")

	// TODO add json output

	return cmd
}

func (r *runners) prepareCluster(_ *cobra.Command, args []string) error {
	log := logger.NewLogger(r.w)

	// this only supports charts and builders teams for now (no kots&kurl)
	if r.args.prepareClusterChart == "" {
		return errors.New(`The "cluster prepare" command only supports builders plan (direct helm install) at this time.`)
	}

	release, err := prepareRelease(r, log)
	if err != nil {
		return errors.Wrap(err, "prepare release")
	}

	wg := sync.WaitGroup{}
	clusterName := ""
	clusterID := ""

	if r.args.prepareClusterID == "" {
		log.ChildActionWithoutSpinner("SEQUENCE: %d", release.Sequence)

		if r.args.createClusterName == "" {
			r.args.prepareClusterName = generateClusterName()
		}

		clusterOpts := kotsclient.CreateClusterOpts{
			Name:                   r.args.prepareClusterName,
			KubernetesDistribution: r.args.prepareClusterKubernetesDistribution,
			KubernetesVersion:      r.args.prepareClusterKubernetesVersion,
			NodeCount:              r.args.prepareClusterNodeCount,
			VCpus:                  r.args.prepareClusterVCpus,
			MemoryGiB:              r.args.prepareClusterMemoryGiB,
			DiskGiB:                r.args.prepareClusterDiskGiB,
			TTL:                    r.args.prepareClusterTTL,
			InstanceType:           r.args.prepareClusterInstanceType,
		}
		clusterName = r.args.prepareClusterName

		cl, err := createCluster(r.kotsAPI, clusterOpts, r.args.createClusterWaitDuration)
		if err != nil {
			return errors.Wrap(err, "create cluster")
		}
		clusterID = cl.ID

		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			// TODO should this 5 minutes be confiugrable?
			if _, err := waitForCluster(r.kotsAPI, cl.ID, time.Minute*5); err != nil {
				fmt.Printf("Failed to wait for cluster %s to be ready: %v\n", cl.ID, err)
			}

			fmt.Fprintf(r.w, "Cluster %s (%s) created.\n", cl.Name, cl.ID)
		}(&wg)
	} else {
		// need to get the cluster info to get the name to pass to the customer
		clusters, err := r.kotsAPI.ListClusters(false, nil, nil)
		if err != nil {
			return errors.Wrap(err, "list clusters")
		}
		for _, cl := range clusters {
			if cl.ID == r.args.prepareClusterID {
				clusterName = cl.Name
				clusterID = cl.ID
				break
			}
		}
	}

	if clusterID == "" {
		return errors.New("failed to find cluster")
	}

	a, err := r.kotsAPI.GetApp(r.appID)
	if err != nil {
		return errors.Wrap(err, "get app")
	}

	// create a test customer with the correct entitlement values
	email := fmt.Sprintf("%s@relicated.com", clusterName)
	customerOpts := kotsclient.CreateCustomerOpts{
		Name:        clusterName,
		ChannelID:   "",
		AppID:       a.ID,
		LicenseType: "test",
		Email:       email,
	}
	customer, err := r.api.CreateCustomer(r.appType, customerOpts)
	if err != nil {
		return errors.Wrap(err, "create customer")
	}

	_, err = fmt.Fprintf(r.w, "Customer %s (%s) created.\n", customer.Name, customer.ID)
	if err != nil {
		return errors.Wrap(err, "write to stdout")
	}

	// wait for the wait group
	wg.Wait()

	// get the kubeconfig
	kubeconfig, err := r.kotsAPI.GetClusterKubeconfig(clusterID)
	if err != nil {
		return errors.Wrap(err, "get cluster kubeconfig")
	}

	isReleaseReady, err := isReleaseReadyToInstall(r, log, *release)
	if err != nil || !isReleaseReady {
		return errors.Wrap(err, "release not ready")
	}

	// write the kubeconfig to a file
	kubeconfigFile, err := os.CreateTemp("", "kubeconfig")
	if err != nil {
		return errors.Wrap(err, "create kubeconfig file")
	}
	defer os.Remove(kubeconfigFile.Name())
	if err := os.WriteFile(kubeconfigFile.Name(), kubeconfig, 0644); err != nil {
		return errors.Wrap(err, "write kubeconfig file")
	}

	// write registry credentials to a file
	authConfig := dockertypes.AuthConfig{
		Username: email,
		Password: customer.InstallationID,
		Auth:     base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", email, customer.InstallationID))),
	}
	encodedAuthConfigJSON, err := json.Marshal(authConfig)
	if err != nil {
		return errors.Wrap(err, "failed to marshal auth config")
	}

	registryHostname := `registry.replicated.com`
	if os.Getenv("REPLICATED_REGISTRY_ORIGIN") != "" {
		registryHostname = os.Getenv("REPLICATED_REGISTRY_ORIGIN")
	}

	configJSON := fmt.Sprintf(`{"auths":{"%s":%s}}`, registryHostname, encodedAuthConfigJSON)

	credentialsFile, err := os.CreateTemp("", "credentials")
	if err != nil {
		return errors.Wrap(err, "failed to create credentials file")
	}
	defer os.Remove(credentialsFile.Name())

	if _, err := credentialsFile.Write([]byte(configJSON)); err != nil {
		return errors.Wrap(err, "failed to write credentials file")
	}

	//  TODO: implement valuesPath 

	// build the values map
	vals, err := buildValuesMap(r.args.prepareClusterValueItems)
	if err != nil {
		return errors.Wrap(err, "build values map")
	}

	// install the chart or application
	for _, chart := range release.Charts {

		//  dry run and get manifest
		dryRunRelease, err := installChartRelease(a.Slug, release.Sequence, chart.Name, vals, kubeconfigFile, credentialsFile, registryHostname, true)
		if err != nil {
			return errors.Wrap(err, "dry run release")
		}

		ctx := context.Background()
		kinds, err := tsloader.LoadSpecs(ctx, tsloader.LoadOptions{
			RawSpec: string(dryRunRelease.Manifest),
		})
		if err != nil {
			return err
		}

		p := preflight.PreflightSpecs{}
		for _, kind := range kinds.PreflightsV1Beta2 {
			p.PreflightSpec = preflight.ConcatPreflightSpec(p.PreflightSpec, &kind)
		}

		// run the preflights
		if p.PreflightSpec != nil {
			fmt.Fprintln(r.w)
			log.ActionWithSpinner("Running preflights")
			// results, err := preflight.Run(p.PreflightSpec, preflight.RunOptions{
			// 	Namespace: "default",
			// 	Timeout:   time.Minute * 5,
			// 	Kubectl:   r.kubectl,
			// 	ProgressChan: func(progress string) {
			// 		fmt.Fprintf(r.w, "%s\n", progress)
			// 	},
			// })
			// if err != nil {
			// 	log.FinishSpinnerWithError()
			// 	return errors.Wrap(err, "run preflights")
			// }
			log.FinishSpinner()

			// if results.IsFail() {
			// 	return errors.New("preflights failed")
			// }
		}



		
		



		release, err := installChartRelease(a.Slug, release.Sequence, chart.Name, vals, kubeconfigFile, credentialsFile, registryHostname, false)
		if err != nil {
			return errors.Wrap(err, "install release")
		}

		// print the output
		fmt.Fprintf(r.w, "%s\n", release.Info.Notes)
	}

	return nil
}

func installChartRelease(appSlug string, releaseSequence int64, chartName string,  values map[string]interface{}, kubeconfigFile  *os.File, credentialsFile *os.File, registryHostname string, dryRun bool) (*release.Release, error) {
	settings := cli.New()
	settings.KubeConfig = kubeconfigFile.Name()

	registryClient, err := registry.NewClient(
		registry.ClientOptDebug(settings.Debug),
		registry.ClientOptWriter(os.Stdout),
		registry.ClientOptCredentialsFile(credentialsFile.Name()),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create registry client")
	}

	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), "secret", golog.Printf); err != nil {
		return nil, errors.Wrap(err, "init action config")
	}
	actionConfig.RegistryClient = registryClient

	client := action.NewInstall(actionConfig)
	client.ReleaseName = fmt.Sprintf("%s-%d", appSlug, releaseSequence)
	client.Namespace = "default"
	client.Wait = true
	client.Timeout = 5 * time.Minute
	client.DryRun = dryRun

	cp, err := client.ChartPathOptions.LocateChart(fmt.Sprintf("oci://%s/%s/release__%d/%s", registryHostname, appSlug, releaseSequence, chartName), settings)
	if err != nil {
		return nil, errors.Wrap(err, "locate chart")
	}

	chartReq, err := loader.Load(cp)
	if err != nil {
		return nil, errors.Wrap(err, "load chart")
	}

	helmRelease, err := client.Run(chartReq, values)
	if err != nil {
		return nil, errors.Wrap(err, "run helm install")
	}

	return helmRelease, nil
}

func prepareRelease(r *runners, log *logger.Logger) (*types.ReleaseInfo, error) {
	// create the release first because we want to fail early if there are linting issues
	// or if it's a builder plan team submitting a kots app or the other way around
	if r.args.prepareClusterYaml == "-" {
		bytes, err := ioutil.ReadAll(r.stdin)
		if err != nil {
			return nil, errors.Wrap(err, "read stdin")
		}
		r.args.prepareClusterYaml = string(bytes)
	}

	if r.args.prepareClusterYamlFile != "" {
		bytes, err := ioutil.ReadFile(r.args.prepareClusterYamlFile)
		if err != nil {
			return nil, errors.Wrap(err, "read release yaml file")
		}
		r.args.prepareClusterYaml = string(bytes)
	}

	if r.args.prepareClusterYamlDir != "" {
		fmt.Fprintln(r.w)
		log.ActionWithSpinner("Reading manifests from %s", r.args.prepareClusterYamlDir)
		var err error
		r.args.prepareClusterYaml, err = makeReleaseFromDir(r.args.prepareClusterYamlDir)
		if err != nil {
			log.FinishSpinnerWithError()
			return nil, errors.Wrap(err, "make release from dir")
		}
		log.FinishSpinner()
	}

	if r.args.prepareClusterChart != "" {
		fmt.Fprintln(r.w)
		log.ActionWithSpinner("Reading chart from %s", r.args.prepareClusterChart)
		var err error
		r.args.prepareClusterYaml, err = makeReleaseFromChart(r.args.prepareClusterChart)
		if err != nil {
			log.FinishSpinnerWithError()
			return nil, errors.Wrap(err, "make release from chart")
		}
		log.FinishSpinner()
	}

	log.ActionWithSpinner("Creating Release")
	release, err := r.api.CreateRelease(r.appID, r.appType, r.args.prepareClusterYaml)
	if err != nil {
		log.FinishSpinnerWithError()
		return nil, errors.Wrap(err, "create release")
	}
	log.FinishSpinner()

	return release, nil
}

func isReleaseReadyToInstall(r *runners, log *logger.Logger, release types.ReleaseInfo) (bool, error) {
	if len(release.Charts) == 0 {
		return false, errors.New("no charts found in release")
	}

	timeout := time.Duration(10*len(release.Charts)) * time.Second
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			return false, errors.Errorf("timed out waiting for release to be ready after %s", timeout)
		default:
			appRelease, err := r.api.GetRelease(r.appID, r.appType, release.Sequence)
			if err != nil {
				return false, errors.Wrap(err, "failed to get release")
			}

			ready, err := areReleaseChartsPushed(appRelease.Charts)
			if err != nil {
				return false, errors.Wrap(err, "failed to check release charts")
			} else if ready {
				return true, nil
			}

			time.Sleep(time.Second * 2)
		}
	}
}

func areReleaseChartsPushed(charts []types.Chart) (bool, error) {
	if len(charts) == 0 {
		return false, errors.New("no charts found in release")
	}

	pushedChartsCount := 0
	for _, chart := range charts {
		switch chart.Status {
		case types.ChartStatusPushed:
			pushedChartsCount++
		case types.ChartStatusUnknown, types.ChartStatusPushing:
			// wait for the chart to be pushed
		case types.ChartStatusError:
			return false, errors.Errorf("chart %q failed to push: %s", chart.Name, chart.Error)
		default:
			return false, errors.Errorf("unknown release chart status %q", chart.Status)
		}
	}

	return pushedChartsCount == len(charts), nil
}

// good to use helm library for this
// https://github.com/helm/helm/blob/main/pkg/cli/values/options.go
func buildValuesMap(valueItems []string) (map[string]interface{}, error){

	vals := map[string]interface{}{}
	for _, set := range valueItems {
		setParts := strings.SplitN(set, "=", 2)
		if len(setParts) != 2 {
			return nil, errors.Errorf("invalid set %q", set)
		}

		key := setParts[0]
		val := setParts[1]
		if !strings.Contains(key, ".") {
			vals[key] = setParts[1]
			continue
		}

		// convert the key.part set command to a map[string]interface{}
		parts := strings.Split(key, ".")
		m := vals
		for i, part := range parts {
			if i == len(parts)-1 {
				m[part] = val
				continue
			}

			if _, ok := m[part]; !ok {
				m[part] = map[string]interface{}{}
			}

			m = m[part].(map[string]interface{})
		}
	}

	return vals, nil
}