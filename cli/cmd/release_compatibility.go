package cmd

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (r *runners) InitReleaseCompatibility(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compatibility SEQUENCE",
		Short: "report release compatibility",
		Long:  "report release compatibility for a kubernetes distribution and version",
		RunE:  r.reportReleaseCompatibility,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.compatibilityKubernetesDistribution, "distribution", "kind", "Kubernetes distribution of the cluster to report on.")
	cmd.Flags().StringVar(&r.args.compatibilityKubernetesVersion, "version", "v1.25.3", "Kubernetes version of the cluster to report on (format is distribution dependent)")
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
		return fmt.Errorf("Failed to parse sequence argument %s", args[0])
	}

	//Check if one of success or failure is set
	if r.args.compatibilitySuccess && r.args.compatibilityFailure {
		return fmt.Errorf("cannot set both success and failure")
	}
	if !r.args.compatibilitySuccess && !r.args.compatibilityFailure {
		return fmt.Errorf("must set either success or failure")
	}

	err = r.kotsAPI.ReportReleaseCompatibility(r.appID, seq)
	if err != nil {
		return fmt.Errorf("report release compatibility")
	}

	return nil
}
