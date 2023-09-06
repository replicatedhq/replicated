package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitReleaseCompatibility(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compatibility SEQUENCE",
		Short: "Report release compatibility",
		Long:  "Report release compatibility for a kubernetes distribution and version",
		RunE:  r.reportReleaseCompatibility,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.compatibilityKubernetesDistribution, "distribution", "", "Kubernetes distribution of the cluster to report on.")
	cmd.Flags().StringVar(&r.args.compatibilityKubernetesVersion, "version", "", "Kubernetes version of the cluster to report on (format is distribution dependent)")
	cmd.Flags().BoolVar(&r.args.compatibilitySuccess, "success", false, "If set, the compatibility will be reported as a success.")
	cmd.Flags().BoolVar(&r.args.compatibilityFailure, "failure", false, "If set, the compatibility will be reported as a failure.")
	cmd.Flags().StringVar(&r.args.compatibilityNotes, "notes", "", "Additional notes to report.")

	cmd.MarkFlagRequired("distribution")
	cmd.MarkFlagRequired("version")

	return cmd
}

func (r *runners) reportReleaseCompatibility(_ *cobra.Command, args []string) error {
	// parse sequence positional arguments
	if len(args) != 1 {
		return errors.New("release sequence is required")
	}
	seq, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse sequence argument %s", args[0])
	}

	//Check if one of success or failure is set
	if r.args.compatibilitySuccess && r.args.compatibilityFailure {
		return fmt.Errorf("cannot set both success and failure")
	}
	if !r.args.compatibilitySuccess && !r.args.compatibilityFailure {
		return fmt.Errorf("must set either success or failure")
	}

	success := false
	if r.args.compatibilitySuccess {
		success = true
	}

	ve, err := r.kotsAPI.ReportReleaseCompatibility(r.appID, seq, r.args.compatibilityKubernetesDistribution, r.args.compatibilityKubernetesVersion, success, r.args.compatibilityNotes)
	if ve != nil && len(ve.Errors) > 0 {
		if len(ve.SupportedDistributions) > 0 {
			print.ClusterVersions("table", r.w, ve.SupportedDistributions)
		}
		return fmt.Errorf("%s", errors.New(strings.Join(ve.Errors, ",")))
	}
	if err != nil {
		return err
	}

	return nil
}
