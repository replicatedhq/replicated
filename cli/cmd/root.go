package cmd

import (
	"errors"
	"io"
	"os"
	"text/tabwriter"

	"github.com/replicatedhq/replicated/client"
	"github.com/spf13/cobra"
)

// table output settings
const (
	minWidth = 0
	tabWidth = 8
	padding  = 4
	padChar  = ' '
)

var appSlug string
var apiToken string
var apiOrigin = "https://api.replicated.com/vendor"

func init() {
	RootCmd.PersistentFlags().StringVar(&appSlug, "app", "", "The app slug to use in all calls")
	RootCmd.PersistentFlags().StringVar(&apiToken, "token", "", "The API token to use to access your app in the Vendor API")

	originFromEnv := os.Getenv("REPLICATED_API_ORIGIN")
	if originFromEnv != "" {
		apiOrigin = originFromEnv
	}
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "replicated",
	Short: "Manage channels and releases",
	Long:  `The replicated CLI allows vendors to manage their apps' channels and releases.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(w io.Writer) error {
	// get api client and app ID after flags are parsed
	runCmds := &runners{w: tabwriter.NewWriter(w, minWidth, tabWidth, padding, padChar, tabwriter.TabIndent)}

	channelCreateCmd.RunE = runCmds.channelCreate
	channelInspectCmd.RunE = runCmds.channelInspect
	channelLsCmd.RunE = runCmds.channelList
	channelRmCmd.RunE = runCmds.channelRemove
	releaseCreateCmd.RunE = runCmds.releaseCreate
	releaseInspectCmd.RunE = runCmds.releaseInspect
	releaseLsCmd.RunE = runCmds.releaseList
	releaseUpdateCmd.RunE = runCmds.releaseUpdate
	releasePromoteCmd.RunE = runCmds.releasePromote

	RootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if apiToken == "" {
			apiToken = os.Getenv("REPLICATED_API_TOKEN")
			if apiToken == "" {
				return errors.New("Please provide your API token")
			}
		}
		api := client.NewHTTPClient(apiOrigin, apiToken)
		runCmds.api = api

		if appSlug == "" {
			appSlug = os.Getenv("REPLICATED_APP_SLUG")
			if appSlug == "" {
				return errors.New("Please provide your app slug")
			}
		}

		// resolve app ID from slug
		app, err := api.GetAppBySlug(appSlug)
		if err != nil {
			return err
		}
		runCmds.appID = app.Id

		return nil
	}

	return RootCmd.Execute()
}
