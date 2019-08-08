package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitEntitlementsCommand(parent *cobra.Command) *cobra.Command {
	entitlementsCmd := &cobra.Command{
		Use:   "entitlements",
		Short: "Manage customer entitlements",
		Long:  `The entitlements command allows vendors to create, display, modify entitlement values for end customer licensing.`,
	}
	parent.AddCommand(entitlementsCmd)

	entitlementsCmd.PersistentFlags().StringVar(&r.args.entitlementsAPIServer, "replicated-api-server", "https://g.replicated.com/graphql", "upstream g. address")
	entitlementsCmd.PersistentFlags().BoolVarP(&r.args.entitlementsVerbose, "verbose", "p", false, "verbose logging")

	return entitlementsCmd
}
