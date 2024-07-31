package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterSupportBundle(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "support ID [ID …]",
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
	_, err := r.kotsAPI.GetClusterKubeconfig(clusterID)
	if err != nil {
		return errors.Wrap(err, "failed to get cluster kubeconfig")
	}

	// Get support bundle spec
	if r.args.supportBundleSpec == "cluster" {
		// Get support bundle spec from cluster

	}
	// Generate support bundle + Analyze
	// Report compatibility results

	return nil

}
