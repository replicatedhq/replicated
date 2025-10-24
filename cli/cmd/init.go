package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/lint2"
	"github.com/replicatedhq/replicated/pkg/tools"
	"github.com/spf13/cobra"
)

func (r *runners) InitInitCommand(parent *cobra.Command) *cobra.Command {
	var nonInteractive bool
	var skipDetection bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a .replicated config file for your project",
		Long: `Initialize a .replicated config file for your project.

This command will guide you through setting up a .replicated configuration file
by prompting for common settings like app ID, chart paths, and linting preferences.

It will also attempt to auto-detect Helm charts and preflight specs in your project.`,
		Example: `# Initialize with interactive prompts
replicated config init

# Initialize with auto-detected resources only (no prompts)
replicated config init --non-interactive

# Initialize without auto-detection
replicated config init --skip-detection`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.initConfig(cmd, nonInteractive, skipDetection)
		},
	}

	cmd.Flags().BoolVar(&nonInteractive, "non-interactive", false, "Run without prompts, using defaults and auto-detected values")
	cmd.Flags().BoolVar(&skipDetection, "skip-detection", false, "Skip auto-detection of resources")

	parent.AddCommand(cmd)
	return cmd
}

func (r *runners) initConfig(cmd *cobra.Command, nonInteractive bool, skipDetection bool) error {
	// Check if we're in a non-interactive environment
	if !nonInteractive && tools.IsNonInteractive() {
		nonInteractive = true
		fmt.Fprintf(r.w, "Detected non-interactive environment, using defaults\n\n")
	}

	// Check if config already exists
	exists, existingPath, err := tools.ConfigExists(".")
	if err != nil {
		return errors.Wrap(err, "checking for existing config")
	}

	if exists {
		if nonInteractive {
			return fmt.Errorf("config file already exists at %s (use --force to overwrite)", existingPath)
		}

		// Ask if they want to overwrite
		prompt := promptui.Select{
			Label: fmt.Sprintf("Config file already exists at %s. What would you like to do?", existingPath),
			Items: []string{"Cancel", "Overwrite", "Edit/Merge"},
		}

		_, result, err := prompt.Run()
		if err != nil {
			return errors.Wrap(err, "prompting for action")
		}

		if result == "Cancel" {
			fmt.Fprintf(r.w, "Cancelled\n")
			return nil
		}

		if result == "Edit/Merge" {
			fmt.Fprintf(r.w, "Merge functionality not yet implemented. Please edit %s manually.\n", existingPath)
			return nil
		}

		// If "Overwrite", continue with init
		fmt.Fprintf(r.w, "Overwriting existing config...\n\n")
	}

	// Create new config
	config := &tools.Config{}

	// If API is available (profile flag used), offer to select from apps
	var selectedAppSlug string
	if r.kotsAPI != nil && !nonInteractive {
		appSlug, err := r.promptForAppSelection(cmd.Context())
		if err != nil {
			// If error fetching apps, just continue without it
			if !strings.Contains(err.Error(), "cancelled") {
				fmt.Fprintf(r.w, "Note: Could not fetch apps from API (%v)\n\n", err)
			} else {
				return err
			}
		} else if appSlug != "" {
			selectedAppSlug = appSlug
		}
	}

	// Auto-detect resources unless skipped
	var detected *tools.DetectedResources
	if !skipDetection {
		fmt.Fprintf(r.w, "Scanning project for resources...\n")
		detected, err = tools.AutoDetectResources(".")
		if err != nil {
			return errors.Wrap(err, "auto-detecting resources")
		}

		if len(detected.Charts) > 0 {
			fmt.Fprintf(r.w, "  Found %d Helm chart(s):\n", len(detected.Charts))
			for _, chart := range detected.Charts {
				fmt.Fprintf(r.w, "    - %s\n", chart)
			}
		}
		if len(detected.Preflights) > 0 {
			fmt.Fprintf(r.w, "  Found %d preflight spec(s):\n", len(detected.Preflights))
			for _, preflight := range detected.Preflights {
				fmt.Fprintf(r.w, "    - %s\n", preflight)
			}
		}
		if len(detected.SupportBundles) > 0 {
			fmt.Fprintf(r.w, "  Found %d support bundle spec(s):\n", len(detected.SupportBundles))
			for _, sb := range detected.SupportBundles {
				fmt.Fprintf(r.w, "    - %s\n", sb)
			}
		}
		if len(detected.ValuesFiles) > 0 {
			fmt.Fprintf(r.w, "  Found %d values file(s):\n", len(detected.ValuesFiles))
			for _, vf := range detected.ValuesFiles {
				fmt.Fprintf(r.w, "    - %s\n", vf)
			}
		}
		if len(detected.Manifests) > 0 {
			fmt.Fprintf(r.w, "  Found %d manifest directory pattern(s)\n", len(detected.Manifests))
		}
		fmt.Fprintf(r.w, "\n")
	}

	// Interactive prompts
	if !nonInteractive {
		// Use selected app from API if available, otherwise prompt
		if selectedAppSlug != "" {
			config.AppSlug = selectedAppSlug
		} else {
			// Prompt for app ID or slug
			appPrompt := promptui.Prompt{
				Label:   "App ID or App Slug (optional, check vendor.replicated.com)",
				Default: "",
			}
			appValue, err := appPrompt.Run()
			if err != nil {
				if err == promptui.ErrInterrupt {
					fmt.Fprintf(r.w, "\nCancelled\n")
					return nil
				}
				return errors.Wrap(err, "failed to read app value")
			}

			// Store in AppSlug by default since that's more commonly used
			// The API accepts both, and commands will resolve it
			if appValue != "" {
				config.AppSlug = appValue
			}
		}

		// Prompt for charts
		if len(detected.Charts) > 0 {
			useDetectedCharts := promptui.Select{
				Label: fmt.Sprintf("Use detected Helm charts? (%d found)", len(detected.Charts)),
				Items: []string{"Yes", "No", "Let me specify custom paths"},
			}
			_, chartChoice, err := useDetectedCharts.Run()
			if err != nil {
				if err == promptui.ErrInterrupt {
					fmt.Fprintf(r.w, "\nCancelled\n")
					return nil
				}
				return errors.Wrap(err, "failed to read chart choice")
			}

			switch chartChoice {
			case "Yes":
				for _, chartPath := range detected.Charts {
					// Convert to relative path with ./ prefix
					if !strings.HasPrefix(chartPath, ".") {
						chartPath = "./" + chartPath
					}
					config.Charts = append(config.Charts, tools.ChartConfig{
						Path: chartPath,
					})
				}
			case "Let me specify custom paths":
				charts, err := r.promptForChartPaths()
				if err != nil {
					return err
				}
				config.Charts = charts
			}
		} else {
			addCharts := promptui.Select{
				Label: "Add Helm charts?",
				Items: []string{"Yes", "No"},
			}
			_, addChartsResult, err := addCharts.Run()
			if err != nil {
				if err == promptui.ErrInterrupt {
					fmt.Fprintf(r.w, "\nCancelled\n")
					return nil
				}
				return errors.Wrap(err, "failed to read chart option")
			}

			if addChartsResult == "Yes" {
				charts, err := r.promptForChartPaths()
				if err != nil {
					return err
				}
				config.Charts = charts
			}
		}

		// Prompt for manifests
		if len(detected.Manifests) > 0 {
			useDetectedManifests := promptui.Select{
				Label: fmt.Sprintf("Use detected manifest patterns? (%d found)", len(detected.Manifests)),
				Items: []string{"Yes", "No", "Let me specify custom patterns"},
			}
			_, manifestChoice, err := useDetectedManifests.Run()
			if err != nil {
				if err == promptui.ErrInterrupt {
					fmt.Fprintf(r.w, "\nCancelled\n")
					return nil
				}
				return errors.Wrap(err, "failed to read manifest choice")
			}

			switch manifestChoice {
			case "Yes":
				config.Manifests = detected.Manifests
				// Add detected support bundles
				for _, sbPath := range detected.SupportBundles {
					if !strings.HasPrefix(sbPath, ".") {
						sbPath = "./" + sbPath
					}
					config.Manifests = append(config.Manifests, sbPath)
				}
			case "Let me specify custom patterns":
				manifests, err := r.promptForManifests()
				if err != nil {
					return err
				}
				config.Manifests = manifests
			}
		} else {
			// No manifest directories detected, but check for support bundles
			if len(detected.SupportBundles) > 0 {
				useSupportBundles := promptui.Select{
					Label: fmt.Sprintf("Add detected support bundle specs to manifests? (%d found)", len(detected.SupportBundles)),
					Items: []string{"Yes", "No"},
				}
				_, sbChoice, err := useSupportBundles.Run()
				if err != nil {
					if err == promptui.ErrInterrupt {
						fmt.Fprintf(r.w, "\nCancelled\n")
						return nil
					}
					return errors.Wrap(err, "failed to read support bundle choice")
				}

				if sbChoice == "Yes" {
					for _, sbPath := range detected.SupportBundles {
						if !strings.HasPrefix(sbPath, ".") {
							sbPath = "./" + sbPath
						}
						config.Manifests = append(config.Manifests, sbPath)
					}
				}
			}

			// Ask if they want to add manifest files manually
			addManifests := promptui.Select{
				Label: "Do you want to add any Kubernetes manifest files?",
				Items: []string{"No", "Yes"},
			}
			_, manifestsResult, err := addManifests.Run()
			if err != nil {
				if err == promptui.ErrInterrupt {
					fmt.Fprintf(r.w, "\nCancelled\n")
					return nil
				}
				return errors.Wrap(err, "failed to read manifest option")
			}

			if manifestsResult == "Yes" {
				manifests, err := r.promptForManifests()
				if err != nil {
					return err
				}
				config.Manifests = manifests
			}
		}

		// Prompt for preflights
		if len(detected.Preflights) > 0 {
			useDetectedPreflights := promptui.Select{
				Label: fmt.Sprintf("Use detected preflight specs? (%d found)", len(detected.Preflights)),
				Items: []string{"Yes", "No", "Let me specify custom paths"},
			}
			_, preflightChoice, err := useDetectedPreflights.Run()
			if err != nil {
				if err == promptui.ErrInterrupt {
					fmt.Fprintf(r.w, "\nCancelled\n")
					return nil
				}
				return errors.Wrap(err, "failed to read preflight choice")
			}

			switch preflightChoice {
			case "Yes":
				// Prompt user to select chart for preflights (if charts are available)
				var selectedChartName, selectedChartVersion string
				if len(config.Charts) > 0 {
					chartName, chartVersion, err := r.promptForChart(config.Charts)
					if err != nil {
						return err
					}
					selectedChartName = chartName
					selectedChartVersion = chartVersion
				}

				// Add all detected preflights with the selected chart
				for _, preflightPath := range detected.Preflights {
					// Convert to relative path with ./ prefix
					if !strings.HasPrefix(preflightPath, ".") {
						preflightPath = "./" + preflightPath
					}

					preflight := tools.PreflightConfig{
						Path:         preflightPath,
						ChartName:    selectedChartName,
						ChartVersion: selectedChartVersion,
					}

					config.Preflights = append(config.Preflights, preflight)
				}
			case "Let me specify custom paths":
				preflights, err := r.promptForPreflightPathsWithCharts(config.Charts, detected.ValuesFiles)
				if err != nil {
					return err
				}
				config.Preflights = preflights
			}
		} else {
			addPreflights := promptui.Select{
				Label: "Add preflight specs?",
				Items: []string{"No", "Yes"},
			}
			_, addPreflightsResult, err := addPreflights.Run()
			if err != nil {
				if err == promptui.ErrInterrupt {
					fmt.Fprintf(r.w, "\nCancelled\n")
					return nil
				}
				return errors.Wrap(err, "failed to read preflight option")
			}

			if addPreflightsResult == "Yes" {
				preflights, err := r.promptForPreflightPathsWithCharts(config.Charts, detected.ValuesFiles)
				if err != nil {
					return err
				}
				config.Preflights = preflights
			}
		}

		// Prompt for linting configuration
		if len(config.Charts) > 0 || len(config.Preflights) > 0 {
			configureLinting := promptui.Select{
				Label: "Configure linting? (recommended)",
				Items: []string{"Yes", "No"},
			}
			_, lintingResult, err := configureLinting.Run()
			if err != nil {
				if err == promptui.ErrInterrupt {
					fmt.Fprintf(r.w, "\nCancelled\n")
					return nil
				}
				return errors.Wrap(err, "failed to read linting option")
			}

			if lintingResult == "Yes" {
				lintConfig, err := r.promptForLintConfig(len(config.Charts) > 0, len(config.Preflights) > 0)
				if err != nil {
					return err
				}
				config.ReplLint = lintConfig
			}
		}
	} else {
		// Non-interactive mode: use detected resources
		if detected != nil {
			for _, chartPath := range detected.Charts {
				if !strings.HasPrefix(chartPath, ".") {
					chartPath = "./" + chartPath
				}
				config.Charts = append(config.Charts, tools.ChartConfig{
					Path: chartPath,
				})
			}

			// Auto-assign first detected chart to preflights (if available)
			var autoChartName, autoChartVersion string
			if len(config.Charts) > 0 {
				// Get metadata from first chart
				metadata, err := lint2.GetChartMetadata(config.Charts[0].Path)
				if err == nil {
					autoChartName = metadata.Name
					autoChartVersion = metadata.Version
				}
			}

			// Add detected preflights with auto-assigned chart (if available)
			for _, preflightPath := range detected.Preflights {
				if !strings.HasPrefix(preflightPath, ".") {
					preflightPath = "./" + preflightPath
				}

				preflight := tools.PreflightConfig{
					Path:         preflightPath,
					ChartName:    autoChartName,    // Auto-assigned from first chart
					ChartVersion: autoChartVersion, // Auto-assigned from first chart
				}

				config.Preflights = append(config.Preflights, preflight)
			}

			// Use detected manifest patterns
			config.Manifests = detected.Manifests

			// Add detected support bundles to manifests
			for _, sbPath := range detected.SupportBundles {
				if !strings.HasPrefix(sbPath, ".") {
					sbPath = "./" + sbPath
				}
				config.Manifests = append(config.Manifests, sbPath)
			}
		}
	}

	// Apply defaults
	parser := tools.NewConfigParser()
	parser.ApplyDefaults(config)

	// If no lint config was set but we have resources, add default
	if config.ReplLint == nil && (len(config.Charts) > 0 || len(config.Preflights) > 0) {
		config.ReplLint = &tools.ReplLintConfig{
			Version: 1,
			Linters: tools.LintersConfig{},
			Tools:   make(map[string]string),
		}
		parser.ApplyDefaults(config)
	}

	// Write config file
	configPath := filepath.Join(".", ".replicated")
	if err := tools.WriteConfigFile(config, configPath); err != nil {
		return errors.Wrap(err, "writing config file")
	}

	fmt.Fprintf(r.w, "\nCreated %s with:\n", configPath)
	if config.AppSlug != "" {
		fmt.Fprintf(r.w, "  App: %s\n", config.AppSlug)
	} else if config.AppId != "" {
		fmt.Fprintf(r.w, "  App: %s\n", config.AppId)
	}
	if len(config.Charts) > 0 {
		fmt.Fprintf(r.w, "  Charts: %d configured\n", len(config.Charts))
		for _, chart := range config.Charts {
			fmt.Fprintf(r.w, "    - %s\n", chart.Path)
		}
	}
	if len(config.Preflights) > 0 {
		fmt.Fprintf(r.w, "  Preflights: %d configured\n", len(config.Preflights))
		for _, preflight := range config.Preflights {
			if preflight.ChartName != "" && preflight.ChartVersion != "" {
				fmt.Fprintf(r.w, "    - %s (chart: %s:%s)\n", preflight.Path, preflight.ChartName, preflight.ChartVersion)
			} else {
				fmt.Fprintf(r.w, "    - %s\n", preflight.Path)
			}
		}
	}
	if len(config.Manifests) > 0 {
		fmt.Fprintf(r.w, "  Manifests: %d pattern(s) configured\n", len(config.Manifests))
		for _, manifest := range config.Manifests {
			fmt.Fprintf(r.w, "    - %s\n", manifest)
		}
	}
	if config.ReplLint != nil {
		fmt.Fprintf(r.w, "  Linting: Configured\n")
	}

	fmt.Fprintf(r.w, "\nNext steps:\n")
	if len(config.Charts) > 0 || len(config.Preflights) > 0 {
		fmt.Fprintf(r.w, "  Run 'replicated lint' to validate your resources\n")
	}
	fmt.Fprintf(r.w, "  Run 'replicated release create' to create a release\n")

	return nil
}

