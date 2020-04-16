package cmd

import (
	"io"
	"text/tabwriter"
	"time"

	"github.com/replicatedhq/replicated/pkg/kotsclient"

	"github.com/replicatedhq/replicated/pkg/shipclient"
	"github.com/spf13/cobra"

	"github.com/replicatedhq/replicated/client"
	"github.com/replicatedhq/replicated/pkg/platformclient"

	"github.com/replicatedhq/replicated/pkg/enterpriseclient"
)

// Runner holds the I/O dependencies and configurations used by individual
// commands, which are defined as methods on this type.
type runners struct {
	appID            string
	appType          string
	api              client.Client
	enterpriseClient *enterpriseclient.HTTPClient
	platformAPI      *platformclient.HTTPClient
	shipAPI          *shipclient.GraphQLClient
	kotsAPI          *kotsclient.GraphQLClient
	stdin            io.Reader
	dir              string
	w                *tabwriter.Writer

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

	createReleaseYaml                 string
	createReleaseYamlFile             string
	createReleaseYamlDir              string
	createReleaseConfigYaml           string
	createReleaseDeploymentYaml       string
	createReleaseServiceYaml          string
	createReleasePreflightYaml        string
	createReleaseSupportBundleYaml    string
	createReleasePromote              string
	createReleasePromoteDir           string
	createReleasePromoteRequired      bool
	createReleasePromoteNotes         string
	createReleasePromoteVersion       string
	createReleasePromoteEnsureChannel bool
	lintReleaseYamlDir                string
	lintReleaseFailOn                 string
	lintBetaLinter                    bool
	releaseOptional                   bool
	releaseNotes                      string
	releaseVersion                    string
	updateReleaseYaml                 string
	updateReleaseYamlDir              string
	updateReleaseYamlFile             string

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

	customerCreateName           string
	customerCreateChannel        string
	customerCreateEnsureChannel  bool
	customerCreateExpiryDuration time.Duration

	createInstallerYaml                 string
	createInstallerYamlFile             string
	createInstallerPromote              string
	createInstallerPromoteEnsureChannel bool

	enterpriseChannelCreateName        string
	enterpriseChannelCreateDescription string

	enterpriseChannelUpdateID          string
	enterpriseChannelUpdateName        string
	enterpriseChannelUpdateDescription string

	enterpriseChannelRmId string

	enterprisePolicyCreateName        string
	enterprisePolicyCreateDescription string
	enterprisePolicyCreateFile        string

	enterprisePolicyUpdateID          string
	enterprisePolicyUpdateName        string
	enterprisePolicyUpdateDescription string
	enterprisePolicyUpdateFile        string

	enterprisePolicyRmId string

	enterprisePolicyAssignPolicyID  string
	enterprisePolicyAssignChannelID string

	enterprisePolicyUnassignPolicyID  string
	enterprisePolicyUnassignChannelID string

	enterpriseInstallerCreateFile string

	enterpriseInstallerUpdateID   string
	enterpriseInstallerUpdateFile string

	enterpriseInstallerRmId string

	enterpriseInstallerAssignInstallerID string
	enterpriseInstallerAssignChannelID   string
}
