package cmd

import (
	"errors"
	"io"
	"os"
	"text/tabwriter"

	"github.com/replicatedhq/replicated/pkg/shipclient"

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
var platformOrigin = "https://api.replicated.com/vendor"
var graphqlOrigin = "https://g.replicated.com/graphql"

// var graphqlOrigin = "http://localhost:30013/graphql"

func init() {
	originFromEnv := os.Getenv("REPLICATED_API_ORIGIN")
	if originFromEnv != "" {
		platformOrigin = originFromEnv
	}

	shipOriginFromEnv := os.Getenv("REPLICATED_SHIP_ORIGIN")
	if shipOriginFromEnv != "" {
		graphqlOrigin = shipOriginFromEnv
	}
}

// RootCmd represents the base command when called without any subcommands
func GetRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "replicated",
		Short: "Manage channels and releases",
		Long:  `The replicated CLI allows vendors to manage their apps' channels and releases.`,
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
		runCmds.rootCmd.SetOutput(stderr)
	}

	channelCmd := &cobra.Command{
		Use:   "channel",
		Short: "Manage and review channels",
		Long:  "Manage and review channels",
	}
	var releaseCmd = &cobra.Command{
		Use:   "release",
		Short: "Manage app releases",
		Long:  `The release command allows vendors to create, display, modify, and archive their releases.`,
	}

	runCmds.rootCmd.AddCommand(channelCmd)
	runCmds.InitChannelCreate(channelCmd)
	runCmds.InitChannelInspect(channelCmd)
	runCmds.InitChannelAdoption(channelCmd)
	runCmds.InitChannelReleases(channelCmd)
	runCmds.InitChannelCounts(channelCmd)
	runCmds.InitChannelList(channelCmd)
	runCmds.InitChannelRemove(channelCmd)

	runCmds.rootCmd.AddCommand(releaseCmd)
	runCmds.InitReleaseCreate(releaseCmd)
	runCmds.InitReleaseInspect(releaseCmd)
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

	entitlementsCmd := runCmds.InitEntitlementsCommand(runCmds.rootCmd)
	runCmds.InitEntitlementsDefineFields(entitlementsCmd)
	runCmds.InitEntitlementsSetValueCommand(entitlementsCmd)
	runCmds.InitEntitlementsGetCustomerReleaseCommand(entitlementsCmd)

	runCmds.rootCmd.SetUsageTemplate(rootCmdUsageTmpl)

	runCmds.rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if apiToken == "" {
			apiToken = os.Getenv("REPLICATED_API_TOKEN")
			if apiToken == "" {
				return errors.New("Please provide your API token")
			}
		}
		platformAPI := platformclient.NewHTTPClient(platformOrigin, graphqlOrigin, apiToken)
		runCmds.platformAPI = platformAPI

		shipAPI := shipclient.NewGraphQLClient(graphqlOrigin, apiToken)
		runCmds.shipAPI = shipAPI

		commonAPI := client.NewClient(platformOrigin, graphqlOrigin, apiToken)
		runCmds.api = commonAPI

		if appSlugOrID == "" {
			appSlugOrID = os.Getenv("REPLICATED_APP")
		}

		appType, err := commonAPI.GetAppType(appSlugOrID)
		if err != nil {
			return err
		}
		runCmds.appType = appType

		if appType == "platform" {
			app, err := platformAPI.GetApp(appSlugOrID)
			if err != nil {
				return err
			}
			runCmds.appID = app.Id
		} else if appType == "ship" {
			app, err := shipAPI.GetApp(appSlugOrID)
			if err != nil {
				return err
			}
			runCmds.appID = app.ID
		}

		return nil
	}

	return runCmds.rootCmd.Execute()
}
