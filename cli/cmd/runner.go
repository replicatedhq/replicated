package cmd

import (
	"io"
	"text/tabwriter"
	"time"

	"github.com/replicatedhq/replicated/client"
	"github.com/replicatedhq/replicated/pkg/enterpriseclient"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/shipclient"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/cli/values"
)

// Runner holds the I/O dependencies and configurations used by individual
// commands, which are defined as methods on this type.
type runners struct {
	appID            string
	appSlug          string
	appType          string
	isFoundationApp  bool
	api              client.Client
	enterpriseClient *enterpriseclient.HTTPClient
	platformAPI      *platformclient.HTTPClient
	shipAPI          *shipclient.GraphQLClient
	kotsAPI          *kotsclient.VendorV3Client
	stdin            io.Reader
	dir              string
	outputFormat     string
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
	createReleaseChart                string
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
	lintReleaseChart      string
	lintReleaseFailOn     string
	releaseOptional       bool
	releaseRequired       bool
	releaseNotes          string
	releaseVersion        string
	updateReleaseYaml     string
	updateReleaseYamlDir  string
	updateReleaseYamlFile string
	updateReleaseChart    string

	customerLsIncludeTest bool

	customerCreateName                string
	customerCreateChannel             string
	customerCreateEnsureChannel       bool
	customerCreateExpiryDuration      time.Duration
	customerCreateIsAirgapEnabled     bool
	customerCreateIsGitopsSupported   bool
	customerCreateIsSnapshotSupported bool
	customerCreateIsKotInstallEnabled bool
	customerCreateEmail               string
	customerCreateType                string

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

	createClusterName                   string
	createClusterKubernetesDistribution string
	createClusterKubernetesVersion      string
	createClusterNodeCount              int
	createClusterDiskGiB                int64
	createClusterDryRun                 bool
	createClusterTTL                    string
	createClusterWaitDuration           time.Duration
	createClusterInstanceType           string

	prepareClusterID                     string
	prepareClusterName                   string
	prepareClusterKubernetesDistribution string
	prepareClusterKubernetesVersion      string
	prepareClusterNodeCount              int
	prepareClusterDiskGiB                int64
	prepareClusterTTL                    string
	prepareClusterInstanceType           string
	prepareClusterWaitDuration           time.Duration
	prepareClusterEntitlements           []string
	prepareClusterYaml                   string
	prepareClusterYamlFile               string
	prepareClusterYamlDir                string
	prepareClusterChart                  string
	prepareClusterValueOpts              values.Options
	prepareClusterNamespace              string
	prepareClusterKotsConfigValuesFile   string
	prepareClusterKotsSharedPassword     string
	prepareClusterAppReadyTimeout        time.Duration

	removeClusterAll bool

	lsAppVersion                            string
	lsVersionsClusterKubernetesDistribution string

	lsClusterShowTerminated bool
	lsClusterStartTime      string
	lsClusterEndTime        string
	lsClusterWatch          bool

	kubeconfigClusterName string
	kubeconfigClusterID   string
	kubeconfigPath        string
	kubeconfigStdout      bool

	loginEndpoint string

	apiPostBody string
	apiPutBody  string

	channelInspectOutputFormat string

	customerInspectOutputFormat string
	customerInspectCustomer     string

	compatibilityKubernetesDistribution string
	compatibilityKubernetesVersion      string
	compatibilitySuccess                bool
	compatibilityFailure                bool
	compatibilityNotes                  string
}
