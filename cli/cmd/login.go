package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/credentials"
	"github.com/spf13/cobra"
)

func (r *runners) InitLoginCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "login",
		Short:        "Log in to Replicated",
		Long:         `This command will open your browser to ask you authentication details and create / retrieve an API token for the CLI to use.`,
		SilenceUsage: true,
		RunE:         r.login,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&r.args.loginEndpoint, "endpoint", "https://vendor.replicated.com", "The endpoint to use for the login process. Defaults to https://vendor.replicated.com")

	return cmd
}

func (r *runners) login(_ *cobra.Command, _ []string) error {
	endpoint := r.args.loginEndpoint

	// if we are already logged in, just return
	currentCredentials, err := credentials.GetCurrentCredentials()
	if err == nil {
		if currentCredentials.IsEnv {
			return errors.New("REPLICATED_API_TOKEN is set. Please unset it to login with a different account.")
		}
		return errors.New("There are already credentials on this machine. Please run `replicated logout` to remove them.")
	}
	if err != credentials.ErrCredentialsNotFound {
		return err
	}

	// open a browser to the login page
	if err := credentials.Fetch(endpoint); err != nil {
		if err == credentials.ErrNoBrowser {
			return fmt.Errorf("No browser could be found. Please visit %s to login and create an API token.", endpoint)
		}
		return err
	}

	return nil
}
