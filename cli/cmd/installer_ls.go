package cmd

import (
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
	cmd.RunE = r.installerList
}

func (r *runners) installerList(_ *cobra.Command, _ []string) error {
	installers, err := r.api.ListInstallers(r.appID, r.appType)
	if err != nil {
		return err
	}

	return print.Installers(r.w, installers)
}
