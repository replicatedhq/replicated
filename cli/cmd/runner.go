package cmd

import (
	"io"
	"text/tabwriter"

	"github.com/replicatedhq/replicated/pkg/shipclient"
	"github.com/spf13/cobra"

	"github.com/replicatedhq/replicated/client"
	"github.com/replicatedhq/replicated/pkg/platformclient"
)

// Runner holds the I/O dependencies and configurations used by individual
// commands, which are defined as methods on this type.
type runners struct {
	appID       string
	appType     string
	api         client.Client
	platformAPI platformclient.Client
	shipAPI     shipclient.Client
	stdin       io.Reader
	w           *tabwriter.Writer

	rootCmd *cobra.Command
	args    runnerArgs
}

type runnerArgs struct {
	channelCreateName        string
	channelCreateDescription string

	createCollectorName     string
	createCollectorYaml     string
	createCollectorYamlFile string
	updateCollectorYaml     string
	updateCollectorYamlFile string
	updateCollectorName     string

	createReleaseYaml            string
	createReleaseYamlFile        string
	createReleasePromote         string
	createReleasePromoteRequired bool
	createReleasePromoteNotes    string
	createReleasePromoteVersion  string
	lintReleaseYaml              string
	lintReleaseYamlFile          string
	releaseOptional              bool
	releaseNotes                 string
	releaseVersion               string
	updateReleaseYaml            string
	updateReleaseYamlFile        string

	entitlementsAPIServer                string
	entitlementsVerbose                  bool
	entitlementsDefineFieldsFile         string
	entitlementsDefineFieldsName         string
	entitlementsGetReleaseCustomerID     string
	entitlementsGetReleaseInstallationID string
	entitlementsGetReleaseAPIServer      string
	entitlementsSetValueCustomerID       string
	entitlementsSetValueDefinitionsID    string
	entitlementsSetValueKey              string
	entitlementsSetValueValue            string
	entitlementsSetValueType             string
}
