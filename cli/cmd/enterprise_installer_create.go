package cmd

import (
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitEnterpriseInstallerCreate(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:          "create",
		SilenceUsage: true,
		Short:        "Create a new custom installer",
		Long: `Create a new custom installer.

  Example:
  replicated enteprise installer create --yaml-file myinstaller.yaml`,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.enterpriseInstallerCreateFile, "yaml-file", "", "The filename containing the installer yaml")
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	cmd.RunE = r.enterpriseInstallerCreate
}

func (r *runners) enterpriseInstallerCreate(cmd *cobra.Command, args []string) error {
	b, err := ioutil.ReadFile(r.args.enterpriseInstallerCreateFile)
	if err != nil {
		return errors.Wrap(err, "failed to read file")
	}

	installer, err := r.enterpriseClient.CreateInstaller(string(b))
	if err != nil {
		return err
	}

	print.EnterpriseInstaller(r.outputFormat, r.w, installer)
	return nil
}
