package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/logger"
	"github.com/replicatedhq/replicated/pkg/tools"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitAppHostnameListCommand(parent *cobra.Command) *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "List custom hostnames for an application",
		Long: `List all custom hostnames configured for an application.

This command fetches and displays all custom hostname configurations including:
- Registry hostnames
- Proxy hostnames
- Download Portal hostnames
- Replicated App hostnames

The app ID or slug can be provided via the --app flag or from the .replicated config file.`,
		Example: `# List all custom hostnames for an app
replicated app hostname ls --app myapp

# List hostnames and output as JSON
replicated app hostname ls --app myapp --output json`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// The parent chain is: rootCmd -> appCmd -> appHostnameCmd -> ls cmd
			// We need to call the app command's PersistentPreRunE (which is preRunSetupAPIs)
			// cmd.Parent() = appHostnameCmd
			// cmd.Parent().Parent() = appCmd
			hostnameCmd := cmd.Parent()
			if hostnameCmd != nil {
				appCmd := hostnameCmd.Parent()
				if appCmd != nil && appCmd.PersistentPreRunE != nil {
					if err := appCmd.PersistentPreRunE(cmd, args); err != nil {
						return err
					}
				}
			}

			// Load app from .replicated config if not provided via --app flag
			if r.appSlug == "" && r.appID == "" {
				parser := tools.NewConfigParser()
				config, err := parser.FindAndParseConfig(".")
				if err == nil && (config.AppSlug != "" || config.AppId != "") {
					if config.AppSlug != "" {
						r.appSlug = config.AppSlug
					} else if config.AppId != "" {
						r.appID = config.AppId
					}
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			return r.listAppHostnames(ctx, outputFormat)
		},
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "The output format to use. One of: json|table")

	return cmd
}

func (r *runners) listAppHostnames(ctx context.Context, outputFormat string) error {
	// Only show spinners for table output
	showSpinners := outputFormat == "table"
	log := logger.NewLogger(r.w)

	// Resolve app ID from slug or ID
	appSlugOrID := r.appSlug
	if appSlugOrID == "" {
		appSlugOrID = r.appID
	}
	if appSlugOrID == "" {
		return errors.New("app ID or slug is required (use --app flag or set in .replicated config)")
	}

	if showSpinners {
		log.ActionWithSpinner("Fetching app")
	}
	app, err := r.kotsAPI.GetApp(ctx, appSlugOrID, true)
	if err != nil {
		if showSpinners {
			log.FinishSpinnerWithError()
		}
		return errors.Wrap(err, "get app")
	}
	if showSpinners {
		log.FinishSpinner()
	}

	// Fetch default hostnames
	if showSpinners {
		log.ActionWithSpinner("Fetching default hostnames")
	}
	defaultHostnames, err := r.kotsAPI.ListDefaultHostnames(app.ID)
	if err != nil {
		if showSpinners {
			log.FinishSpinnerWithError()
		}
		return errors.Wrap(err, "list default hostnames")
	}
	if showSpinners {
		log.FinishSpinner()
	}

	// Fetch custom hostnames
	if showSpinners {
		log.ActionWithSpinner("Fetching custom hostnames")
	}
	customHostnames, err := r.kotsAPI.ListCustomHostnames(app.ID)
	if err != nil {
		if showSpinners {
			log.FinishSpinnerWithError()
		}
		return errors.Wrap(err, "list custom hostnames")
	}
	if showSpinners {
		log.FinishSpinner()
	}

	// Merge hostnames: start with defaults, override with custom values
	mergedHostnames := mergeHostnames(defaultHostnames, customHostnames)

	// Extract just the hostname strings from the merged result
	hostnameStrings := extractHostnameStrings(mergedHostnames)

	// Output based on format
	if outputFormat == "json" {
		jsonBytes, err := json.MarshalIndent(hostnameStrings, "", "  ")
		if err != nil {
			return errors.Wrap(err, "marshal json")
		}
		// Print directly without log prefix
		r.w.Write(jsonBytes)
		r.w.Write([]byte("\n"))
		r.w.Flush()
		return nil
	}

	if outputFormat == "table" {
		return printHostnamesTable(r.w, hostnameStrings)
	}

	return errors.Errorf("unsupported output format: %s", outputFormat)
}

