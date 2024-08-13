package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterCompatibility(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compatibility",
		Short: "Check cluster compatibility",
		Long:  "Check cluster compatibility.",
		RunE:  r.checkCompatibility,
	}
	parent.AddCommand(cmd)

	cmd.Flags().Int64Var(&r.args.compatibilityReleaseSequence, "release-sequence", -1, "Release sequence to check results for.")
	cmd.Flags().StringVar(&r.args.compatibilityAppVersion, "app-version", "", "App version to check results for.")

	return cmd
}

func (r *runners) checkCompatibility(_ *cobra.Command, args []string) error {
	if r.args.compatibilityReleaseSequence == -1 && r.args.compatibilityAppVersion == "" {
		return errors.New("One of --release-sequence or --app-version is required")
	}
	if r.args.compatibilityReleaseSequence != -1 && r.args.compatibilityAppVersion != "" {
		return errors.New("Only one of --release-sequence or --app-version is allowed")
	}

	// GET /cluster/validations
	validateVersions, err := r.kotsAPI.ListClusterValidations(r.appID, r.args.compatibilityReleaseSequence, r.args.compatibilityAppVersion)
	if err != nil {
		return errors.Wrap(err, "failed to list cluster validations")
	}
	// For each validation
	// get the cluster details based on the cluster ID
	// get the support bundle details based on the support bundle ID
	for _, validateVersion := range validateVersions {
		cl, err := r.kotsAPI.GetCluster(validateVersion.ClusterGUID)
		if err != nil {
			return errors.Wrap(err, "failed to get cluster")
		}

		sb, err := r.kotsAPI.GetSupportBundle(validateVersion.SupportBundleID)
		if err != nil {
			return errors.Wrap(err, "failed to get support bundle")
		}

		fmt.Printf("Compatibility for %s - %s\n", cl.KubernetesDistribution, cl.KubernetesVersion)
		if sb.Status == kotsclient.SupportBundleStatusPending {
			fmt.Println("Insights: Pending")
		} else {
			fmt.Println("Insights:")
			for _, insight := range sb.Insights {
				fmt.Printf("  %s: %s\n", insight.Level, insight.Detail)
			}
		}

	}

	// Print the results of the compatibility check

	return nil
}