func (r *runners) promptForChartPaths() ([]tools.ChartConfig, error) {
	var charts []tools.ChartConfig

	for {
		pathPrompt := promptui.Prompt{
			Label:   "Chart path (glob patterns supported, e.g., ./charts/*)",
			Default: "",
		}
		path, err := pathPrompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				return nil, errors.New("cancelled")
			}
			return nil, errors.Wrap(err, "failed to read chart path")
		}
		if path == "" {
			break
		}

		chart := tools.ChartConfig{Path: path}

		// Ask if they want to specify versions (optional)
		addVersions := promptui.Select{
			Label: "Specify chart/app versions? (optional)",
			Items: []string{"No", "Yes"},
		}
		_, addVersionsResult, err := addVersions.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				return nil, errors.New("cancelled")
			}
			return nil, errors.Wrap(err, "failed to read version option")
		}

		if addVersionsResult == "Yes" {
			chartVersionPrompt := promptui.Prompt{
				Label:   "Chart version (optional)",
				Default: "",
			}
			chart.ChartVersion, err = chartVersionPrompt.Run()
			if err != nil {
				if err == promptui.ErrInterrupt {
					return nil, errors.New("cancelled")
				}
				return nil, errors.Wrap(err, "failed to read chart version")
			}

			appVersionPrompt := promptui.Prompt{
				Label:   "App version (optional)",
				Default: "",
			}
			chart.AppVersion, err = appVersionPrompt.Run()
			if err != nil {
				if err == promptui.ErrInterrupt {
					return nil, errors.New("cancelled")
				}
				return nil, errors.Wrap(err, "failed to read app version")
			}
		}

		charts = append(charts, chart)

		addAnother := promptui.Select{
			Label: "Add another chart?",
			Items: []string{"No", "Yes"},
		}
		_, result, err := addAnother.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				return nil, errors.New("cancelled")
			}
			return nil, errors.Wrap(err, "failed to read add another option")
		}

		if result == "No" {
			break
		}
	}

	return charts, nil
}

