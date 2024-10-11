package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type listEnterprisePortalUsersOpts struct {
	includeInvites bool
}

func (r *runners) InitEnterprisePortalUserLsCmd(parent *cobra.Command) *cobra.Command {
	opts := listEnterprisePortalUsersOpts{}
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List Enterprise Portal users",
		Long: `List all users associated with the Enterprise Portal for an application.

This command retrieves and displays information about users who have access to
the Enterprise Portal. By default, it shows only active users, but you can include
pending invitations as well.

Use the --include-invites flag to also list users who have been invited but
haven't yet accepted their invitations.`,
		Example: `  # List all active Enterprise Portal users
  replicated enterprise-portal user ls

  # List all Enterprise Portal users, including pending invitations
  replicated enterprise-portal user ls --include-invites

  # List Enterprise Portal users for a specific application
  replicated enterprise-portal user ls --app myapp

  # List Enterprise Portal users and output in JSON format
  replicated enterprise-portal user ls --output json

  # List all users, including invites, for a specific app in table format
  replicated enterprise-portal user ls --app myapp --include-invites --output table`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.enterprisePortalUserLs(cmd, r.appID, opts, outputFormat)
		},
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().BoolVar(&opts.includeInvites, "include-invites", false, "Include pending invitations in the list")
	cmd.Flags().StringVar(&outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	return cmd
}

func (r *runners) enterprisePortalUserLs(cmd *cobra.Command, appID string, opts listEnterprisePortalUsersOpts, outputFormat string) error {
	users, err := r.kotsAPI.ListEnterprisePortalUsers(appID, opts.includeInvites)
	if err != nil {
		return err
	}

	if len(users) == 0 {
		return nil
	}

	for _, user := range users {
		fmt.Printf("%s\n", user.Email)
	}
	return nil
}
