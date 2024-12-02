package cmd

import (
	"io"
	"text/tabwriter"
	"time"

	"github.com/replicatedhq/replicated/client"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/cli/values"
)

// Runner holds the I/O dependencies and configurations used by individual
// commands, which are defined as methods on this type.
type runners struct {
	appID           string
	appSlug         string
	appType         string
	isFoundationApp bool
	api             client.Client
	platformAPI     *platformclient.HTTPClient
	kotsAPI         *kotsclient.VendorV3Client
	stdin           io.Reader
	dir             string
	outputFormat    string
	w               *tabwriter.Writer

	rootCmd *cobra.Command
	args    runnerArgs
}

func (r *runners) hasApp() bool {
	if r.appID == "" && r.appType == "" && r.appSlug == "" {
		return false
	}
	return true
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

	instanceInspectCustomer string
	instanceInspectInstance string
	instanceListCustomer    string
	instanceListTags        []string
	instanceTagCustomer     string
	instanceTagInstacne     string
	instanceTagTags         []string

	createInstallerYaml                 string
	createInstallerYamlFile             string
	createInstallerPromote              string
	createInstallerPromoteEnsureChannel bool

	createReleaseAutoDefaults       bool
	createReleaseAutoDefaultsAccept bool

	releaseDownloadDest               string
	createInstallerAutoDefaults       bool
	createInstallerAutoDefaultsAccept bool

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
	createClusterIPFamily               string
	createClusterLicenseID              string
	createClusterNodeCount              int
	createClusterMinNodeCount           string
	createClusterMaxNodeCount           string
	createClusterDiskGiB                int64
	createClusterDryRun                 bool
	createClusterTTL                    string
	createClusterWaitDuration           time.Duration
	createClusterInstanceType           string
	createClusterNodeGroups             []string
	createClusterTags                   []string

	upgradeClusterKubernetesVersion string
	upgradeClusterDryRun            bool
	upgradeClusterWaitDuration      time.Duration

	updateClusterName string
	updateClusterID   string

	updateClusterTTL string

	updateClusterNodeGroupID       string
	updateClusterNodeGroupName     string
	updateClusterNodeGroupCount    int
	updateClusterNodeGroupMinCount string
	updateClusterNodeGroupMaxCount string

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

	removeClusterAll    bool
	removeClusterTags   []string
	removeClusterNames  []string
	removeClusterDryRun bool

	modelCollectionCreateName           string
	modelCollectionAddModelName         string
	modelCollectionAddModelCollectionID string
	modelCollectionRmModelName          string
	modelCollectionRmModelCollectionID  string

	lsAppVersion           string
	lsVersionsDistribution string

	lsClusterShowTerminated bool
	lsClusterStartTime      string
	lsClusterEndTime        string
	lsClusterWatch          bool

	kubeconfigClusterName string
	kubeconfigClusterID   string
	kubeconfigPath        string
	kubeconfigStdout      bool

	shellClusterName string
	shellClusterID   string

	clusterExposePortPort       int
	clusterExposePortProtocols  []string
	clusterExposePortIsWildcard bool

	clusterPortRemoveAddonID   string
	clusterPortRemovePort      int
	clusterPortRemoveProtocols []string

	loginEndpoint string

	apiPostBody  string
	apiPutBody   string
	apiPatchBody string

	compatibilityKubernetesDistribution string
	compatibilityKubernetesVersion      string
	compatibilitySuccess                bool
	compatibilityFailure                bool
	compatibilityNotes                  string

	createVMName         string
	createVMDistribution string
	createVMVersion      string
	createVMCount        int
	createVMDiskGiB      int64
	createVMTTL          string
	createVMInstanceType string
	createVMWaitDuration time.Duration
	createVMTags         []string
	createVMNetwork      string
	createVMDryRun       bool

	lsVMShowTerminated bool
	lsVMStartTime      string
	lsVMEndTime        string
	lsVMWatch          bool

	removeVMAll    bool
	removeVMTags   []string
	removeVMNames  []string
	removeVMDryRun bool

	updateVMTTL string

	updateVMName string
	updateVMID   string

	vmExposePortPort       int
	vmExposePortProtocols  []string
	vmExposePortIsWildcard bool
	vmPortRemoveAddonID    string

	lsNetworkStartTime string
	lsNetworkEndTime   string
	lsNetworkWatch     bool
}
