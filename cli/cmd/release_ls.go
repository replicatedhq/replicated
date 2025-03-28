package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) IniReleaseList(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "List all of an app's releases",
		Long:    "List all of an app's releases",
	}

	parent.AddCommand(cmd)
	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "The output format to use. One of: json|table")

	cmd.RunE = r.releaseList
}

func (r *runners) releaseList(cmd *cobra.Command, args []string) error {
	if !r.hasApp() {
		return errors.New("no app specified")
	}

	releases, err := r.api.ListReleases(r.appID, r.appType)
	if err != nil {
		return err
	}

	return print.Releases(r.outputFormat, r.w, releases)
}
