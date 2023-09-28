package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	golog "log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/pkg/errors"
	releaseTypes "github.com/replicatedhq/replicated/pkg/kots/release/types"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/kotsutil"
	"github.com/replicatedhq/replicated/pkg/logger"
	"github.com/replicatedhq/replicated/pkg/types"
	troubleshootanalyze "github.com/replicatedhq/troubleshoot/pkg/analyze"
	troubleshootv1beta2 "github.com/replicatedhq/troubleshoot/pkg/apis/troubleshoot/v1beta2"
	troubleshootcollect "github.com/replicatedhq/troubleshoot/pkg/collect"
	troubleshootloader "github.com/replicatedhq/troubleshoot/pkg/loader"
	"github.com/replicatedhq/troubleshoot/pkg/preflight"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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
		RunE: r.prepareCluster,
	}

	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		return validateClusterPrepareFlags(r.args)
	}

	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.prepareClusterName, "name", "", "Cluster name")
	cmd.Flags().StringVar(&r.args.prepareClusterKubernetesDistribution, "distribution", "", "Kubernetes distribution of the cluster to provision")
	cmd.Flags().StringVar(&r.args.prepareClusterKubernetesVersion, "version", "", "Kubernetes version to provision (format is distribution dependent)")
	cmd.Flags().IntVar(&r.args.prepareClusterNodeCount, "node-count", int(1), "Node count.")
	cmd.Flags().Int64Var(&r.args.prepareClusterDiskGiB, "disk", int64(50), "Disk Size (GiB) to request per node.")
	cmd.Flags().StringVar(&r.args.prepareClusterTTL, "ttl", "", "Cluster TTL (duration, max 48h)")
	cmd.Flags().StringVar(&r.args.prepareClusterInstanceType, "instance-type", "", "the type of instance to use clusters (e.g. x5.xlarge)")
	cmd.Flags().DurationVar(&r.args.prepareClusterWaitDuration, "wait", time.Minute*5, "Wait duration for cluster to be ready.")

	// todo maybe remove
	cmd.Flags().StringVar(&r.args.prepareClusterID, "cluster-id", "", "The ID of an existing cluster to use instead of creating a new one.")

	cmd.Flags().StringSliceVar(&r.args.prepareClusterEntitlements, "entitlements", []string{}, "The entitlements to set on the customer. Can be specified multiple times.")

	// for premium plans (kots etc)
	cmd.Flags().StringVar(&r.args.prepareClusterYaml, "yaml", "", "The YAML config for this release. Use '-' to read from stdin. Cannot be used with the --yaml-file flag.")
	cmd.Flags().StringVar(&r.args.prepareClusterYamlFile, "yaml-file", "", "The YAML config for this release. Cannot be used with the --yaml flag.")
	cmd.Flags().StringVar(&r.args.prepareClusterYamlDir, "yaml-dir", "", "The directory containing multiple yamls for a KOTS release. Cannot be used with the --yaml flag.")
	cmd.Flags().StringVar(&r.args.prepareClusterKotsConfigValuesFile, "config-values-file", "", "Path to a manifest containing config values (must be apiVersion: kots.io/v1beta1, kind: ConfigValues).")
	cmd.Flags().StringVar(&r.args.prepareClusterKotsSharedPassword, "shared-password", "", "Shared password for the KOTS admin console.")

	// for builders plan (chart only)
	cmd.Flags().StringVar(&r.args.prepareClusterChart, "chart", "", "Path to the helm chart to deploy")
	addValueOptionsFlags(cmd.Flags(), &r.args.prepareClusterValueOpts)

	cmd.Flags().StringVar(&r.args.prepareClusterNamespace, "namespace", "default", "The namespace into which to deploy the KOTS application or Helm chart.")
	cmd.Flags().DurationVar(&r.args.prepareClusterAppReadyTimeout, "app-ready-timeout", time.Minute*5, "Timeout to wait for the application to be ready. Must be in Go duration format (e.g., 10s, 2m).")

	_ = cmd.MarkFlagRequired("distribution")

	// TODO add json output
	return cmd
}