// promptForChart prompts the user to select a chart for preflight configuration.
// Returns chart name and version to be stored in the config.
// Displays charts as "name:version (from path)" for clarity.
func (r *runners) promptForChart(charts []tools.ChartConfig) (string, string, error) {
	if len(charts) == 0 {
		return "", "", errors.New("no charts available to select from")
	}

	// Build chart selection list with metadata
	type chartOption struct {
		displayName string // For UI: "myapp:1.0.0 (from ./charts/myapp)"
		name        string // For config
		version     string // For config
		path        string // For context
	}

	var options []chartOption
	for _, chart := range charts {
		// Get chart metadata to extract name and version
		metadata, err := lint2.GetChartMetadata(chart.Path)
		if err != nil {
			// Skip charts we can't read
			continue
		}
		options = append(options, chartOption{
			displayName: fmt.Sprintf("%s:%s (from %s)", metadata.Name, metadata.Version, chart.Path),
			name:        metadata.Name,
			version:     metadata.Version,
			path:        chart.Path,
		})
	}

	if len(options) == 0 {
		return "", "", errors.New("no valid charts found (could not read Chart.yaml from any chart)")
	}

	// Extract just display names for prompt
	displayNames := make([]string, len(options))
	for i, opt := range options {
		displayNames[i] = opt.displayName
	}

	prompt := promptui.Select{
		Label: "Select chart for preflight checks",
		Items: displayNames,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrInterrupt {
			return "", "", errors.New("cancelled")
		}
		return "", "", errors.Wrap(err, "failed to select chart")
	}

	selected := options[idx]
	return selected.name, selected.version, nil
}

