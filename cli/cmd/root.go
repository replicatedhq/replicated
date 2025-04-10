package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/Masterminds/sprig/v3"
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/client"
	replicatedcache "github.com/replicatedhq/replicated/pkg/cache"
	"github.com/replicatedhq/replicated/pkg/credentials"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/replicatedhq/replicated/pkg/version"
	"github.com/spf13/cobra"
)

// table output settings
const (
	minWidth = 0
	tabWidth = 8
	padding  = 4
	padChar  = ' '
)

var (
	appSlugOrID     string
	apiToken        string
	platformOrigin  = "https://api.replicated.com/vendor"
	kurlDotSHOrigin = "https://kurl.sh"
	cache           *replicatedcache.Cache
	debugFlag       bool
)

func init() {
	originFromEnv := os.Getenv("REPLICATED_API_ORIGIN")
	if originFromEnv != "" {
		platformOrigin = originFromEnv
	}

	c, err := replicatedcache.InitCache()
	if err != nil {
		panic(err)
	}
	cache = c

	// Set debug mode from environment variable
	if os.Getenv("REPLICATED_DEBUG") == "1" || os.Getenv("REPLICATED_DEBUG") == "true" {
		debugFlag = true
		version.SetDebugMode(true)
	}
}

// RootCmd represents the base command when called without any subcommands
func GetRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "replicated",
		Short: "Manage your Commercial Software Distribution Lifecycle using Replicated",
		Long:  `The 'replicated' CLI allows Replicated customers (vendors) to manage their Commercial Software Distribution Lifecycle (CSDL) using the Replicated API.`,
	}
	rootCmd.PersistentFlags().StringVar(&appSlugOrID, "app", "", "The app slug or app id to use in all calls")
	rootCmd.PersistentFlags().StringVar(&apiToken, "token", "", "The API token to use to access your app in the Vendor API")
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Enable debug output")

	return rootCmd
}

