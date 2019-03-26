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


func (r * runners) InitEntitlementsGetCustomerReleaseCommand(parent *cobra.Command) {
	var entitlementsGetCustomerRelease = &cobra.Command{
		Use:   "get-customer-release",
		Short: "Preview the end customer release view",
		Long: `Preview the end customer release view for a given customer ID + installation ID`,
		RunE: r.entitlementsGetCustomerRelease,
	}


	entitlementsGetCustomerRelease.Flags().StringVar(&r.args.entitlementsGetReleaseCustomerID,"customer-id", "", "customer id")
	entitlementsGetCustomerRelease.Flags().StringVar(&r.args.entitlementsGetReleaseInstallationID,"installation-id", "", "installation id")
	// change the default from g.replicated.com to pg.replicated.com
	entitlementsGetCustomerRelease.Flags().StringVar(&r.args.entitlementsGetReleaseAPIServer,"replicated-api-server", "https://pg.replicated.com/graphql", "Upstream api server")

	parent.AddCommand(entitlementsGetCustomerRelease)
}

func (r *runners) entitlementsGetCustomerRelease(cmd *cobra.Command, args []string) error {
	stdoutLogger := log.NewLogfmtLogger(os.Stdout)
	stdoutLogger = log.With(stdoutLogger, "ts", log.DefaultTimestampUTC)
	if r.args.entitlementsVerbose {
		stdoutLogger = level.NewFilter(stdoutLogger, level.AllowDebug())
	} else {
		stdoutLogger = level.NewFilter(stdoutLogger, level.AllowWarn())
	}

	upstream, err := url.Parse(r.args.entitlementsGetReleaseAPIServer)

	if err != nil {
		return errors.Wrapf(err, "parse replicated-api-server URL %s", r.args.entitlementsGetReleaseAPIServer)
	}

	if r.args.entitlementsGetReleaseCustomerID == "" {
		return errors.New("missing parameter: customer-id")
	}

	if r.args.entitlementsGetReleaseInstallationID == "" {
		return errors.New("missing parameter: installation-id")
	}

	client := &entitlements.PremGraphQLClient{
		GQLServer: upstream,
		CustomerID: r.args.entitlementsGetReleaseCustomerID,
		InstallationID: r.args.entitlementsGetReleaseInstallationID,
		Logger:    stdoutLogger,
	}

	espec, err := client.FetchCustomerRelease()

	if err != nil {
		return errors.Wrap(err, "fetch release")
	}

	bytes, _ := json.MarshalIndent(espec, "", "  ")
	fmt.Printf("%s\n", bytes)

	return nil
}