// extractHostnameStrings extracts just the hostname strings from the merged hostnames
func extractHostnameStrings(merged *types.KotsAppCustomHostnames) map[string]string {
	result := make(map[string]string)

	// Take the first (default) hostname from each category
	if len(merged.Registry) > 0 {
		result["registry"] = merged.Registry[0].Hostname
	}

	if len(merged.Proxy) > 0 {
		result["proxy"] = merged.Proxy[0].Hostname
	}

	if len(merged.DownloadPortal) > 0 {
		result["downloadPortal"] = merged.DownloadPortal[0].Hostname
	}

	if len(merged.ReplicatedApp) > 0 {
		result["replicatedApp"] = merged.ReplicatedApp[0].Hostname
	}

	return result
}

// printHostnamesTable prints hostnames in a table format
func printHostnamesTable(w *tabwriter.Writer, hostnames map[string]string) error {
	fmt.Fprintln(w, "TYPE\tHOSTNAME")
	
	if registry, ok := hostnames["registry"]; ok && registry != "" {
		fmt.Fprintf(w, "Registry\t%s\n", registry)
	}
	
	if proxy, ok := hostnames["proxy"]; ok && proxy != "" {
		fmt.Fprintf(w, "Proxy\t%s\n", proxy)
	}
	
	if downloadPortal, ok := hostnames["downloadPortal"]; ok && downloadPortal != "" {
		fmt.Fprintf(w, "Download Portal\t%s\n", downloadPortal)
	}
	
	if replicatedApp, ok := hostnames["replicatedApp"]; ok && replicatedApp != "" {
		fmt.Fprintf(w, "Replicated App\t%s\n", replicatedApp)
	}
	
	w.Flush()
	return nil
}

// mergeHostnames merges default and custom hostnames.
// Defaults are simple strings, custom hostnames are arrays of detailed objects.
// For each category, if custom hostnames exist, use them; otherwise create a basic hostname from the default string.
func mergeHostnames(defaults *types.DefaultHostnames, custom *types.KotsAppCustomHostnames) *types.KotsAppCustomHostnames {
	if custom == nil && defaults == nil {
		return &types.KotsAppCustomHostnames{}
	}
	
	if custom == nil {
		custom = &types.KotsAppCustomHostnames{}
	}
	
	if defaults == nil {
		return custom
	}

	result := &types.KotsAppCustomHostnames{
		Registry:       mergeHostnameList(defaults.Registry, custom.Registry),
		Proxy:          mergeHostnameList(defaults.Proxy, custom.Proxy),
		DownloadPortal: mergeHostnameList(defaults.DownloadPortal, custom.DownloadPortal),
		ReplicatedApp:  mergeHostnameList(defaults.ReplicatedApp, custom.ReplicatedApp),
	}

	return result
}

// mergeHostnameList merges a default hostname string with custom hostname objects.
// If custom hostnames exist and contain the default hostname, use the custom data.
// If custom hostnames exist but don't contain the default, add both.
// If no custom hostnames, create a basic entry from the default string.
func mergeHostnameList(defaultHostname string, custom []types.KotsAppCustomHostname) []types.KotsAppCustomHostname {
	// If there's no default hostname, just return custom
	if defaultHostname == "" {
		return custom
	}

	// If there are no custom hostnames, create a basic one from the default
	if len(custom) == 0 {
		return []types.KotsAppCustomHostname{
			{
				IsDefault: true,
				CustomHostname: types.CustomHostname{
					Hostname: defaultHostname,
				},
			},
		}
	}

	// Check if any custom hostname matches the default
	foundDefault := false
	result := make([]types.KotsAppCustomHostname, 0, len(custom))
	for _, ch := range custom {
		if ch.Hostname == defaultHostname {
			foundDefault = true
			// Mark this as the default
			ch.IsDefault = true
		}
		result = append(result, ch)
	}

	// If the default hostname wasn't in the custom list, add it
	if !foundDefault {
		result = append(result, types.KotsAppCustomHostname{
			IsDefault: true,
			CustomHostname: types.CustomHostname{
				Hostname: defaultHostname,
			},
		})
	}

	return result
}
