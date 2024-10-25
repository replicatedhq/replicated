package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitInstallerList(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List an app's Kubernetes Installers",
		Long:  "List an app's https://kurl.sh Kubernetes Installers",
	}

	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	cmd.RunE = r.installerList
}

func (r *runners) installerList(_ *cobra.Command, _ []string) error {
	if !r.hasApp() {
		return errors.New("no app specified")
	}

	installers, err := r.api.ListInstallers(r.appID, r.appType)
	if err != nil {
		return err
	}

	return print.Installers(r.outputFormat, r.w, installers)
}
