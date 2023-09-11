package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/credentials"
	"github.com/spf13/cobra"
)

func (r *runners) InitLogoutCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "logout",
		Short:        "Logout from Replicated",
		Long:         `This command will remove any stored credentials from the CLI.`,
		SilenceUsage: true,
		RunE:         r.logout,
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) logout(_ *cobra.Command, _ []string) error {
	// if we are already logged in, just return
	currentCredentials, err := credentials.GetCurrentCredentials()

	if err != nil {
		return err
	}
	if currentCredentials.IsEnv {
		return errors.New("REPLICATED_API_TOKEN is set. Please unset it to logout.")
	}

	if err = credentials.RemoveCurrentCredentials(); err != nil {
		return err
	}

	return nil
}
