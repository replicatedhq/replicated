package cmd

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitRegistryList(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "ls [NAME]",
		Aliases:      []string{"list"},
		Short:        "list registries",
		Long:         `list registries, or a single registry by name`,
		RunE:         r.listRegistries,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "The output format to use. One of: json|table")

	return cmd
}

func (r *runners) listRegistries(_ *cobra.Command, args []string) error {
	kotsRegistries, err := r.kotsAPI.ListRegistries()
	if err != nil {
		return errors.Wrap(err, "list registries")
	}

	if len(args) == 0 {
		return print.Registries(r.outputFormat, r.w, kotsRegistries)
	}

	registrySearch := args[0]
	var resultRegistries []types.Registry
	for _, registry := range kotsRegistries {
		if strings.Contains(registry.Endpoint, registrySearch) ||
			strings.Contains(registry.Slug, registrySearch) {
			resultRegistries = append(resultRegistries, registry)
		}
	}

	return print.Registries(r.outputFormat, r.w, resultRegistries)
}
