package cmd

import (
	"io"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/pkg/errors"

	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/shipclient"
	"github.com/replicatedhq/replicated/pkg/version"

	"github.com/replicatedhq/replicated/client"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

// table output settings
const (
	minWidth = 0
	tabWidth = 8
	padding  = 4
	padChar  = ' '
)

var appSlugOrID string
var apiToken string
var enterprisePrivateKeyPath = filepath.Join(homeDir(), ".replicated", "enterprise", "ecdsa")
var platformOrigin = "https://api.replicated.com/vendor"
var graphqlOrigin = "https://g.replicated.com/graphql"
var kurlDotSHOrigin = "https://kurl.sh"
var enterpriseOrigin = "https://api.replicated.com/enterprise"

func init() {
	originFromEnv := os.Getenv("REPLICATED_API_ORIGIN")
	if originFromEnv != "" {
		platformOrigin = originFromEnv
	}

	shipOriginFromEnv := os.Getenv("REPLICATED_SHIP_ORIGIN")
	if shipOriginFromEnv != "" {
		graphqlOrigin = shipOriginFromEnv
	}

	enterpriseOriginFromEnv := os.Getenv("REPLICATED_ENTERPRISE_ORIGIN")
	if enterpriseOriginFromEnv != "" {
		enterpriseOrigin = enterpriseOriginFromEnv
	}

	graphqlOriginFromEnv := os.Getenv("REPLICATED_GRAPHQL_ORIGIN")
	if graphqlOriginFromEnv != "" {
		graphqlOrigin = graphqlOriginFromEnv
	}
}

// RootCmd represents the base command when called without any subcommands
func GetRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "replicated",
		Short: "Manage channels, releases and collectors",
		Long:  `The replicated CLI allows vendors to manage their apps, channels, releases and collectors.`,
	}
	rootCmd.PersistentFlags().StringVar(&appSlugOrID, "app", "", "The app slug or app id to use in all calls")
	rootCmd.PersistentFlags().StringVar(&apiToken, "token", "", "The API token to use to access your app in the Vendor API")

	return rootCmd
}

