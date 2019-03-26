package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/client/entitlements"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/spf13/cobra"
)




func init() {
}
func (r *runners) InitEntitlementsDefineFields(parent *cobra.Command) {
	entitlementsDefineFields := &cobra.Command{
		Use:   "define-fields",
		Short: "Define entitlements fields",
		Long: `Define the fields that can be assigned on a per customer 
basis and delivered securely to your on-prem application`,
		RunE: r.entitlementsDefineFields,
	}


	entitlementsDefineFields.Flags().StringVar(&r.args.entitlementsDefineFieldsFile, "file", "entitlements.yaml", "definitions file to promote")
	entitlementsDefineFields.Flags().StringVar(&r.args.entitlementsDefineFieldsName, "name", "", "name for this definition")

	parent.AddCommand(entitlementsDefineFields)
}

func (r *runners) entitlementsDefineFields(cmd *cobra.Command, args []string) error {
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

	spec, err := ioutil.ReadFile(r.args.entitlementsDefineFieldsFile)
	if err != nil {
		return errors.Wrapf(err, "read file %s", r.args.entitlementsDefineFieldsFile)
	}

	if r.args.entitlementsDefineFieldsName == "" {
		return errors.Errorf("missing parameter: name ")
	}

	client := &entitlements.GraphQLClient{
		GQLServer: upstream,
		Token:     apiToken,
		Logger:    stdoutLogger,
	}

	definitions, err := client.CreateEntitlementSpec(r.args.entitlementsDefineFieldsName, string(spec), r.appID)
	if err != nil {
		return errors.Wrap(err, "create definitions")
	}

	_, err = client.SetDefaultEntitlementSpec(definitions.ID)
	if err != nil {
		return errors.Wrap(err, "set as default definitions")
	}

	bytes, _ := json.MarshalIndent(definitions, "", "  ")
	fmt.Printf("%s\n", bytes)

	return nil
}
