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
		Use:     "ls",
		Short:   "List enterprise portal users",
		Long:    `List all users associated with the enterprise portal for an application.`,
		Example: "",
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.enterprisePortalUserLs(cmd, r.appID, opts, outputFormat)
		},
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().BoolVar(&opts.includeInvites, "include-invites", false, "Include invites")
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
