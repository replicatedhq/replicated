package cmd

import (
	"errors"
	"fmt"

	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitCollectorInspect(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "inspect SPEC_ID",
		Short: "Print the YAML config for a collector",
		Long:  "Print the YAML config for a collector",
	}

	parent.AddCommand(cmd)
	cmd.RunE = r.collectorInspect
}

func (r *runners) collectorInspect(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("collector ID is required")
	}
	id := args[0]

	collector, err := r.platformAPI.GetCollector(r.appID, id)
	if err != nil {
		if err == platformclient.ErrNotFound {
			return fmt.Errorf("No such collector %d", id)
		}
		return err
	}

	return print.Collector(r.w, collector)
}