func (r *runners) promptForPreflightPathsWithCharts(charts []tools.ChartConfig, detectedValuesFiles []string) ([]tools.PreflightConfig, error) {
	var preflights []tools.PreflightConfig
	var sharedChartName, sharedChartVersion string
	var checkedForChart bool

	for {
		pathPrompt := promptui.Prompt{
			Label:   "Preflight spec path (e.g., ./preflight.yaml)",
			Default: "",
		}
		path, err := pathPrompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				return nil, errors.New("cancelled")
			}
			return nil, errors.Wrap(err, "failed to read preflight path")
		}
		if path == "" {
			break
		}

		preflight := tools.PreflightConfig{Path: path}

		// If we haven't prompted for chart yet and charts are available, prompt now
		if !checkedForChart && len(charts) > 0 {
			chartName, chartVersion, err := r.promptForChart(charts)
			if err != nil {
				return nil, err
			}
			sharedChartName = chartName
			sharedChartVersion = chartVersion
			checkedForChart = true
		}

		// Apply shared chart selection
		preflight.ChartName = sharedChartName
		preflight.ChartVersion = sharedChartVersion

		preflights = append(preflights, preflight)

		addAnother := promptui.Select{
			Label: "Add another preflight spec?",
			Items: []string{"No", "Yes"},
		}
		_, result, err := addAnother.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				return nil, errors.New("cancelled")
			}
			return nil, errors.Wrap(err, "failed to read add another option")
		}

		if result == "No" {
			break
		}
	}

	return preflights, nil
}

