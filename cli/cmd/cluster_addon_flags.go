package cmd

import (
	"time"

	"github.com/spf13/cobra"
)

type clusterAddonArgs struct {
	outputFormat string
}

type clusterAddonCreateArgs struct {
	clusterID    string
	waitDuration time.Duration
	dryRun       bool
}

func clusterAddonFlags(cmd *cobra.Command, args *clusterAddonArgs) error {
	cmd.Flags().StringVar(&args.outputFormat, "output", "table", "The output format to use. One of: json|table|wide (default: table)")
	return nil
}

func clusterAddonCreateFlags(cmd *cobra.Command, args *clusterAddonCreateArgs) error {
	cmd.Flags().DurationVar(&args.waitDuration, "wait", 0, "Wait duration for addon to be ready (leave empty to not wait)")
	cmd.Flags().BoolVar(&args.dryRun, "dry-run", false, "Dry run")
	return nil
}
