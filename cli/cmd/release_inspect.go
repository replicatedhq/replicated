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
		Use:   "inspect SEQUENCE",
		Short: "Print the YAML config for a release",
		Long:  "Print the YAML config for a release",
	}
	cmd.Hidden = true // Not supported in KOTS
	parent.AddCommand(cmd)
	cmd.RunE = r.releaseInspect
}

func (r *runners) releaseInspect(_ *cobra.Command, args []string) error {
	if !r.hasApp() {
		return errors.New("no app specified")
	}

	if len(args) != 1 {
		return errors.New("release sequence is required")
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

	return print.Release(r.w, release)
}
