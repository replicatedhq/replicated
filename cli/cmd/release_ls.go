package cmd

import (
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) IniReleaseList(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List all of an app's releases",
		Long:  "List all of an app's releases",
	}

	parent.AddCommand(cmd)
	cmd.RunE = r.releaseList
}

func (r *runners) releaseList(cmd *cobra.Command, args []string) error {
	releases, err := r.api.ListReleases(r.appID, r.appType)
	if err != nil {
		return err
	}

	return print.Releases(r.stdoutWriter, releases)
}