// https://github.com/helm/helm/blob/37cc2fa5cefb7f5bb97905b09a2a19b8c05c989f/cmd/helm/flags.go#L45
func addValueOptionsFlags(f *pflag.FlagSet, v *values.Options) {
	f.StringSliceVar(&v.ValueFiles, "values", []string{}, "Specify values in a YAML file or a URL (can specify multiple).")
	f.StringArrayVar(&v.Values, "set", []string{}, "Set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2).")
	f.StringArrayVar(&v.StringValues, "set-string", []string{}, "Set STRING values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2).")
	f.StringArrayVar(&v.FileValues, "set-file", []string{}, "Set values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2).")
	f.StringArrayVar(&v.JSONValues, "set-json", []string{}, "Set JSON values on the command line (can specify multiple or separate values with commas: key1=jsonval1,key2=jsonval2).")
	f.StringArrayVar(&v.LiteralValues, "set-literal", []string{}, "Set a literal STRING value on the command line.")
}

func validateClusterPrepareFlags(args runnerArgs) error {
	if args.prepareClusterChart == "" && args.prepareClusterYaml == "" && args.prepareClusterYamlFile == "" && args.prepareClusterYamlDir == "" {
		return errors.New("The --chart, --yaml, --yaml-file, or --yaml-dir flag is required")
	}

	if args.prepareClusterChart != "" && (args.prepareClusterYaml != "" || args.prepareClusterYamlFile != "" || args.prepareClusterYamlDir != "") {
		return errors.New("The --chart flag cannot be used with the --yaml, --yaml-file, or --yaml-dir flag")
	}

	if args.prepareClusterYaml != "" && (args.prepareClusterYamlFile != "" || args.prepareClusterYamlDir != "") {
		return errors.New("The --yaml flag cannot be used with the --yaml-file or --yaml-dir flag")
	}

	if args.prepareClusterYamlFile != "" && args.prepareClusterYamlDir != "" {
		return errors.New("The --yaml-file flag cannot be used with the --yaml-dir flag")
	}

	if args.prepareClusterChart != "" {
		if args.prepareClusterKotsSharedPassword != "" {
			return errors.New("The --shared-password flag cannot be used when deploying a Helm chart")
		}

		if args.prepareClusterKotsConfigValuesFile != "" {
			return errors.New("The --config-values-file flag cannot be used when deploying a Helm chart")
		}
	} else {
		if args.prepareClusterKotsSharedPassword == "" {
			return errors.New("The --shared-password flag is required when deploying a KOTS app")
		}

		if len(args.prepareClusterValueOpts.FileValues) > 0 || len(args.prepareClusterValueOpts.JSONValues) > 0 || len(args.prepareClusterValueOpts.LiteralValues) > 0 ||
			len(args.prepareClusterValueOpts.StringValues) > 0 || len(args.prepareClusterValueOpts.Values) > 0 || len(args.prepareClusterValueOpts.ValueFiles) > 0 {
			return errors.New("The --set, --set-file, --set-json, --set-literal, --set-string, and --values flags cannot be used when deploying a KOTS app")
		}
	}

	return nil
}

