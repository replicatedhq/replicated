package cmd

import (
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitEnterpriseInstallerUpdate(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:          "update",
		SilenceUsage: true,
		Short:        "update an existing installer",
		Long: `Update an existing installer.

  Example:
  replicated enteprise installer update --id MyInstallerID --yaml-file myinstaller.yaml`,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.enterpriseInstallerUpdateID, "id", "", "The id of the installer to be updated")
	cmd.Flags().StringVar(&r.args.enterpriseInstallerUpdateFile, "yaml-file", "", "The filename containing the installer yaml")

	cmd.RunE = r.enterpriseInstallerUpdate
}

func (r *runners) enterpriseInstallerUpdate(cmd *cobra.Command, args []string) error {
	b, err := ioutil.ReadFile(r.args.enterpriseInstallerUpdateFile)
	if err != nil {
		return errors.Wrap(err, "failed to read file")
	}

	installer, err := r.enterpriseClient.UpdateInstaller(r.args.enterpriseInstallerUpdateID, string(b))
	if err != nil {
		return err
	}

	print.EnterpriseInstaller(r.w, installer)
	return nil
}
