package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (r *runners) InitEnterpriseInstallerAssign(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:          "assign",
		SilenceUsage: true,
		Short:        "Assigns an installer to a channel",
		Long: `Assigns an installer to a channel.

  Example:
  replicated enteprise installer assign --installer-id 123 --channel-id abc`,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.enterpriseInstallerAssignInstallerID, "installer-id", "", "The id of the installer to assign")
	cmd.Flags().StringVar(&r.args.enterpriseInstallerAssignChannelID, "channel-id", "", "The id of channel")

	cmd.RunE = r.enterpriseInstallerAssign
}

func (r *runners) enterpriseInstallerAssign(cmd *cobra.Command, args []string) error {
	err := r.enterpriseClient.AssignInstaller(r.args.enterpriseInstallerAssignInstallerID, r.args.enterpriseInstallerAssignChannelID)
	if err != nil {
		return err
	}

	fmt.Fprintf(r.w, "Installer successfully assigned\n")
	r.w.Flush()

	return nil
}
