package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitEnterpriseAuthInit(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "init",
		Short:        "initialize authentication",
		Long:         `Create a keypair to begin authentication`,
		RunE:         r.enterpriseAuthInit,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) enterpriseAuthInit(cmd *cobra.Command, args []string) error {
	err := r.enterpriseClient.AuthInit()
	if err != nil {
		return err
	}

	return nil
}
