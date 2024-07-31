package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/replicatedhq/replicated/cli/print"
	analyzer "github.com/replicatedhq/troubleshoot/pkg/analyze"
	"github.com/replicatedhq/troubleshoot/pkg/apis/troubleshoot/v1beta2"
	"github.com/replicatedhq/troubleshoot/pkg/loader"
	"github.com/replicatedhq/troubleshoot/pkg/supportbundle"
)

type analysisOutput struct {
	Analysis    []*analyzer.AnalyzeResult
	ArchivePath string
}

func (r *runners) InitClusterSupportBundle(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "support ID",
		Short: "Generate a support bundle on test cluster",
		Long:  "Generate a support bundle on test cluster.",
		RunE:  r.generateSupportBundle,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.supportBundleClusterName, "name", "", "Name of the cluster to generate a support bundle for.")
	cmd.Flags().Int64Var(&r.args.supportBundleReleaseSequence, "release-sequence", -1, "Release sequence of the support bundle to generate.")
	cmd.Flags().StringVar(&r.args.supportBundleFile, "support-bundle-file", "", "Support bundle file to use.")

	// Which support bundle spec to use: --support-bundle-spec
	// 1. Option: "file" is provided, use the local file (--support-bundle-file required)
	// 2. Option: "cluster" flag is provided, search for the spec in the cmx cluster
	// 3. Option: "release" flag is provided, search for the spec in the release (vendor portal) (--release-sequence required)
	cmd.Flags().StringVar(&r.args.supportBundleSpec, "support-bundle-spec", "cluster", "Support bundle spec to use. Options: file, cluster, release")

	// Report results compatibility (--release-sequence required)
	cmd.Flags().BoolVar(&r.args.supportBundleReportCompatibility, "report-compatibility", false, "Report results compatibility")

	return cmd
}

func (r *runners) generateSupportBundle(_ *cobra.Command, args []string) error {
	if len(args) == 0 && r.args.supportBundleClusterName == "" {
		return errors.New("One of ID or --name required")
	} else if len(args) > 0 && r.args.supportBundleClusterName != "" {
		return errors.New("cannot specify ID and --name flag")
	}

	if r.args.supportBundleReportCompatibility && r.args.supportBundleReleaseSequence == -1 {
		return errors.New("--report-compatibility requires --release-sequence")
	}
	if r.args.supportBundleSpec == "file" && r.args.supportBundleFile == "" {
		return errors.New("--support-bundle-file required when --support-bundle-spec=file")
	}
	if r.args.supportBundleSpec == "release" && r.args.supportBundleReleaseSequence == -1 {
		return errors.New("--release-sequence required when --support-bundle-spec=release")
	}

	// Get kubeconfig
	var clusterID string
	var kubernetesDistribution string
	var kubernetesVersion string
	if len(args) > 0 {
		clusterID = args[0]
	} else {
		// Get cluster ID by name
		clusters, err := r.kotsAPI.ListClusters(false, nil, nil)
		if err != nil {
			return errors.Wrap(err, "list clusters")
		}
		for _, cluster := range clusters {
			if cluster.Name == r.args.supportBundleClusterName {
				clusterID = cluster.ID
				break
			}
		}
	}
	kubeconfig, err := r.kotsAPI.GetClusterKubeconfig(clusterID)
	if err != nil {
		return errors.Wrap(err, "failed to get cluster kubeconfig")
	}
	cluster, err := r.kotsAPI.GetCluster(clusterID)
	if err != nil {
		return errors.Wrap(err, "failed to get cluster")
	}
	kubernetesDistribution = cluster.KubernetesDistribution
	kubernetesVersion = cluster.KubernetesVersion
	restKubeConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return errors.Wrap(err, "failed to create rest kubeconfig")
	}

	// Get support bundle spec
	nonInteractiveOutput := analysisOutput{}
	if r.args.supportBundleSpec == "file" {
		// Get support bundle spec from cluster
		collectorCB := func(c chan interface{}, msg string) { c <- msg }
		var wg sync.WaitGroup
		progressChan := make(chan interface{})
		isProgressChanClosed := false
		defer func() {
			if !isProgressChanClosed {
				close(progressChan)
			}
			wg.Wait()
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			for msg := range progressChan {
				fmt.Printf("Collecting support bundle: %v\n", msg)
			}
		}()
		createOpts := supportbundle.SupportBundleCreateOpts{
			CollectorProgressCallback: collectorCB,
			CollectWithoutPermissions: true,
			KubernetesRestConfig:      restKubeConfig,
			Namespace:                 "default", // hard coded for now. This should be configurable
			ProgressChan:              progressChan,
			Redact:                    true,
			FromCLI:                   true,
		}

		rawSpecs := []string{}
		if _, err := os.Stat(r.args.supportBundleFile); err == nil {
			b, err := os.ReadFile(r.args.supportBundleFile)
			if err != nil {
				return errors.Wrap(err, "failed to read support bundle spec file")
			}

			rawSpecs = append(rawSpecs, string(b))
		}
		ctx := context.Background()
		kinds, err := loader.LoadSpecs(ctx, loader.LoadOptions{
			RawSpecs: rawSpecs,
		})
		if err != nil {
			return err
		}

		if len(kinds.SupportBundlesV1Beta2) != 1 {
			return errors.New("expected exactly one support bundle spec")
		}
		var additionalRedactors v1beta2.Redactor
		if len(kinds.RedactorsV1Beta2) == 1 {
			additionalRedactors = kinds.RedactorsV1Beta2[0]
		}
		response, err := supportbundle.CollectSupportBundleFromSpec(&kinds.SupportBundlesV1Beta2[0].Spec, &additionalRedactors, createOpts)
		if err != nil {
			return errors.Wrap(err, "failed to run collect and analyze process")
		}

		close(progressChan) // this removes the spinner in interactive mode
		isProgressChanClosed = true

		if len(response.AnalyzerResults) > 0 {
			nonInteractiveOutput.Analysis = response.AnalyzerResults
		}

	}
	// Report compatibility results
	passed := true
	notes := []string{"Support bundle generated from the cluster"}
	for _, analysis := range nonInteractiveOutput.Analysis {
		if analysis.IsFail || analysis.IsWarn {
			passed = false
			notes = append(notes, fmt.Sprintf("Analysis %s failed", analysis.Title))
			fmt.Printf("Analysis %s failed\n", analysis.Title)
		}
	}

	if passed {
		notes = append(notes, "All analyses passed")
		fmt.Println("All analyses passed")
	}

	if r.args.supportBundleReportCompatibility {
		ve, err := r.kotsAPI.ReportReleaseCompatibility(r.appID, r.args.supportBundleReleaseSequence, kubernetesDistribution, kubernetesVersion, passed, strings.Join(notes, "\n"))
		if ve != nil && len(ve.Errors) > 0 {
			if len(ve.SupportedDistributions) > 0 {
				print.ClusterVersions("table", r.w, ve.SupportedDistributions)
			}
			return fmt.Errorf("%s", errors.New(strings.Join(ve.Errors, ",")))
		}
		if err != nil {
			return err
		}
	}

	return nil
}
