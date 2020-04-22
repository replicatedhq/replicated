package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitEnterpriseAuthInit(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:          "init",
		Short:        "initialize authentication",
		Long:         `Create a keypair to begin authentication`,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.enterpriseAuthInitCreateOrg, "create-org", "", "If this flag is provided, a new organization will be created with the specified name. If not, the auth request will have to be approved by Replicated or your already authenticated organization")

	cmd.RunE = r.enterpriseAuthInit
}

func (r *runners) enterpriseAuthInit(cmd *cobra.Command, args []string) error {
	err := r.enterpriseClient.AuthInit(r.args.enterpriseAuthInitCreateOrg)
	if err != nil {
		return err
	}
	return nil
}
