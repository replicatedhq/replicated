package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

func (r *runners) InitEntitlementsDefineFields(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "define-fields",
		Short: "Define entitlements fields",
		Long: `Define the fields that can be assigned on a per customer
basis and delivered securely to your on-prem application`,
		RunE: r.entitlementsDefineFields,
	}

	cmd.Flags().StringVar(&r.args.entitlementsDefineFieldsFile, "file", "entitlements.yaml", "definitions file to promote")
	cmd.Flags().StringVar(&r.args.entitlementsDefineFieldsName, "name", "", "name for this definition")
	cmd.Hidden=true; // Not supported in KOTS (ch #22646)
	parent.AddCommand(cmd)
}

func (r *runners) entitlementsDefineFields(cmd *cobra.Command, args []string) error {
	spec, err := ioutil.ReadFile(r.args.entitlementsDefineFieldsFile)
	if err != nil {
		return err
	}

	definitions, err := r.api.CreateEntitlementSpec(r.appID, r.appType, r.args.entitlementsDefineFieldsName, string(spec))
	if err != nil {
		return errors.Wrap(err, "create definitions")
	}

	err = r.api.SetDefaultEntitlementSpec(definitions.ID)
	if err != nil {
		return errors.Wrap(err, "set as default definitions")
	}

	bytes, _ := json.MarshalIndent(definitions, "", "  ")
	fmt.Printf("%s\n", bytes)

	return nil
}
