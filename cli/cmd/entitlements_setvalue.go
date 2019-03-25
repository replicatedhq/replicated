package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/client/entitlements"
	"net/url"
	"os"

	"github.com/spf13/cobra"
)


func (r *runners) InitEntitlementsSetValueCommand(parent *cobra.Command) {
	entitlementsSetValue := &cobra.Command{
		Use:   "set-value",
		Short: "Set a customer value for an entitlement field defined via 'replicated entitlements define-fields'",
		Long:  `Set a customer value for an entitlement field defined via 'replicated entitlements define-fields'`,
		RunE: r.entitlementsSetValue,
	}
	parent.AddCommand(entitlementsSetValue)

	entitlementsSetValue.Flags().StringVar(&r.args.entitlementsSetValueDefinitionsID, "definitions-id", "", "definitions id created with define-fields command")
	entitlementsSetValue.Flags().StringVar(&r.args.entitlementsSetValueCustomerID, "customer-id", "", "customer id to assign the value to")
	entitlementsSetValue.Flags().StringVar(&r.args.entitlementsSetValueKey, "key", "", "field key")
	entitlementsSetValue.Flags().StringVar(&r.args.entitlementsSetValueValue, "value", "", "value to set")
	entitlementsSetValue.Flags().StringVar(&r.args.entitlementsSetValueType, "type", "string", "type of data. Much match 'type' in the field definition. Defaults to 'string'.")
}

func (r *runners) entitlementsSetValue(cmd *cobra.Command, args []string) error {
	stdoutLogger := log.NewLogfmtLogger(os.Stdout)
	stdoutLogger = log.With(stdoutLogger, "ts", log.DefaultTimestampUTC)
	if r.args.entitlementsVerbose {
		stdoutLogger = level.NewFilter(stdoutLogger, level.AllowDebug())
	} else {
		stdoutLogger = level.NewFilter(stdoutLogger, level.AllowWarn())
	}

	upstream, err := url.Parse(r.args.entitlementsAPIServer)

	if err != nil {
		return errors.Wrapf(err, "parse replicated-api-server URL %s", r.args.entitlementsAPIServer)
	}

	if r.args.entitlementsSetValueDefinitionsID == "" {
		return errors.New("missing parameter: definitions-id")
	}

	if r.args.entitlementsSetValueCustomerID == "" {
		return errors.New("missing parameter: customer-id")
	}

	if r.args.entitlementsSetValueKey == "" {
		return errors.New("missing parameter: key")
	}

	if r.args.entitlementsSetValueValue == "" {
		return errors.New("missing parameter: value")
	}

	client := &entitlements.GraphQLClient{
		GQLServer: upstream,
		Token:     apiToken,
		Logger:    stdoutLogger,
	}

	created, err := client.SetEntitlementValue(
		r.args.entitlementsSetValueCustomerID,
		r.args.entitlementsSetValueDefinitionsID,
		r.args.entitlementsSetValueKey,
		r.args.entitlementsSetValueValue,
		r.args.entitlementsSetValueType,
		r.appID,
	)

	if err != nil {
		return errors.Wrap(err, "set value")
	}

	bytes, _ := json.MarshalIndent(created, "", "  ")
	fmt.Printf("%s\n", bytes)

	return nil
}