func (r *runners) prepareCluster(_ *cobra.Command, args []string) error {
	log := logger.NewLogger(r.w)

	release, err := prepareRelease(r, log)
	if err != nil {
		return errors.Wrap(err, "prepare release")
	}

	wg := sync.WaitGroup{}
	clusterName := ""
	clusterID := ""

	if r.args.prepareClusterID == "" {
		if r.args.prepareClusterName == "" {
			r.args.prepareClusterName = generateClusterName()
		}

		clusterOpts := kotsclient.CreateClusterOpts{
			Name:                   r.args.prepareClusterName,
			KubernetesDistribution: r.args.prepareClusterKubernetesDistribution,
			KubernetesVersion:      r.args.prepareClusterKubernetesVersion,
			NodeCount:              r.args.prepareClusterNodeCount,
			DiskGiB:                r.args.prepareClusterDiskGiB,
			TTL:                    r.args.prepareClusterTTL,
			InstanceType:           r.args.prepareClusterInstanceType,
		}
		clusterName = r.args.prepareClusterName

		cl, err := r.createAndWaitForCluster(clusterOpts)
		if err != nil {
			return errors.Wrap(err, "create cluster")
		}
		clusterID = cl.ID

		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			if _, err := waitForCluster(r.kotsAPI, cl.ID, r.args.prepareClusterWaitDuration); err != nil {
				fmt.Printf("Failed to wait for cluster %s to be ready: %v\n", cl.ID, err)
			}

			log.ActionWithoutSpinner("Cluster %s (%s) created.\n", cl.Name, cl.ID)
		}(&wg)
	} else {
		// need to get the cluster info to get the name to pass to the customer
		cluster, err := r.kotsAPI.GetCluster(r.args.prepareClusterID)
		if err != nil {
			return errors.Wrap(err, "get cluster")
		}
		clusterName = cluster.Name
		clusterID = cluster.ID
	}

	if clusterID == "" {
		return errors.New("failed to find cluster")
	}

	log.ActionWithSpinner("Waiting for release to be ready")
	appRelease, err := getReadyAppRelease(r, log, *release)
	if err != nil || appRelease == nil {
		log.FinishSpinnerWithError()
		if err == nil {
			return errors.New("release not ready")
		}
		return errors.Wrap(err, "release not ready")
	}
	log.FinishSpinner()

	var entitlements []kotsclient.EntitlementValue
	for _, set := range r.args.prepareClusterEntitlements {
		setParts := strings.SplitN(set, "=", 2)
		if len(setParts) != 2 {
			return errors.Errorf("invalid entitlement %q", set)
		}
		entitlements = append(entitlements, kotsclient.EntitlementValue{
			Name:  setParts[0],
			Value: setParts[1],
		})
	}

	// create a test customer with the correct entitlement values
	email := fmt.Sprintf("%s@replicated.com", clusterName)
	customerOpts := kotsclient.CreateCustomerOpts{
		Name:                clusterName,
		ChannelID:           "",
		AppID:               r.appID,
		LicenseType:         "test",
		Email:               email,
		EntitlementValues:   entitlements,
		IsKotInstallEnabled: true,
	}

	if appRelease.IsHelmOnly {
		customerOpts.IsKotInstallEnabled = false
	}
	log.ActionWithSpinner("Creating Customer")
	customer, err := r.api.CreateCustomer(r.appType, customerOpts)
	if err != nil {
		log.FinishSpinnerWithError()
		return errors.Wrap(err, "failed to create customer")
	}
	log.FinishSpinner()

	log.ChildActionWithoutSpinner("Customer %s (%s) created.\n", customer.Name, customer.ID)

	// wait for the wait group
	wg.Wait()

	kubeConfig, err := r.kotsAPI.GetClusterKubeconfig(clusterID)
	if err != nil {
		return errors.Wrap(err, "failed to get cluster kubeconfig")
	}

	if appRelease.IsHelmOnly {
		if err := installBuilderApp(r, log, kubeConfig, customer, appRelease); err != nil {
			return errors.Wrap(err, "failed to install builder app")
		}
	} else {
		if err := installKotsApp(r, log, kubeConfig, customer, appRelease); err != nil {
			return errors.Wrap(err, "failed to install kots")
		}
	}

	return nil
}