// Almost the same as the default but don't print help subcommand
var rootCmdUsageTmpl = `
Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Example:
{{.Example |  indent 2}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if .IsAvailableCommand}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(rootCmd *cobra.Command, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	w := tabwriter.NewWriter(stdout, minWidth, tabWidth, padding, padChar, tabwriter.TabIndent)

	// get api client and app ID after flags are parsed
	runCmds := &runners{
		rootCmd: rootCmd,
		stdin:   stdin,
		w:       w,
	}
	if runCmds.rootCmd == nil {
		runCmds.rootCmd = GetRootCmd()
	}
	if stderr != nil {
		runCmds.rootCmd.SetErr(stderr)
	}
	if stdout != nil {
		runCmds.rootCmd.SetOut(stdout)
	}

	// Setup PersistentPreRun to handle --debug flag
	runCmds.rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		// Enable debug mode if flag is set
		if debugFlag {
			version.SetDebugMode(true)
		}
	}

	channelCmd := &cobra.Command{
		Use:   "channel",
		Short: "List channels",
		Long:  "List channels",
	}
	releaseCmd := &cobra.Command{
		Use:   "release",
		Short: "Manage app releases",
		Long:  `The release command allows vendors to create, display, and promote their releases.`,
	}

	runCmds.InitCompletionCommand(runCmds.rootCmd)

	runCmds.rootCmd.AddCommand(channelCmd)
	runCmds.InitChannelCreate(channelCmd)
	runCmds.InitChannelInspect(channelCmd)
	runCmds.InitChannelAdoption(channelCmd)
	runCmds.InitChannelReleases(channelCmd)
	runCmds.InitChannelCounts(channelCmd)
	runCmds.InitChannelList(channelCmd)
	runCmds.InitChannelRemove(channelCmd)
	runCmds.InitChannelEnableSemanticVersioning(channelCmd)
	runCmds.InitChannelDisableSemanticVersioning(channelCmd)
	runCmds.InitChannelReleaseDemote(channelCmd)
	runCmds.InitChannelReleaseUnDemote(channelCmd)

	runCmds.rootCmd.AddCommand(releaseCmd)
	err := runCmds.InitReleaseCreate(releaseCmd)
	if err != nil {
		return errors.Wrap(err, "initialize release create command")
	}
	runCmds.InitReleaseInspect(releaseCmd)
	runCmds.InitReleaseDownload(releaseCmd)
	runCmds.IniReleaseList(releaseCmd)
	runCmds.InitReleaseUpdate(releaseCmd)
	runCmds.InitReleasePromote(releaseCmd)
	runCmds.InitReleaseLint(releaseCmd)
	runCmds.InitReleaseTest(releaseCmd)
	runCmds.InitReleaseCompatibility(releaseCmd)

	collectorsCmd := runCmds.InitCollectorsCommand(runCmds.rootCmd)
	runCmds.InitCollectorList(collectorsCmd)
	runCmds.InitCollectorUpdate(collectorsCmd)
	runCmds.InitCollectorPromote(collectorsCmd)
	runCmds.InitCollectorCreate(collectorsCmd)
	runCmds.InitCollectorInspect(collectorsCmd)

	customersCmd := runCmds.InitCustomersCommand(runCmds.rootCmd)
	runCmds.InitCustomersLSCommand(customersCmd)
	runCmds.InitCustomersCreateCommand(customersCmd)
	runCmds.InitCustomersDownloadLicenseCommand(customersCmd)
	runCmds.InitCustomersArchiveCommand(customersCmd)
	runCmds.InitCustomersInspectCommand(customersCmd)
	runCmds.InitCustomerUpdateCommand(customersCmd)

	instanceCmd := runCmds.InitInstanceCommand(runCmds.rootCmd)
	runCmds.InitInstanceLSCommand(instanceCmd)
	runCmds.InitInstanceInspectCommand(instanceCmd)
	runCmds.InitInstanceTagCommand(instanceCmd)

	installerCmd := runCmds.InitInstallerCommand(runCmds.rootCmd)
	runCmds.InitInstallerCreate(installerCmd)
	runCmds.InitInstallerList(installerCmd)

	appCmd := runCmds.InitAppCommand(runCmds.rootCmd)
	runCmds.InitAppList(appCmd)
	runCmds.InitAppCreate(appCmd)
	runCmds.InitAppRm(appCmd)

	defaultCmd := runCmds.InitDefaultCommand(runCmds.rootCmd)
	runCmds.InitDefaultShowCommand(defaultCmd)
	runCmds.InitDefaultSetCommand(defaultCmd)
	runCmds.InitDefaultClearCommand(defaultCmd)
	runCmds.InitDefaultClearAllCommand(defaultCmd)

	enterprisePortalCmd := runCmds.InitEnterprisePortalCommand(runCmds.rootCmd)
	enterprisePortalStatusCmd := runCmds.InitEnterprisePortalStatusCmd(enterprisePortalCmd)
	runCmds.InitEnterprisePortalStatusGetCmd(enterprisePortalStatusCmd)
	runCmds.InitEnterprisePortalStatusUpdateCmd(enterprisePortalStatusCmd)
	runCmds.InitEnterprisePortalInviteCmd(enterprisePortalCmd)
	enterprisePortalUserCmd := runCmds.InitEnterprisePortalUserCmd(enterprisePortalCmd)
	runCmds.InitEnterprisePortalUserLsCmd(enterprisePortalUserCmd)

	registryCmd := runCmds.InitRegistryCommand(runCmds.rootCmd)
	runCmds.InitRegistryList(registryCmd)
	runCmds.InitRegistryRemove(registryCmd)
	runCmds.InitRegistryTest(registryCmd)
	runCmds.InitRegistryLogs(registryCmd)
	registryAddCmd := runCmds.InitRegistryAdd(registryCmd)
	runCmds.InitRegistryAddDockerHub(registryAddCmd)
	runCmds.InitRegistryAddECR(registryAddCmd)
	runCmds.InitRegistryAddGAR(registryAddCmd)
	runCmds.InitRegistryAddGCR(registryAddCmd)
	runCmds.InitRegistryAddGHCR(registryAddCmd)
	runCmds.InitRegistryAddQuay(registryAddCmd)
	runCmds.InitRegistryAddOther(registryAddCmd)

	modelCmd := runCmds.InitModelCommand(runCmds.rootCmd)
	runCmds.InitModelList(modelCmd)
	runCmds.InitModelRemove(modelCmd)
	runCmds.InitModelPush(modelCmd)
	runCmds.InitModelPull(modelCmd)
	collectionCmd := runCmds.InitCollectionCommand(modelCmd)
	runCmds.InitCollectionList(collectionCmd)
	runCmds.InitCollectionCreate(collectionCmd)
	runCmds.InitCollectionRemove(collectionCmd)
	runCmds.InitCollectionAddModel(collectionCmd)
	runCmds.InitCollectionRemoveModel(collectionCmd)

	clusterCmd := runCmds.InitClusterCommand(runCmds.rootCmd)
	runCmds.InitClusterCreate(clusterCmd)
	runCmds.InitClusterUpgrade(clusterCmd)
	runCmds.InitClusterList(clusterCmd)
	runCmds.InitClusterKubeconfig(clusterCmd)
	runCmds.InitClusterRemove(clusterCmd)
	runCmds.InitClusterVersions(clusterCmd)
	runCmds.InitClusterShell(clusterCmd)

	clusterNodeGroupCmd := runCmds.InitClusterNodeGroup(clusterCmd)
	runCmds.InitClusterNodeGroupList(clusterNodeGroupCmd)

	clusterAddonCmd := runCmds.InitClusterAddon(clusterCmd)
	runCmds.InitClusterAddonLs(clusterAddonCmd)
	runCmds.InitClusterAddonRm(clusterAddonCmd)
	clusterAddonCreateCmd := runCmds.InitClusterAddonCreate(clusterAddonCmd)
	runCmds.InitClusterAddonCreateObjectStore(clusterAddonCreateCmd)

	clusterPortCmd := runCmds.InitClusterPort(clusterCmd)
	runCmds.InitClusterPortLs(clusterPortCmd)
	runCmds.InitClusterPortExpose(clusterPortCmd)
	runCmds.InitClusterPortRm(clusterPortCmd)

	clusterPrepareCmd := runCmds.InitClusterPrepare(clusterCmd)

	clusterUpdateCmd := runCmds.InitClusterUpdateCommand(clusterCmd)
	runCmds.InitClusterUpdateTTL(clusterUpdateCmd)
	runCmds.InitClusterUpdateNodegroup(clusterUpdateCmd)

	vmCmd := runCmds.InitVMCommand(runCmds.rootCmd)
	runCmds.InitVMCreate(vmCmd)
	runCmds.InitVMList(vmCmd)
	runCmds.InitVMVersions(vmCmd)
	runCmds.InitVMRemove(vmCmd)
	runCmds.InitVMSSH(vmCmd)
	runCmds.InitVMSCP(vmCmd)
	runCmds.InitVMSSHEndpoint(vmCmd)
	vmUpdateCmd := runCmds.InitVMUpdateCommand(vmCmd)
	runCmds.InitVMUpdateTTL(vmUpdateCmd)

	vmPortCmd := runCmds.InitVMPort(vmCmd)
	runCmds.InitVMPortLs(vmPortCmd)
	runCmds.InitVMPortExpose(vmPortCmd)
	runCmds.InitVMPortRm(vmPortCmd)

	networkCmd := runCmds.InitNetworkCommand(runCmds.rootCmd)
	runCmds.InitNetworkCreate(networkCmd)
	runCmds.InitNetworkList(networkCmd)
	runCmds.InitNetworkRemove(networkCmd)
	runCmds.InitNetworkJoin(networkCmd)

	networkUpdateCmd := runCmds.InitNetworkUpdateCommand(networkCmd)
	runCmds.InitNetworkUpdateOutbound(networkUpdateCmd)
	runCmds.InitNetworkUpdatePolicy(networkUpdateCmd)

	runCmds.InitLoginCommand(runCmds.rootCmd)
	runCmds.InitLogoutCommand(runCmds.rootCmd)

	apiCmd := runCmds.InitAPICommand(runCmds.rootCmd)
	runCmds.InitAPIGet(apiCmd)
	runCmds.InitAPIPost(apiCmd)
	runCmds.InitAPIPut(apiCmd)
	runCmds.InitAPIPatch(apiCmd)

	cobra.AddTemplateFunc("indent", sprig.FuncMap()["indent"])
	runCmds.rootCmd.SetUsageTemplate(rootCmdUsageTmpl)

	preRunSetupAPIs := func(cmd *cobra.Command, args []string) error {
		if apiToken == "" {
			creds, err := credentials.GetCurrentCredentials()
			if err != nil {
				if err == credentials.ErrCredentialsNotFound {
					return errors.New("Please provide your API token or log in with `replicated login`")
				}
				return errors.Wrap(err, "get current credentials")
			}

			apiToken = creds.APIToken
		}

		// allow override
		if os.Getenv("KURL_SH_ORIGIN") != "" {
			kurlDotSHOrigin = os.Getenv("KURL_SH_ORIGIN")
		}

		platformAPI := platformclient.NewHTTPClient(platformOrigin, apiToken)
		runCmds.platformAPI = platformAPI

		httpClient := platformclient.NewHTTPClient(platformOrigin, apiToken)
		kotsAPI := &kotsclient.VendorV3Client{HTTPClient: *httpClient}
		runCmds.kotsAPI = kotsAPI

		commonAPI := client.NewClient(platformOrigin, apiToken, kurlDotSHOrigin)
		runCmds.api = commonAPI

		// Print update info from cache, then start background update for next time
		version.PrintIfUpgradeAvailable()
		version.CheckForUpdatesInBackground(version.Version(), "replicatedhq/replicated/cli")

		return nil
	}

	prerunCommand := func(cmd *cobra.Command, args []string) (err error) {
		if cmd.SilenceErrors { // when SilenceErrors is set, command wants to use custom error printer
			defer func() {
				printIfError(cmd, err)
			}()
		}

		if err = preRunSetupAPIs(cmd, args); err != nil {
			return errors.Wrap(err, "set up APIs")
		}

		if appSlugOrID == "" {
			if cache.DefaultApp != "" {
				appSlugOrID = cache.DefaultApp
			}
		}

		if appSlugOrID == "" {
			appSlugOrID = os.Getenv("REPLICATED_APP")
		}

		// attempt to load the app from cache
		if appSlugOrID != "" {
			app, err := cache.GetApp(appSlugOrID)
			if err != nil {
				return errors.Wrap(err, "get app from cache")
			}

			if app != nil {
				if app.Scheduler == "native" {
					runCmds.appType = "platform"
				} else {
					runCmds.appType = app.Scheduler
				}
				runCmds.appID = app.ID
				runCmds.appSlug = app.Slug
			}
		}

		if appSlugOrID != "" && (runCmds.appType == "" || runCmds.appID == "" || runCmds.appSlug == "") {
			app, appType, err := runCmds.api.GetAppType(context.TODO(), appSlugOrID, true)
			if err != nil {
				return errors.Wrap(err, "get app type")
			}

			if err := cache.SetApp(app); err != nil {
				return errors.Wrap(err, "set app in cache")
			}

			runCmds.appType = appType
			runCmds.appID = app.ID
			runCmds.appSlug = app.Slug
		}

		return nil
	}

	channelCmd.PersistentPreRunE = prerunCommand
	releaseCmd.PersistentPreRunE = prerunCommand
	collectorsCmd.PersistentPreRunE = prerunCommand
	installerCmd.PersistentPreRunE = prerunCommand
	customersCmd.PersistentPreRunE = prerunCommand
	instanceCmd.PersistentPreRunE = prerunCommand
	clusterPrepareCmd.PersistentPreRunE = prerunCommand
	enterprisePortalCmd.PersistentPreRunE = prerunCommand

	defaultCmd.PersistentPreRunE = preRunSetupAPIs
	appCmd.PersistentPreRunE = preRunSetupAPIs
	registryCmd.PersistentPreRunE = preRunSetupAPIs
	clusterCmd.PersistentPreRunE = preRunSetupAPIs
	vmCmd.PersistentPreRunE = preRunSetupAPIs
	networkCmd.PersistentPreRunE = preRunSetupAPIs
	apiCmd.PersistentPreRunE = preRunSetupAPIs
	modelCmd.PersistentPreRunE = preRunSetupAPIs

	runCmds.rootCmd.AddCommand(Version())

	return runCmds.rootCmd.Execute()
}

func printIfError(cmd *cobra.Command, err error) {
	if err == nil {
		return
	}

	cmd.SilenceUsage = true

	switch err := errors.Cause(err).(type) {
	case platformclient.APIError:
		fmt.Fprintln(os.Stderr, fmt.Sprintf("ERROR: %d", err.StatusCode))
		fmt.Fprintln(os.Stderr, fmt.Sprintf("METHOD: %s", err.Method))
		fmt.Fprintln(os.Stderr, fmt.Sprintf("ENDPOINT: %s", err.Endpoint))
		fmt.Fprintln(os.Stderr, err.Message) // note that this can have multiple lines
	default:
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err.Error())
	}
}

func parseTags(tags []string) ([]types.Tag, error) {
	parsedTags := []types.Tag{}
	for _, tag := range tags {
		tagParts := strings.SplitN(tag, "=", 2)
		if len(tagParts) != 2 {
			return nil, errors.Errorf("invalid tag format: %s", tag)
		}

		parsedTags = append(parsedTags, types.Tag{
			Key:   tagParts[0],
			Value: tagParts[1],
		})
	}
	return parsedTags, nil
}
