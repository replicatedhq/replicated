package cmd

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitReleaseInspect(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "inspect RELEASE_SEQUENCE",
		Short: "Long: information about a release",
		Long: `Show information about the specified application release.

This command displays detailed information about a specific release of an application.

The output can be customized using the --output flag to display results in
either table or JSON format.
		`,
		Example: `  # Display information about a release
  replicated release inspect 123

  # Display information about a release in JSON format
  replicated release inspect 123 --output json`,
		Args: cobra.ExactArgs(1),
	}
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")
	parent.AddCommand(cmd)

	cmd.RunE = r.releaseInspect
}

func (r *runners) releaseInspect(_ *cobra.Command, args []string) error {
	if !r.hasApp() {
		return errors.New("no app specified")
	}

	seq, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("Failed to parse sequence argument %s", args[0])
	}

	release, err := r.api.GetRelease(r.appID, r.appType, seq)
	if err != nil {
		if err == platformclient.ErrNotFound {
			return fmt.Errorf("No such release %d", seq)
		}
		return err
	}

	return print.Release(r.outputFormat, r.w, release)
}