// Almost the same as the default but don't print help subcommand
var rootCmdUsageTmpl = `
Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

  Examples:
  {{.Example}}{{end}}{{if .HasAvailableSubCommands}}

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

	channelCmd := &cobra.Command{
		Use:   "channel",
		Short: "List channels",
		Long:  "List channels",
	}
	var releaseCmd = &cobra.Command{
		Use:   "release",
		Short: "Manage app releases",
		Long:  `The release command allows vendors to create, display, and promote their releases.`,
	}

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

	installerCmd := runCmds.InitInstallerCommand(runCmds.rootCmd)
	runCmds.InitInstallerCreate(installerCmd)
	runCmds.InitInstallerList(installerCmd)

	enterpriseCmd := runCmds.InitEnterpriseCommand(runCmds.rootCmd)
	enterpriseAuthCmd := runCmds.InitEnterpriseAuth(enterpriseCmd)
	runCmds.InitEnterpriseAuthInit(enterpriseAuthCmd)
	runCmds.InitEnterpriseAuthApprove(enterpriseAuthCmd)
	enterpriseChannelCmd := runCmds.InitEnterpriseChannel(enterpriseCmd)
	runCmds.InitEnterpriseChannelLS(enterpriseChannelCmd)
	runCmds.InitEnterpriseChannelCreate(enterpriseChannelCmd)
	runCmds.InitEnterpriseChannelUpdate(enterpriseChannelCmd)
	runCmds.InitEnterpriseChannelRM(enterpriseChannelCmd)
	runCmds.InitEnterpriseChannelAssign(enterpriseChannelCmd)
	enterprisePolicyCmd := runCmds.InitEnterprisePolicy(enterpriseCmd)
	runCmds.InitEnterprisePolicyLS(enterprisePolicyCmd)
	runCmds.InitEnterprisePolicyCreate(enterprisePolicyCmd)
	runCmds.InitEnterprisePolicyUpdate(enterprisePolicyCmd)
	runCmds.InitEnterprisePolicyRM(enterprisePolicyCmd)
	runCmds.InitEnterprisePolicyAssign(enterprisePolicyCmd)
	runCmds.InitEnterprisePolicyUnassign(enterprisePolicyCmd)
	enterpriseInstallerCmd := runCmds.InitEnterpriseInstaller(enterpriseCmd)
	runCmds.InitEnterpriseInstallerLS(enterpriseInstallerCmd)
	runCmds.InitEnterpriseInstallerCreate(enterpriseInstallerCmd)
	runCmds.InitEnterpriseInstallerUpdate(enterpriseInstallerCmd)
	runCmds.InitEnterpriseInstallerRM(enterpriseInstallerCmd)
	runCmds.InitEnterpriseInstallerAssign(enterpriseInstallerCmd)

	appCmd := runCmds.InitAppCommand(runCmds.rootCmd)
	runCmds.InitAppList(appCmd)
	runCmds.InitAppCreate(appCmd)
	runCmds.InitAppDelete(appCmd)

	registryCmd := runCmds.InitRegistryCommand(runCmds.rootCmd)
	runCmds.InitRegistryList(registryCmd)
	runCmds.InitRegistryRemove(registryCmd)
	runCmds.InitRegistryTest(registryCmd)
	runCmds.InitRegistryLogs(registryCmd)
	registryAddCmd := runCmds.InitRegistryAdd(registryCmd)
	runCmds.InitRegistryAddECR(registryAddCmd)
	runCmds.InitRegistryAddGCR(registryAddCmd)
	runCmds.InitRegistryAddGHCR(registryAddCmd)
	runCmds.InitRegistryAddDockerHub(registryAddCmd)
	runCmds.InitRegistryAddQuay(registryAddCmd)
	runCmds.InitRegistryAddOther(registryAddCmd)

	runCmds.rootCmd.SetUsageTemplate(rootCmdUsageTmpl)

	preRunSetupAPIs := func(_ *cobra.Command, _ []string) error {
		if apiToken == "" {
			apiToken = os.Getenv("REPLICATED_API_TOKEN")
			if apiToken == "" {
				return errors.New("Please provide your API token")
			}
		}

		// allow override
		if os.Getenv("KURL_SH_ORIGIN") != "" {
			kurlDotSHOrigin = os.Getenv("KURL_SH_ORIGIN")
		}

		platformAPI := platformclient.NewHTTPClient(platformOrigin, apiToken)
		runCmds.platformAPI = platformAPI

		shipAPI := shipclient.NewGraphQLClient(graphqlOrigin, apiToken)
		runCmds.shipAPI = shipAPI

		httpClient := platformclient.NewHTTPClient(platformOrigin, apiToken)
		kotsAPI := &kotsclient.VendorV3Client{HTTPClient: *httpClient}
		runCmds.kotsAPI = kotsAPI

		commonAPI := client.NewClient(platformOrigin, graphqlOrigin, apiToken, kurlDotSHOrigin)
		runCmds.api = commonAPI

		version.PrintIfUpgradeAvailable()

		return nil
	}

	prerunCommand := func(cmd *cobra.Command, args []string) error {
		if err := preRunSetupAPIs(cmd, args); err != nil {
			return errors.Wrap(err, "set up APIs")
		}

		if appSlugOrID == "" {
			appSlugOrID = os.Getenv("REPLICATED_APP")
		}

		app, appType, err := runCmds.api.GetAppType(appSlugOrID)
		if err != nil {
			return err
		}

		runCmds.appType = appType

		runCmds.appID = app.ID
		runCmds.appSlug = app.Slug

		return nil
	}

	channelCmd.PersistentPreRunE = prerunCommand
	releaseCmd.PersistentPreRunE = prerunCommand
	collectorsCmd.PersistentPreRunE = prerunCommand
	customersCmd.PersistentPreRunE = prerunCommand
	installerCmd.PersistentPreRunE = prerunCommand
	appCmd.PersistentPreRunE = preRunSetupAPIs
	registryCmd.PersistentPreRunE = preRunSetupAPIs

	runCmds.rootCmd.AddCommand(Version())

	return runCmds.rootCmd.Execute()
}
