package cmd

import (
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitEnterpriseInstallerLS(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "ls",
		Short:        "lists enterprise installers",
		Long:         `lists all installers that have been created`,
		RunE:         r.enterpriseInstallerList,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) enterpriseInstallerList(cmd *cobra.Command, args []string) error {
	installers, err := r.enterpriseClient.ListInstallers()
	if err != nil {
		return err
	}

	print.EnterpriseInstallers(r.w, installers)
	return nil
}
