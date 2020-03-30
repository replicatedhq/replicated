package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func (r *runners) InitEntitlementsSetValueCommand(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "set-value",
		Short: "Set a customer value for an entitlement field defined via 'replicated entitlements define-fields'",
		Long:  `Set a customer value for an entitlement field defined via 'replicated entitlements define-fields'`,
		RunE:  r.entitlementsSetValue,
	}
	cmd.Hidden = true // Not supported in KOTS
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.entitlementsSetValueDefinitionsID, "definitions-id", "", "definitions id created with define-fields command")
	cmd.Flags().StringVar(&r.args.entitlementsSetValueCustomerID, "customer-id", "", "customer id to assign the value to")
	cmd.Flags().StringVar(&r.args.entitlementsSetValueKey, "key", "", "field key")
	cmd.Flags().StringVar(&r.args.entitlementsSetValueValue, "value", "", "value to set")
	cmd.Flags().StringVar(&r.args.entitlementsSetValueType, "type", "string", "type of data. Much match 'type' in the field definition. Defaults to 'string'.")

	cmd.MarkFlagRequired("definitions-id")
	cmd.MarkFlagRequired("customer-id")
	cmd.MarkFlagRequired("key")
	cmd.MarkFlagRequired("value")
}

func (r *runners) entitlementsSetValue(cmd *cobra.Command, args []string) error {
	created, err := r.api.SetEntitlementValue(
		r.args.entitlementsSetValueCustomerID,
		r.args.entitlementsSetValueDefinitionsID,
		r.args.entitlementsSetValueKey,
		r.args.entitlementsSetValueValue,
		r.args.entitlementsSetValueType,
		r.appID,
	)

	if err != nil {
		return err
	}

	bytes, _ := json.MarshalIndent(created, "", "  ")
	fmt.Printf("%s\n", bytes)

	return nil
}
