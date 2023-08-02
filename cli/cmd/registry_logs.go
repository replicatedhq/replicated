package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitRegistryLogs(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "logs [NAME]",
		Short:        "show registry logs",
		Long:         `show registry logs for a single registry`,
		RunE:         r.logsRegistry,
		Hidden:       true,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) logsRegistry(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("missing endpoint")
	}
	hostname := args[0]

	logs, err := r.kotsAPI.LogsRegistry(hostname)
	if err != nil {
		return errors.Wrap(err, "registry logs")
	}

	return print.RegistryLogs(r.w, logs)
}
