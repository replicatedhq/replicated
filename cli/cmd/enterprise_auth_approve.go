package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitEnterpriseAuthApprove(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:          "approve",
		SilenceUsage: true,
		Short:        "approve an auth key request",
		Long: `approve an auth key request given a fingerprint
		
  Example:
  replicated enteprise auth approve --fingerprint <fingerprint>`,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.enterpriseAuthApproveFingerprint, "fingerprint", "", "The fingerprint provided on auth init")

	cmd.RunE = r.enterpriseAuthApprove
}

func (r *runners) enterpriseAuthApprove(cmd *cobra.Command, args []string) error {
	err := r.enterpriseClient.AuthApprove(r.args.enterpriseAuthApproveFingerprint)
	if err != nil {
		return err
	}

	return nil
}
