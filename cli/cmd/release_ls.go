package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) IniReleaseList(parent *cobra.Command) {
	var limit, page, pageSize int

	cmd := &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "List all of an app's releases",
		Long:    "List all of an app's releases, newest first",
		Example: `# List all releases
replicated release ls

# List the 500 most recent releases
replicated release ls --limit 500

# List the 500 most recent releases, fetching 100 per request
replicated release ls --limit 500 --page-size 100

# Fetch a single page (second page of 50)
replicated release ls --page 1 --page-size 50`,
		SilenceUsage: true,
	}

	parent.AddCommand(cmd)
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of most recent releases to list, fetched across as many pages as needed (0 means no limit)")
	cmd.Flags().IntVar(&page, "page", 0, "The page to fetch, starting at 0. Use with --page-size (KOTS apps only).")
	cmd.Flags().IntVar(&pageSize, "page-size", 0, "The number of releases per page. Without --limit, only one page is fetched (KOTS apps only).")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return r.releaseList(cmd, limit, page, pageSize)
	}
}

func (r *runners) releaseList(cmd *cobra.Command, limit int, page int, pageSize int) error {
	if !r.hasApp() {
		return errors.New("no app specified")
	}
	if limit < 0 {
		return errors.New("--limit must be 0 or greater")
	}
	if cmd.Flags().Changed("page") && limit != 0 {
		return errors.New("--page cannot be used with --limit")
	}

	var releases []types.ReleaseInfo
	var err error
	if (cmd.Flags().Changed("page") || cmd.Flags().Changed("page-size")) && limit == 0 {
		releases, err = r.api.ListReleasesPaged(r.appID, r.appType, page, pageSize)
	} else {
		releases, err = r.api.ListReleases(r.appID, r.appType, limit, pageSize)
	}
	if err != nil {
		return err
	}

	return print.Releases(r.outputFormat, r.w, releases)
}
