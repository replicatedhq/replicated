package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (r *runners) InitEnterpriseInstallerRM(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:          "rm",
		SilenceUsage: true,
		Short:        "Remove an installer",
		Long: `Remove an installer.

  Example:
  replicated enteprise installer rm --id MyInstallerID`,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.enterpriseInstallerRmId, "id", "", "The id of the installer to remove")

	cmd.RunE = r.enterpriseInstallerRemove
}

func (r *runners) enterpriseInstallerRemove(cmd *cobra.Command, args []string) error {
	err := r.enterpriseClient.RemoveInstaller(r.args.enterpriseInstallerRmId)
	if err != nil {
		return err
	}

	fmt.Fprintf(r.w, "Installer %s successfully removed\n", r.args.enterpriseInstallerRmId)
	r.w.Flush()

	return nil
}