func (r *runners) promptForPreflightValues(preflightPath string, charts []tools.ChartConfig) (string, error) {
	if len(charts) == 0 {
		// No charts configured, just ask if they want to specify a custom path
		addValuesPath := promptui.Select{
			Label: fmt.Sprintf("Does '%s' use Helm chart values?", filepath.Base(preflightPath)),
			Items: []string{"No", "Yes - specify path"},
		}
		_, result, err := addValuesPath.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				return "", errors.New("cancelled")
			}
			return "", errors.Wrap(err, "failed to read values option")
		}

		if result == "Yes - specify path" {
			valuesPathPrompt := promptui.Prompt{
				Label:   "Values file path",
				Default: "",
			}
			valuesPath, err := valuesPathPrompt.Run()
			if err != nil {
				if err == promptui.ErrInterrupt {
					return "", errors.New("cancelled")
				}
				return "", errors.Wrap(err, "failed to read values path")
			}
			return valuesPath, nil
		}
		return "", nil
	}

	// Charts are configured, offer them as options
	options := []string{"No"}
	for _, chart := range charts {
		options = append(options, fmt.Sprintf("Yes - use %s", chart.Path))
	}
	options = append(options, "Yes - other path")

	valuesPrompt := promptui.Select{
		Label: fmt.Sprintf("Does '%s' use Helm chart values?", filepath.Base(preflightPath)),
		Items: options,
	}
	_, result, err := valuesPrompt.Run()
	if err != nil {
		if err == promptui.ErrInterrupt {
			return "", errors.New("cancelled")
		}
		return "", errors.Wrap(err, "failed to read values option")
	}

	if result == "No" {
		return "", nil
	}

	if result == "Yes - other path" {
		valuesPathPrompt := promptui.Prompt{
			Label:   "Values file path",
			Default: "",
		}
		valuesPath, err := valuesPathPrompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				return "", errors.New("cancelled")
			}
			return "", errors.Wrap(err, "failed to read values path")
		}
		return valuesPath, nil
	}

	// Extract the chart path from the selection
	for _, chart := range charts {
		if result == fmt.Sprintf("Yes - use %s", chart.Path) {
			return chart.Path, nil
		}
	}

	return "", nil
}

