package cmd

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"

	analyzer "github.com/replicatedhq/troubleshoot/pkg/analyze"
	"github.com/replicatedhq/troubleshoot/pkg/apis/troubleshoot/v1beta2"
	"github.com/replicatedhq/troubleshoot/pkg/loader"
	"github.com/replicatedhq/troubleshoot/pkg/supportbundle"
)

type analysisOutput struct {
	Analysis    []*analyzer.AnalyzeResult
	ArchivePath string
}

func (r *runners) InitClusterValidate(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate ID",
		Short: "Generate a support bundle on test cluster",
		Long:  "Generate a support bundle on test cluster.",
		RunE:  r.validateCluster,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.validationClusterName, "name", "", "Name of the cluster to generate a support bundle for.")
	cmd.Flags().Int64Var(&r.args.validationReleaseSequence, "release-sequence", -1, "Release sequence to report results for.")
	cmd.Flags().StringVar(&r.args.validationAppVersion, "app-version", "", "App version to report results for.")
	cmd.Flags().StringVar(&r.args.supportBundleFile, "support-bundle-file", "", "Support bundle file to use.")

	// Which support bundle spec to use: --support-bundle-spec
	// 1. Option: "file" is provided, use the local file (--support-bundle-file required)
	// 2. Option: "cluster" flag is provided, search for the spec in the cmx cluster
	// 3. Option: "release" flag is provided, search for the spec in the release (vendor portal) (--release-sequence required)
	cmd.Flags().StringVar(&r.args.supportBundleSpec, "support-bundle-spec", "cluster", "Support bundle spec to use. Options: file, cluster, release")

	// Report results compatibility (--release-sequence or --app-version required)
	cmd.Flags().BoolVar(&r.args.validationReportCompatibility, "report-compatibility", false, "Report results compatibility")

	return cmd
}

func (r *runners) validateCluster(_ *cobra.Command, args []string) error {
	if len(args) == 0 && r.args.validationClusterName == "" {
		return errors.New("One of ID or --name required")
	} else if len(args) > 0 && r.args.validationClusterName != "" {
		return errors.New("cannot specify ID and --name flag")
	}

	if r.args.validationReportCompatibility && (r.args.validationReleaseSequence == -1 && r.args.validationAppVersion == "") {
		return errors.New("--report-compatibility requires --release-sequence or --app-version")
	}
	if r.args.validationReleaseSequence != -1 && r.args.validationAppVersion != "" {
		return errors.New("only one of --release-sequence or --app-version is allowed")
	}
	if r.args.supportBundleSpec == "file" && r.args.supportBundleFile == "" {
		return errors.New("--support-bundle-file required when --support-bundle-spec=file")
	}
	if r.args.supportBundleSpec == "release" && r.args.validationReleaseSequence == -1 {
		return errors.New("--release-sequence required when --support-bundle-spec=release")
	}

	// Get kubeconfig
	var clusterID string
	if len(args) > 0 {
		clusterID = args[0]
	} else {
		// Get cluster ID by name
		clusters, err := r.kotsAPI.ListClusters(false, nil, nil)
		if err != nil {
			return errors.Wrap(err, "list clusters")
		}
		for _, cluster := range clusters {
			if cluster.Name == r.args.validationClusterName {
				clusterID = cluster.ID
				break
			}
		}
	}
	kubeconfig, err := r.kotsAPI.GetClusterKubeconfig(clusterID)
	if err != nil {
		return errors.Wrap(err, "failed to get cluster kubeconfig")
	}
	_, err = r.kotsAPI.GetCluster(clusterID)
	if err != nil {
		return errors.Wrap(err, "failed to get cluster")
	}
	restKubeConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return errors.Wrap(err, "failed to create rest kubeconfig")
	}

	// Get support bundle upload URL
	uploadResponse, err := r.kotsAPI.GetSupportBundleUploadURL()
	if err != nil {
		return errors.Wrap(err, "failed to get support bundle upload URL")
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

		// Inject afterCollection hook to upload support bundle
		kinds.SupportBundlesV1Beta2[0].Spec.AfterCollection = []*v1beta2.AfterCollection{
			{
				UploadResultsTo: &v1beta2.ResultRequest{
					URI:    uploadResponse.URL,
					Method: "PUT",
				},
			},
		}

		response, err := supportbundle.CollectSupportBundleFromSpec(&kinds.SupportBundlesV1Beta2[0].Spec, &additionalRedactors, createOpts)
		if err != nil {
			return errors.Wrap(err, "failed to run collect and analyze process")
		}

		fmt.Printf("Support bundle uploaded %v\n", response.FileUploaded)

		close(progressChan) // this removes the spinner in interactive mode
		isProgressChanClosed = true

		if len(response.AnalyzerResults) > 0 {
			nonInteractiveOutput.Analysis = response.AnalyzerResults
		}

		// Mark bundle as uploaded
		err = r.kotsAPI.SupportBundleUploaded(uploadResponse.BundleID)
		if err != nil {
			return errors.Wrap(err, "failed to mark support bundle as uploaded")
		}
	}

	// Report compatibility results
	passed := true
	for _, analysis := range nonInteractiveOutput.Analysis {
		if analysis.IsFail || analysis.IsWarn {
			passed = false
			fmt.Printf("Analysis %s failed\n", analysis.Title)
		}
	}

	if passed {
		fmt.Println("All analyses passed")
	}

	if r.args.validationReportCompatibility {
		err := r.kotsAPI.ReportClusterCompatibility(clusterID, uploadResponse.BundleID, r.appID, r.args.validationReleaseSequence, r.args.validationAppVersion)
		if err != nil {
			return err
		}
	}

	return nil
}
