package cmd

import (
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

// lsCmd represents the ls command
var releaseLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "list all of an app's releases",
	Long: `List all of an app's releases
	replicatedReleaseLs
`,
}

func init() {
	releaseCmd.AddCommand(releaseLsCmd)
}

func (r *runners) releaseList(cmd *cobra.Command, args []string) error {
	releases, err := r.api.ListReleases(r.appID)
	if err != nil {
		return err
	}

	return print.Releases(r.w, releases)
}