func prepareRelease(r *runners, log *logger.Logger) (*types.ReleaseInfo, error) {
	// create the release first because we want to fail early if there are linting issues
	// or if it's a builder plan team submitting a kots app or the other way around
	if r.args.prepareClusterYaml == "-" {
		bytes, err := io.ReadAll(r.stdin)
		if err != nil {
			return nil, errors.Wrap(err, "read stdin")
		}
		r.args.prepareClusterYaml = string(bytes)
	}

	if r.args.prepareClusterYamlFile != "" {
		bytes, err := os.ReadFile(r.args.prepareClusterYamlFile)
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

	log.ChildActionWithoutSpinner("SEQUENCE: %d", release.Sequence)

	return release, nil
}

func getReadyAppRelease(r *runners, log *logger.Logger, release types.ReleaseInfo) (*types.AppRelease, error) {
	timeout := time.Duration(10 * time.Second)
	if len(release.Charts) > 0 {
		timeout = time.Duration(10*len(release.Charts)) * time.Second
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			return nil, errors.Errorf("timed out waiting for release to be ready after %s", timeout)
		default:
			appRelease, err := r.api.GetRelease(r.appID, r.appType, release.Sequence)
			if err != nil {
				return nil, errors.Wrap(err, "failed to get release")
			}

			ready, err := areReleaseChartsPushed(appRelease.Charts)
			if err != nil {
				return nil, errors.Wrap(err, "failed to check release charts")
			}

			if ready {
				return appRelease, nil
			}

			time.Sleep(time.Second * 2)
		}
	}
}