func (r *runners) promptForManifests() ([]string, error) {
	var manifests []string

	for {
		manifestPrompt := promptui.Prompt{
			Label:   "Manifest path (glob patterns supported, e.g., ./manifests/*.yaml)",
			Default: "",
		}
		path, err := manifestPrompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				return nil, errors.New("cancelled")
			}
			return nil, errors.Wrap(err, "failed to read manifest path")
		}
		if path == "" {
			break
		}

		manifests = append(manifests, path)

		addAnother := promptui.Select{
			Label: "Add another manifest pattern?",
			Items: []string{"No", "Yes"},
		}
		_, result, err := addAnother.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				return nil, errors.New("cancelled")
			}
			return nil, errors.Wrap(err, "failed to read add another option")
		}

		if result == "No" {
			break
		}
	}

	return manifests, nil
}

func (r *runners) promptForLintConfig(hasCharts, hasPreflights bool) (*tools.ReplLintConfig, error) {
	config := &tools.ReplLintConfig{
		Version: 1,
		Linters: tools.LintersConfig{},
		Tools:   make(map[string]string),
	}

	// Prompt for relevant linters based on what resources are configured
	if hasCharts {
		enableHelm := promptui.Select{
			Label: "Enable Helm linting?",
			Items: []string{"Yes", "No"},
		}
		_, result, err := enableHelm.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				return nil, errors.New("cancelled")
			}
			return nil, errors.Wrap(err, "failed to read helm linting option")
		}

		disabled := result == "No"
		config.Linters.Helm.Disabled = &disabled
	}

	if hasPreflights {
		enablePreflight := promptui.Select{
			Label: "Enable preflight linting?",
			Items: []string{"Yes", "No"},
		}
		_, result, err := enablePreflight.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				return nil, errors.New("cancelled")
			}
			return nil, errors.Wrap(err, "failed to read preflight linting option")
		}

		disabled := result == "No"
		config.Linters.Preflight.Disabled = &disabled
	}

	// Support bundle linting (common for troubleshooting)
	enableSupportBundle := promptui.Select{
		Label: "Enable support bundle linting?",
		Items: []string{"Yes", "No"},
	}
	_, sbResult, err := enableSupportBundle.Run()
	if err != nil {
		if err == promptui.ErrInterrupt {
			return nil, errors.New("cancelled")
		}
		return nil, errors.Wrap(err, "failed to read support bundle linting option")
	}

	sbDisabled := sbResult == "No"
	config.Linters.SupportBundle.Disabled = &sbDisabled

	return config, nil
}

