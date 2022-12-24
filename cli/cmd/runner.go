package cmd

import (
	"encoding/json"
	"github.com/pkg/errors"
	"io"
	"text/tabwriter"
	"text/template"
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
	appSlug          string
	appType          string
	api              client.Client
	enterpriseClient *enterpriseclient.HTTPClient
	platformAPI      *platformclient.HTTPClient
	shipAPI          *shipclient.GraphQLClient
	kotsAPI          *kotsclient.VendorV3Client
	stdin            io.Reader
	stdout           io.Writer
	dir              string
	w                *tabwriter.Writer

	rootCmd *cobra.Command
	args    runnerArgs
}

// prototype for allowing json output, not used in many methods yet
func (r *runners) Print(tmpl *template.Template, objects any) error {
	if r.args.output != "json" {
		if err := tmpl.Execute(r.w, objects); err != nil {
			return err
		}
		return r.w.Flush()
	}

	bytes, err := json.Marshal(objects)
	if err != nil {
		return errors.Wrap(err, "marshall objects")
	}
	_, err = r.stdout.Write(bytes)

	if err != nil {
		return errors.Wrap(err, "write bytes")
	}

	return nil
}

type runnerArgs struct {
	output string

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
	// Add Create Release Lint
	createReleaseLint     bool
	lintReleaseYamlDir    string
	lintReleaseFailOn     string
	releaseOptional       bool
	releaseRequired       bool
	releaseNotes          string
	releaseVersion        string
	updateReleaseYaml     string
	updateReleaseYamlDir  string
	updateReleaseYamlFile string

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

	customerCreateName                string
	customerCreateChannel             string
	customerCreateEnsureChannel       bool
	customerCreateExpiryDuration      time.Duration
	customerCreateIsAirgapEnabled     bool
	customerCreateIsGitopsSupported   bool
	customerCreateIsSnapshotSupported bool

	createInstallerYaml                 string
	createInstallerYamlFile             string
	createInstallerPromote              string
	createInstallerPromoteEnsureChannel bool

	enterpriseAuthInitCreateOrg string

	enterpriseAuthApproveFingerprint string

	enterpriseChannelCreateName        string
	enterpriseChannelCreateDescription string

	enterpriseChannelUpdateID          string
	enterpriseChannelUpdateName        string
	enterpriseChannelUpdateDescription string

	enterpriseChannelRmId string

	enterpriseChannelAssignChannelID string
	enterpriseChannelAssignTeamID    string

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
	customerLicenseInspectCustomer       string
	customerLicenseInspectOutput         string
	createReleaseAutoDefaults            bool
	createReleaseAutoDefaultsAccept      bool

	releaseDownloadDest               string
	createInstallerAutoDefaults       bool
	createInstallerAutoDefaultsAccept bool
	deleteAppForceYes                 bool

	addRegistrySkipValidation             bool
	addRegistryAuthType                   string
	addRegistryEndpoint                   string
	addRegistryUsername                   string
	addRegistryPassword                   string
	addRegistryPasswordFromStdIn          bool
	addRegistryAccessKeyID                string
	addRegistrySecretAccessKey            string
	addRegistrySecretAccessKeyFromStdIn   bool
	addRegistryServiceAccountKey          string
	addRegistryServiceAccountKeyFromStdIn bool
	addRegistryToken                      string
	addRegistryTokenFromStdIn             bool

	testRegistryImage string
}