func areReleaseChartsPushed(charts []types.Chart) (bool, error) {
	pushedChartsCount := 0
	for _, chart := range charts {
		switch chart.Status {
		case types.ChartStatusPushed, types.ChartStatusSubchart:
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

func runPreflights(ctx context.Context, r *runners, log *logger.Logger, kubeConfig *rest.Config, release *release.Release) error {
	tsKinds, err := troubleshootloader.LoadSpecs(ctx, troubleshootloader.LoadOptions{
		RawSpec: string(release.Manifest),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to load specs from helm chart release %s", release.Name)
	}

	var preflightSpec = new(troubleshootv1beta2.Preflight)
	for _, kind := range tsKinds.PreflightsV1Beta2 {
		preflightSpec = preflight.ConcatPreflightSpec(preflightSpec, &kind)
	}

	if preflightSpec == nil || len(preflightSpec.Spec.Analyzers) == 0 {
		return nil
	}

	progressChan := make(chan interface{}, 0)
	defer close(progressChan)

	go func() {
		for {
			msg, ok := <-progressChan
			if !ok {
				return
			}

			if err, ok := msg.(error); ok {
				fmt.Fprintf(r.w, "Error running preflights: %v\n", err)
			}
		}
	}()

	preflightSpec.Spec.Collectors = troubleshootcollect.DedupCollectors(preflightSpec.Spec.Collectors)
	preflightSpec.Spec.Analyzers = troubleshootanalyze.DedupAnalyzers(preflightSpec.Spec.Analyzers)

	collectOpts := preflight.CollectOpts{
		Namespace:              "",
		IgnorePermissionErrors: true,
		ProgressChan:           progressChan,
		KubernetesRestConfig:   kubeConfig,
	}

	collectResults, err := preflight.Collect(collectOpts, preflightSpec)
	if err != nil {
		if !errors.Is(err, troubleshootcollect.ErrInsufficientPermissionsToRun) {
			return errors.Wrap(err, "failed to collect preflight data")
		}

		clusterCollectResult, ok := collectResults.(preflight.ClusterCollectResult)
		if !ok {
			return errors.Errorf("unexpected preflight collector result type: %T", collectResults)
		}

		if errors.Is(err, troubleshootcollect.ErrInsufficientPermissionsToRun) {
			log.Info("skipping analyze due to RBAC errors")
			for _, collector := range clusterCollectResult.Collectors {
				for _, e := range collector.GetRBACErrors() {
					log.Info("rbac error: %v", e.Error())
				}
			}
			return errors.Errorf("insufficient permissions to run all collectors")
		}
	}

	analyzeResults := collectResults.Analyze()

	failedStrictPreflights := []string{}
	for _, analyzeResult := range analyzeResults {
		if analyzeResult.IsFail && analyzeResult.Strict {
			failedStrictPreflights = append(failedStrictPreflights, analyzeResult.Title)
		}
	}
	if len(failedStrictPreflights) > 0 {
		return errors.Errorf("Strict preflights failed: %s", strings.Join(failedStrictPreflights, ", "))
	}

	return nil
}

func installBuilderApp(r *runners, log *logger.Logger, kubeConfig []byte, customer *types.Customer, release *types.AppRelease) error {
	kubeconfigFile, err := os.CreateTemp("", "kubeconfig")
	if err != nil {
		return errors.Wrap(err, "failed to create kubeConfigFile file")
	}
	defer func() {
		kubeconfigFile.Close()
		os.Remove(kubeconfigFile.Name())
	}()
	if _, err := kubeconfigFile.Write(kubeConfig); err != nil {
		return errors.Wrap(err, "write kubeConfig file")
	}
	if err := kubeconfigFile.Chmod(0644); err != nil {
		return errors.Wrap(err, "chmod kubeConfig file")
	}

	kubeconfigFlag := flag.String("kubeconfig", kubeconfigFile.Name(), "kubeconfig file")
	restKubeConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfigFlag)
	if err != nil {
		return errors.Wrap(err, "build config from flags")
	}

	authConfig := dockertypes.AuthConfig{
		Username: customer.Email,
		Password: customer.InstallationID,
		Auth:     base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", customer.Email, customer.InstallationID))),
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
	defer func() {
		credentialsFile.Close()
		os.Remove(credentialsFile.Name())
	}()
	if _, err := credentialsFile.Write([]byte(configJSON)); err != nil {
		return errors.Wrap(err, "failed to write credentials file")
	}

	ctx := context.Background()
	for _, chart := range release.Charts {
		dryRunRelease, err := installHelmChart(r, r.appSlug, chart.Name, release.Sequence, registryHostname, kubeconfigFile.Name(), credentialsFile.Name(), true)
		if err != nil {
			return errors.Wrap(err, "dry run release")
		}

		log.ActionWithSpinner("Running preflights")
		if err = runPreflights(ctx, r, log, restKubeConfig, dryRunRelease); err != nil {
			log.FinishSpinnerWithError()
			return errors.Wrap(err, "run preflights")
		}
		log.FinishSpinner()

		release, err := installHelmChart(r, r.appSlug, chart.Name, release.Sequence, registryHostname, kubeconfigFile.Name(), credentialsFile.Name(), false)
		if err != nil {
			return errors.Wrap(err, "install release")
		}

		fmt.Fprintf(r.w, "%s\n", release.Info.Notes)
	}

	return nil
}

func installHelmChart(r *runners, appSlug string, chartName string, releaseSequence int64, registryHostname string, kubeconfigFile string, credentialsFile string, dryRun bool) (*release.Release, error) {
	settings := cli.New()
	settings.KubeConfig = kubeconfigFile

	values, err := r.args.prepareClusterValueOpts.MergeValues(getter.All(settings))
	if err != nil {
		return nil, errors.Wrap(err, "failed to merge values")
	}

	registryClient, err := registry.NewClient(
		registry.ClientOptDebug(settings.Debug),
		registry.ClientOptWriter(io.Discard),
		registry.ClientOptCredentialsFile(credentialsFile),
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
	client.Namespace = r.args.prepareClusterNamespace
	client.Timeout = r.args.prepareClusterAppReadyTimeout
	client.Wait = true

	if dryRun {
		client.DryRun = true
		client.ClientOnly = true
	}

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

func installKotsApp(r *runners, log *logger.Logger, kubeConfig []byte, customer *types.Customer, release *types.AppRelease) error {
	var releaseYamls []releaseTypes.KotsSingleSpec
	if err := json.Unmarshal([]byte(release.Config), &releaseYamls); err != nil {
		return errors.Wrap(err, "failed to unmarshal release yamls")
	}

	kotsApp, err := kotsutil.GetKotsApplicationSpec(releaseYamls)
	if err != nil {
		return errors.Wrap(err, "failed to get KOTS application")
	}

	kotsCliVersion := ""
	if kotsApp != nil && kotsApp.Spec.TargetKotsVersion != "" {
		kotsCliVersion = kotsApp.Spec.TargetKotsVersion
	}

	kotsDir, err := os.MkdirTemp("", "kots")
	if err != nil {
		return errors.Wrap(err, "failed to create temp dir")
	}
	defer os.RemoveAll(kotsDir)
	kotCLI, err := installKotsCLI(r, kotsCliVersion, kotsDir)
	if err != nil {
		return errors.Wrap(err, "failed to install KOTS cli")
	}

	license, err := r.api.DownloadLicense(r.appType, r.appID, customer.ID)
	if err != nil {
		return errors.Wrapf(err, "failed to download license for customer %q", customer.Name)
	}

	licenseFile, err := os.CreateTemp(kotsDir, "license")
	if err != nil {
		return errors.Wrap(err, "failed to create license file")
	}
	defer func() {
		licenseFile.Close()
		os.Remove(licenseFile.Name())
	}()
	if _, err := licenseFile.Write(license); err != nil {
		return errors.Wrap(err, "failed to write license file")
	}

	kubeconfigFile, err := os.CreateTemp(kotsDir, "kubeconfig")
	if err != nil {
		return errors.Wrap(err, "failed to create kubeconfig file")
	}
	defer func() {
		kubeconfigFile.Close()
		os.Remove(kubeconfigFile.Name())
	}()
	if _, err := kubeconfigFile.Write(kubeConfig); err != nil {
		return errors.Wrap(err, "write kubeconfig file")
	}
	if err := kubeconfigFile.Chmod(0644); err != nil {
		return errors.Wrap(err, "chmod kubeconfig file")
	}

	cmd := exec.Command(kotCLI, "install",
		fmt.Sprintf("%s/%s", r.appSlug, "test-channel"),
		"--kubeconfig", kubeconfigFile.Name(),
		"--license-file", licenseFile.Name(),
		"--namespace", r.args.prepareClusterNamespace,
		"--wait-duration", r.args.prepareClusterAppReadyTimeout.String(),
		"--shared-password", r.args.prepareClusterKotsSharedPassword,
		"--app-version-label", fmt.Sprintf("release__%d", release.Sequence),
		"--no-port-forward",
		"--skip-preflights",
	)
	if r.args.prepareClusterKotsConfigValuesFile != "" {
		if _, err := os.Stat(r.args.prepareClusterKotsConfigValuesFile); os.IsNotExist(err) {
			return errors.Wrapf(err, "config values file %q does not exist", r.args.prepareClusterKotsConfigValuesFile)
		}

		cmd.Args = append(cmd.Args, "--config-values", r.args.prepareClusterKotsConfigValuesFile)
	}

	log.Verbose()
	log.Debug(cmd.String())
	cmd.Stdout = r.w
	cmd.Stderr = r.w
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "failed to install KOTS and deploy app")
	}

	return nil
}

func installKotsCLI(r *runners, version string, kotsDir string) (string, error) {
	kotsInstallURL := "https://kots.io/install"
	if version != "" {
		kotsInstallURL = fmt.Sprintf("%s/%s", kotsInstallURL, version)
	}

	resp, err := http.Get(kotsInstallURL)
	if err != nil {
		return "", errors.Wrap(err, "failed to download kots")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("failed to download kots: %s", resp.Status)
	}

	installScript, err := os.CreateTemp(kotsDir, "kots")
	if err != nil {
		return "", errors.Wrap(err, "failed to create temp file")
	}
	defer func() {
		installScript.Close()
		os.Remove(installScript.Name())
	}()
	if _, err := io.Copy(installScript, resp.Body); err != nil {
		return "", errors.Wrap(err, "failed to write KOTS binary")
	}
	if err := installScript.Chmod(0755); err != nil {
		return "", errors.Wrap(err, "chmod KOTS install script")
	}

	cmd := exec.Command(installScript.Name())
	cmd.Env = append(cmd.Env, fmt.Sprintf("REPL_INSTALL_PATH=%s", kotsDir))
	cmd.Stdout = r.w
	cmd.Stderr = r.w
	if err := cmd.Run(); err != nil {
		return "", errors.Wrap(err, "failed to run KOTS cli install")
	}

	kotsInstallPath := filepath.Join(kotsDir, "kubectl-kots")
	return kotsInstallPath, nil
}