func (r *runners) promptForAppSelection(ctx context.Context) (string, error) {
	// Fetch apps from API
	fmt.Fprintf(r.w, "Fetching apps from your account...\n")
	r.w.Flush()

	kotsApps, err := r.kotsAPI.ListApps(ctx, false)
	if err != nil {
		return "", errors.Wrap(err, "failed to list apps")
	}

	if len(kotsApps) == 0 {
		fmt.Fprintf(r.w, "No apps found in your account.\n\n")
		return "", nil
	}

	// Build list of app display names
	type appChoice struct {
		label string
		slug  string
	}

	choices := []appChoice{}
	choices = append(choices, appChoice{label: "Skip (enter manually)", slug: ""})

	for _, app := range kotsApps {
		label := fmt.Sprintf("%s (%s)", app.App.Name, app.App.Slug)
		choices = append(choices, appChoice{label: label, slug: app.App.Slug})
	}

	// Create list of labels for promptui
	labels := make([]string, len(choices))
	for i, choice := range choices {
		labels[i] = choice.label
	}

	prompt := promptui.Select{
		Label: "Select an app",
		Items: labels,
		Size:  10,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrInterrupt {
			return "", errors.New("cancelled")
		}
		return "", errors.Wrap(err, "failed to select app")
	}

	fmt.Fprintf(r.w, "\n")
	return choices[idx].slug, nil
}
