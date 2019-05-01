package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func (r *runners) InitEntitlementsSetValueCommand(parent *cobra.Command) {
	entitlementsSetValue := &cobra.Command{
		Use:   "set-value",
		Short: "Set a customer value for an entitlement field defined via 'replicated entitlements define-fields'",
		Long:  `Set a customer value for an entitlement field defined via 'replicated entitlements define-fields'`,
		RunE:  r.entitlementsSetValue,
	}
	parent.AddCommand(entitlementsSetValue)

	entitlementsSetValue.Flags().StringVar(&r.args.entitlementsSetValueDefinitionsID, "definitions-id", "", "definitions id created with define-fields command")
	entitlementsSetValue.Flags().StringVar(&r.args.entitlementsSetValueCustomerID, "customer-id", "", "customer id to assign the value to")
	entitlementsSetValue.Flags().StringVar(&r.args.entitlementsSetValueKey, "key", "", "field key")
	entitlementsSetValue.Flags().StringVar(&r.args.entitlementsSetValueValue, "value", "", "value to set")
	entitlementsSetValue.Flags().StringVar(&r.args.entitlementsSetValueType, "type", "string", "type of data. Much match 'type' in the field definition. Defaults to 'string'.")

	entitlementsSetValue.MarkFlagRequired("definitions-id")
	entitlementsSetValue.MarkFlagRequired("customer-id")
	entitlementsSetValue.MarkFlagRequired("key")
	entitlementsSetValue.MarkFlagRequired("value")
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
