package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/imageextract"
	"github.com/replicatedhq/replicated/pkg/lint2"
	"github.com/replicatedhq/replicated/pkg/tools"
	"github.com/replicatedhq/replicated/pkg/version"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// InitLint is removed - the standalone "replicated lint" command has been removed.
// The linting functionality is now available via "replicated release lint" with the
// release-validation-v2 feature flag. The runLint function below is still used
// internally by the release lint command.

// getToolVersion returns tool version from config or "latest" as fallback.
func getToolVersion(config *tools.Config, tool string) string {
	if config.ReplLint.Tools != nil {
		if v, ok := config.ReplLint.Tools[tool]; ok {
			return v
		}
	}
	return "latest"
}

// resolveToolVersion gets tool version from config, resolving "latest" to actual version or falling back to default
func resolveToolVersion(ctx context.Context, config *tools.Config, resolver *tools.Resolver, toolName, defaultVersion string) string {
	// Get version from config
	version := "latest"
	if config.ReplLint.Tools != nil {
		if v, ok := config.ReplLint.Tools[toolName]; ok {
			version = v
		}
	}

	// Resolve "latest" to actual version
	if version == "latest" || version == "" {
		if resolved, err := resolver.ResolveLatestVersion(ctx, toolName); err == nil {
			return resolved
		}
		return defaultVersion // Fallback
	}

	return version
}

// extractAllPathsAndMetadata extracts all paths and metadata for linting, optionally including ChartsWithMetadata if verbose
func extractAllPathsAndMetadata(ctx context.Context, config *tools.Config, verbose bool, helmVersion, preflightVersion, sbVersion, ecVersion, kotsVersion string) (*ExtractedPaths, error) {
	result := &ExtractedPaths{
		HelmVersion:      helmVersion,
		PreflightVersion: preflightVersion,
		SBVersion:        sbVersion,
		ECVersion:        ecVersion,
		KotsVersion:      kotsVersion,
	}

	// Extract chart paths (for Helm linting)
	if len(config.Charts) > 0 {
		chartPaths, err := lint2.GetChartPathsFromConfig(config)
		if err != nil {
			return nil, err
		}
		result.ChartPaths = chartPaths
	}

	// Extract preflight paths with values
	if len(config.Preflights) > 0 {
		preflights, err := lint2.GetPreflightWithValuesFromConfig(config)
		if err != nil {
			return nil, err
		}
		result.Preflights = preflights
	}

	// Discover HelmChart manifests ONCE (used by preflight rendering, support bundle analysis, image extraction, validation)
	if len(config.Manifests) > 0 {
		helmChartManifests, err := lint2.DiscoverHelmChartManifests(config.Manifests)
		if err != nil {
			return nil, fmt.Errorf("failed to discover HelmChart manifests: %w", err)
		}
		result.HelmChartManifests = helmChartManifests
	} else {
		// No manifests configured - return empty map (validation will check if needed)
		result.HelmChartManifests = make(map[string]*lint2.HelmChartManifest)
	}

	// Discover support bundles
	if len(config.Manifests) > 0 {
		sbPaths, err := lint2.DiscoverSupportBundlesFromManifests(config.Manifests)
		if err != nil {
			return nil, err
		}
		result.SupportBundles = sbPaths
	}

	// Discover embedded cluster configs
	if len(config.Manifests) > 0 {
		ecPaths := []string{}
		for _, pattern := range config.Manifests {
			paths, err := lint2.DiscoverEmbeddedClusterPaths(pattern)
			if err != nil {
				return nil, fmt.Errorf("failed to discover embedded cluster configs from pattern %s: %w", pattern, err)
			}
			ecPaths = append(ecPaths, paths...)
		}
		result.EmbeddedClusterPaths = ecPaths
	}

	// Discover KOTS manifests
	if len(config.Manifests) > 0 {
		kotsPaths := []string{}
		for _, pattern := range config.Manifests {
			paths, err := lint2.DiscoverKotsPaths(pattern)
			if err != nil {
				return nil, fmt.Errorf("failed to discover KOTS manifests from pattern %s: %w", pattern, err)
			}
			kotsPaths = append(kotsPaths, paths...)
		}
		result.KotsPaths = kotsPaths
	}

	// Extract charts with metadata (needed for validation and image extraction)
	if len(config.Charts) > 0 {
		chartsWithMetadata, err := lint2.GetChartsWithMetadataFromConfig(config)
		if err != nil {
			return nil, err
		}
		result.ChartsWithMetadata = chartsWithMetadata
	}

	return result, nil
}

// resolveAppContext determines app ID using priority: --app-id > --app-slug > config appId > config appSlug > API fetch
// Returns empty string if no app available (not an error), except when user explicitly specified an app that cannot be resolved
func (r *runners) resolveAppContext(ctx context.Context, config *tools.Config) (string, error) {
	selectedApp := ""

	// 1) Highest priority: appID from runners (--app-id flag or env var)
	if r.appID != "" {
		selectedApp = r.appID
	} else if r.appSlug != "" { // 2) appSlug from runners (--app-slug flag)
		selectedApp = r.appSlug
	} else if config.AppId != "" { // 3) .replicated config: appId
		selectedApp = config.AppId
	} else if config.AppSlug != "" { // 4) .replicated config: appSlug
		selectedApp = config.AppSlug
	} else {
		// 5) Fetch apps for current profile; pick one or prompt user if multiple
		// Check if API client is available (may be nil in tests)
		if r.kotsAPI == nil {
			return "", errors.New("API client not initialized")
		}

		apps, err := r.kotsAPI.ListApps(ctx, false)
		if err != nil {
			// API failure: return empty string (graceful degradation)
			return "", nil
		}

		if len(apps) == 0 {
			// No apps available: return empty string (not an error)
			return "", nil
		}

		if len(apps) == 1 {
			// Single app: auto-select
			if apps[0].App != nil && apps[0].App.ID != "" {
				selectedApp = apps[0].App.ID
			} else {
				selectedApp = apps[0].App.Slug
			}
		} else {
			// Multiple apps available - prompt user (interactive mode only)
			// Check if we're in a TTY (required for interactive prompts)
			if !term.IsTerminal(int(os.Stdin.Fd())) {
				return "", errors.New("multiple apps available for the current credentials; specify --app flag or set appId/appSlug in .replicated config")
			}

			items := make([]string, 0, len(apps))
			for _, a := range apps {
				if a.App == nil {
					continue
				}
				items = append(items, fmt.Sprintf("%s (%s / %s)", a.App.Name, a.App.Slug, a.App.ID))
			}
			prompt := promptui.Select{
				Label: "Select app for linting",
				Items: items,
			}
			idx, _, perr := prompt.Run()
			if perr != nil {
				return "", errors.Wrap(perr, "app selection canceled")
			}
			if apps[idx].App != nil && apps[idx].App.ID != "" {
				selectedApp = apps[idx].App.ID
			} else {
				selectedApp = apps[idx].App.Slug
			}
		}
	}

	// If no app selected, return empty string (graceful)
	if selectedApp == "" {
		return "", nil
	}

	// Resolve to canonical app ID (handles either slug or id input)
	appObj, err := r.kotsAPI.GetApp(ctx, selectedApp, true)
	if err != nil {
		return "", errors.Wrapf(err, "failed to resolve app from selection: %s", selectedApp)
	}
	if appObj == nil || appObj.ID == "" {
		return "", errors.Errorf("failed to resolve app id from selection: %s (app not found or invalid)", selectedApp)
	}

	return appObj.ID, nil
}

func (r *runners) runLint(cmd *cobra.Command, args []string) error {
	// Validate output format
	if r.outputFormat != "table" && r.outputFormat != "json" {
		return errors.Errorf("invalid output: %s. Supported output formats: json, table", r.outputFormat)
	}

	// Load .replicated config using tools parser (supports monorepos)
	parser := tools.NewConfigParser()
	config, err := parser.FindAndParseConfig(".")

	if err != nil {
		return errors.Wrap(err, "failed to load .replicated config")
	}

	// Resolve app context for all linters (not just embedded-cluster)
	// This makes REPLICATED_APP, REPLICATED_API_TOKEN, and REPLICATED_API_ORIGIN
	// available to all linter binaries for future vendor portal integrations.
	//
	// If app resolution fails, we store the error and continue with other linters.
	// Only the embedded-cluster linter will fail if app context is required but unavailable.
	appID, appResolveErr := r.resolveAppContext(cmd.Context(), config)

	// Warn user if app resolution failed and EC linter is enabled (table mode only)
	if appResolveErr != nil && config.ReplLint.Linters.EmbeddedCluster.IsEnabled() && r.outputFormat == "table" {
		fmt.Fprintf(r.w, "⚠️  Warning: Could not resolve app context: %v\n", appResolveErr)
		fmt.Fprintf(r.w, "   Embedded-cluster linter will fail. Other linters will continue.\n\n")
		r.w.Flush()
	}

	// Set REPLICATED_APP env var if we successfully resolved an app
	// Note: REPLICATED_API_TOKEN and REPLICATED_API_ORIGIN are already set in root.go
	if appID != "" {
		_ = os.Setenv("REPLICATED_APP", appID)
	}

	// Initialize JSON output structure
	output := &JSONLintOutput{}

	// Resolve all tool versions (including "latest" to actual versions)
	resolver := tools.NewResolver()
	helmVersion := resolveToolVersion(cmd.Context(), config, resolver, tools.ToolHelm, tools.DefaultHelmVersion)
	preflightVersion := resolveToolVersion(cmd.Context(), config, resolver, tools.ToolPreflight, tools.DefaultPreflightVersion)
	supportBundleVersion := resolveToolVersion(cmd.Context(), config, resolver, tools.ToolSupportBundle, tools.DefaultSupportBundleVersion)
	embeddedClusterVersion := resolveToolVersion(cmd.Context(), config, resolver, tools.ToolEmbeddedCluster, "latest")
	kotsVersion := resolveToolVersion(cmd.Context(), config, resolver, tools.ToolKots, "latest")

	// Populate metadata with all resolved versions
	configPath := findConfigFilePath(".")
	output.Metadata = newLintMetadata(configPath, helmVersion, preflightVersion, supportBundleVersion, embeddedClusterVersion, kotsVersion, version.Version())

	// Check if we're in auto-discovery mode (no charts/preflights/manifests configured)
	autoDiscoveryMode := len(config.Charts) == 0 && len(config.Preflights) == 0 && len(config.Manifests) == 0

	if autoDiscoveryMode {
		fmt.Fprintf(r.w, "No .replicated config found. Auto-discovering lintable resources in current directory...\n\n")
		r.w.Flush()

		// Auto-discover Helm charts (for counting and display)
		chartPaths, err := lint2.DiscoverChartPaths(filepath.Join(".", "**"))
		if err != nil {
			return errors.Wrap(err, "failed to discover helm charts")
		}

		// Auto-discover Preflight specs (for counting and display)
		preflightPaths, err := lint2.DiscoverPreflightPaths(filepath.Join(".", "**"))
		if err != nil {
			return errors.Wrap(err, "failed to discover preflight specs")
		}

		// Auto-discover Support Bundle specs (for counting and display)
		sbPaths, err := lint2.DiscoverSupportBundlePaths(filepath.Join(".", "**"))
		if err != nil {
			return errors.Wrap(err, "failed to discover support bundle specs")
		}

		// Auto-discover HelmChart manifests (for counting and display)
		helmChartPaths, err := lint2.DiscoverHelmChartPaths(filepath.Join(".", "**"))
		if err != nil {
			return errors.Wrap(err, "failed to discover HelmChart manifests")
		}

		// Auto-discover Embedded Cluster configs (for counting and display)
		ecPaths, err := lint2.DiscoverEmbeddedClusterPaths(filepath.Join(".", "**"))
		if err != nil {
			return errors.Wrap(err, "failed to discover embedded cluster configs")
		}

		// Auto-discover KOTS manifests (for counting and display)
		kotsPaths, err := lint2.DiscoverKotsPaths(filepath.Join(".", "**"))
		if err != nil {
			return errors.Wrap(err, "failed to discover KOTS manifests")
		}

		// Store glob patterns (not explicit paths) for extraction phase
		// This matches non-autodiscovery behavior and uses lenient filtering
		if len(chartPaths) > 0 {
			config.Charts = []tools.ChartConfig{{Path: "./**"}}
		}
		if len(preflightPaths) > 0 {
			config.Preflights = []tools.PreflightConfig{
				{Path: "./**"},
			}
		}
		// Support Bundles, HelmChart manifests, Embedded Cluster configs, and KOTS manifests go into config.Manifests
		if len(sbPaths) > 0 || len(helmChartPaths) > 0 || len(ecPaths) > 0 || len(kotsPaths) > 0 {
			config.Manifests = []string{"./**"}
		}

		// Print what was discovered
		fmt.Fprintf(r.w, "Discovered resources:\n")
		fmt.Fprintf(r.w, "  - %d Helm chart(s)\n", len(chartPaths))
		fmt.Fprintf(r.w, "  - %d Preflight spec(s)\n", len(preflightPaths))
		fmt.Fprintf(r.w, "  - %d Support Bundle spec(s)\n", len(sbPaths))
		fmt.Fprintf(r.w, "  - %d HelmChart manifest(s)\n", len(helmChartPaths))
		fmt.Fprintf(r.w, "  - %d Embedded Cluster config(s)\n", len(ecPaths))
		fmt.Fprintf(r.w, "  - %d KOTS manifest(s)\n\n", len(kotsPaths))
		r.w.Flush()

		// If nothing was found, exit early
		if len(chartPaths) == 0 && len(preflightPaths) == 0 && len(sbPaths) == 0 && len(ecPaths) == 0 && len(kotsPaths) == 0 {
			fmt.Fprintf(r.w, "No lintable resources found in current directory.\n")
			r.w.Flush()
			return nil
		}
	}

	// Display tool versions if verbose mode is enabled
	if r.args.lintVerbose {
		fmt.Fprintln(r.w, "Tool versions:")

		// Display already resolved versions
		fmt.Fprintf(r.w, "  Helm: %s\n", helmVersion)
		fmt.Fprintf(r.w, "  Preflight: %s\n", preflightVersion)
		fmt.Fprintf(r.w, "  Support Bundle: %s\n", supportBundleVersion)
		fmt.Fprintf(r.w, "  Embedded Cluster: %s\n", embeddedClusterVersion)
		fmt.Fprintf(r.w, "  KOTS: %s\n", kotsVersion)

		fmt.Fprintln(r.w)
		r.w.Flush()
	}

	// Extract all paths and metadata once (consolidates extraction logic across linters)
	extracted, err := extractAllPathsAndMetadata(cmd.Context(), config, r.args.lintVerbose, helmVersion, preflightVersion, supportBundleVersion, embeddedClusterVersion, kotsVersion)
	if err != nil {
		return errors.Wrap(err, "failed to extract paths and metadata")
	}

	// Validate chart-to-HelmChart mapping if charts are configured
	if len(config.Charts) > 0 {
		// Charts configured but no manifests - error early
		if len(config.Manifests) == 0 {
			return errors.New("charts are configured but no manifests paths provided\n\n" +
				"HelmChart manifests (kind: HelmChart) are required for each chart.\n" +
				"Add manifest paths to your .replicated config:\n\n" +
				"manifests:\n" +
				"  - \"./manifests/**/*.yaml\"")
		}

		// Validate mapping using already-extracted metadata
		validationResult, err := lint2.ValidateChartToHelmChartMapping(
			extracted.ChartsWithMetadata, // Already populated in extraction
			extracted.HelmChartManifests,
		)
		if err != nil {
			// Hard error - stop before linting
			return errors.Wrap(err, "chart validation failed")
		}

		// Display warnings (orphaned HelmChart manifests)
		if r.outputFormat == "table" && len(validationResult.Warnings) > 0 {
			for _, warning := range validationResult.Warnings {
				fmt.Fprintf(r.w, "Warning: %s\n", warning)
			}
			fmt.Fprintln(r.w)
			r.w.Flush()
		}
	}

	// Extract and display images if verbose mode is enabled
	if r.args.lintVerbose && len(extracted.ChartsWithMetadata) > 0 {
		imageResults, err := r.extractImagesFromCharts(cmd.Context(), extracted.ChartsWithMetadata, extracted.HelmChartManifests)
		if err != nil {
			// Log warning but don't fail the lint command
			if r.outputFormat == "table" {
				fmt.Fprintf(r.w, "Warning: Failed to extract images: %v\n\n", err)
				r.w.Flush()
			}
		} else {
			output.Images = imageResults
			// Display images (only for table format)
			if r.outputFormat == "table" {
				r.displayImages(imageResults)

				// Print separator
				fmt.Fprintln(r.w, "────────────────────────────────────────────────────────────────────────────")
				fmt.Fprintln(r.w, "\nRunning lint checks...")
				fmt.Fprintln(r.w)
				r.w.Flush()
			}
		}
	}

	// Lint Helm charts if enabled
	if config.ReplLint.Linters.Helm.IsEnabled() {
		if len(extracted.ChartPaths) == 0 {
			output.HelmResults = &HelmLintResults{Enabled: true, Charts: []ChartLintResult{}}
			if r.outputFormat == "table" {
				fmt.Fprintf(r.w, "No Helm charts configured (skipping Helm linting)\n\n")
			}
		} else {
			helmResults, err := r.lintHelmCharts(cmd, extracted.ChartPaths, extracted.HelmVersion)
			if err != nil {
				return err
			}
			output.HelmResults = helmResults
		}
	} else {
		output.HelmResults = &HelmLintResults{Enabled: false, Charts: []ChartLintResult{}}
		if r.outputFormat == "table" {
			fmt.Fprintf(r.w, "Helm linting is disabled in .replicated config\n\n")
		}
	}

	// Lint Preflight specs if enabled
	if config.ReplLint.Linters.Preflight.IsEnabled() {
		if len(extracted.Preflights) == 0 {
			output.PreflightResults = &PreflightLintResults{Enabled: true, Specs: []PreflightLintResult{}}
			if r.outputFormat == "table" {
				fmt.Fprintf(r.w, "No preflight specs configured (skipping preflight linting)\n\n")
			}
		} else {
			preflightResults, err := r.lintPreflightSpecs(cmd, extracted.Preflights, extracted.HelmChartManifests, extracted.PreflightVersion)
			if err != nil {
				return err
			}
			output.PreflightResults = preflightResults
		}
	} else {
		output.PreflightResults = &PreflightLintResults{Enabled: false, Specs: []PreflightLintResult{}}
		if r.outputFormat == "table" {
			fmt.Fprintf(r.w, "Preflight linting is disabled in .replicated config\n\n")
		}
	}

	// Lint Support Bundle specs if enabled
	if config.ReplLint.Linters.SupportBundle.IsEnabled() {
		sbResults, err := r.lintSupportBundleSpecs(cmd, extracted.SupportBundles, extracted.SBVersion)
		if err != nil {
			return err
		}
		output.SupportBundleResults = sbResults
	} else {
		output.SupportBundleResults = &SupportBundleLintResults{Enabled: false, Specs: []SupportBundleLintResult{}}
		if r.outputFormat == "table" {
			fmt.Fprintf(r.w, "Support Bundle linting is disabled in .replicated config\n\n")
		}
	}

	// Lint Embedded Cluster configs if enabled
	if config.ReplLint.Linters.EmbeddedCluster.IsEnabled() {
		if len(extracted.EmbeddedClusterPaths) == 0 {
			output.EmbeddedClusterResults = &EmbeddedClusterLintResults{Enabled: true, Configs: []EmbeddedClusterLintResult{}}
			if r.outputFormat == "table" {
				fmt.Fprintf(r.w, "No embedded cluster configs configured (skipping EC linting)\n\n")
			}
		} else {
			// Pass app resolution error - EC linter will fail gracefully if app context unavailable
			ecResults, err := r.lintEmbeddedClusterConfigs(cmd, extracted.EmbeddedClusterPaths, extracted.ECVersion, appResolveErr)
			if err != nil {
				return err
			}
			output.EmbeddedClusterResults = ecResults
		}
	} else {
		output.EmbeddedClusterResults = &EmbeddedClusterLintResults{Enabled: false, Configs: []EmbeddedClusterLintResult{}}
		if r.outputFormat == "table" {
			fmt.Fprintf(r.w, "Embedded Cluster linting is disabled in .replicated config\n\n")
		}
	}

	// Lint KOTS manifests if enabled
	if config.ReplLint.Linters.Kots.IsEnabled() {
		if len(extracted.KotsPaths) == 0 {
			output.KotsResults = &KotsLintResults{Enabled: true, Manifests: []KotsLintResult{}}
			if r.outputFormat == "table" {
				fmt.Fprintf(r.w, "No KOTS manifests configured (skipping KOTS linting)\n\n")
			}
		} else {
			kotsResults, err := r.lintKotsManifests(cmd, extracted.KotsPaths, extracted.KotsVersion)
			if err != nil {
				return err
			}
			output.KotsResults = kotsResults
		}
	} else {
		output.KotsResults = &KotsLintResults{Enabled: false, Manifests: []KotsLintResult{}}
		if r.outputFormat == "table" {
			fmt.Fprintf(r.w, "KOTS linting is disabled in .replicated config\n\n")
		}
	}

	// Calculate overall summary
	output.Summary = r.calculateOverallSummary(output)

	// Output to stdout
	if r.outputFormat == "json" {
		if err := print.LintResults(r.outputFormat, r.w, output); err != nil {
			return errors.Wrap(err, "failed to print JSON output to stdout")
		}
	} else {
		// Table format was already displayed by individual display functions
		// Just flush the writer
		if err := r.w.Flush(); err != nil {
			return errors.Wrap(err, "failed to flush output")
		}
	}

	// Return error if any linting failed
	if !output.Summary.OverallSuccess {
		return errors.New("linting failed")
	}

	return nil
}

func (r *runners) lintHelmCharts(cmd *cobra.Command, chartPaths []string, helmVersion string) (*HelmLintResults, error) {
	results := &HelmLintResults{
		Enabled: true,
		Charts:  make([]ChartLintResult, 0, len(chartPaths)),
	}

	// Lint all charts and collect results
	for _, chartPath := range chartPaths {
		lint2Result, err := lint2.LintChart(cmd.Context(), chartPath, helmVersion)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to lint chart: %s", chartPath)
		}

		// Convert to structured format
		chartResult := ChartLintResult{
			Path:     chartPath,
			Success:  lint2Result.Success,
			Messages: convertLint2Messages(lint2Result.Messages),
			Summary:  calculateResourceSummary(lint2Result.Messages),
		}
		results.Charts = append(results.Charts, chartResult)
	}

	// Display results in table format (only if table output)
	if r.outputFormat == "table" {
		// Convert to []LintableResult for generic display
		lintableResults := make([]LintableResult, len(results.Charts))
		for i, chart := range results.Charts {
			lintableResults[i] = chart
		}
		if err := r.displayLintResults("HELM CHARTS", "chart", "charts", lintableResults); err != nil {
			return nil, errors.Wrap(err, "failed to display helm results")
		}
	}

	return results, nil
}

func (r *runners) lintPreflightSpecs(cmd *cobra.Command, preflights []lint2.PreflightWithValues, helmChartManifests map[string]*lint2.HelmChartManifest, preflightVersion string) (*PreflightLintResults, error) {
	results := &PreflightLintResults{
		Enabled: true,
		Specs:   make([]PreflightLintResult, 0, len(preflights)),
	}

	// Lint all preflight specs and collect results
	for _, pf := range preflights {
		lint2Result, err := lint2.LintPreflight(
			cmd.Context(),
			pf.SpecPath,
			pf.ValuesPath,
			pf.ChartName,
			pf.ChartVersion,
			helmChartManifests,
			preflightVersion,
		)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to lint preflight spec: %s", pf.SpecPath)
		}

		// Convert to structured format
		preflightResult := PreflightLintResult{
			Path:     pf.SpecPath,
			Success:  lint2Result.Success,
			Messages: convertLint2Messages(lint2Result.Messages),
			Summary:  calculateResourceSummary(lint2Result.Messages),
		}
		results.Specs = append(results.Specs, preflightResult)
	}

	// Display results in table format (only if table output)
	if r.outputFormat == "table" {
		// Convert to []LintableResult for generic display
		lintableResults := make([]LintableResult, len(results.Specs))
		for i, spec := range results.Specs {
			lintableResults[i] = spec
		}
		if err := r.displayLintResults("PREFLIGHT CHECKS", "preflight spec", "preflight specs", lintableResults); err != nil {
			return nil, errors.Wrap(err, "failed to display preflight results")
		}
	}

	return results, nil
}

func (r *runners) lintSupportBundleSpecs(cmd *cobra.Command, sbPaths []string, sbVersion string) (*SupportBundleLintResults, error) {
	results := &SupportBundleLintResults{
		Enabled: true,
		Specs:   make([]SupportBundleLintResult, 0, len(sbPaths)),
	}

	// If no support bundles found, that's not an error - they're optional
	if len(sbPaths) == 0 {
		return results, nil
	}

	// Lint all support bundle specs and collect results
	for _, specPath := range sbPaths {
		lint2Result, err := lint2.LintSupportBundle(cmd.Context(), specPath, sbVersion)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to lint support bundle spec: %s", specPath)
		}

		// Convert to structured format
		sbResult := SupportBundleLintResult{
			Path:     specPath,
			Success:  lint2Result.Success,
			Messages: convertLint2Messages(lint2Result.Messages),
			Summary:  calculateResourceSummary(lint2Result.Messages),
		}
		results.Specs = append(results.Specs, sbResult)
	}

	// Display results in table format (only if table output)
	if r.outputFormat == "table" {
		// Convert to []LintableResult for generic display
		lintableResults := make([]LintableResult, len(results.Specs))
		for i, spec := range results.Specs {
			lintableResults[i] = spec
		}
		if err := r.displayLintResults("SUPPORT BUNDLES", "support bundle spec", "support bundle specs", lintableResults); err != nil {
			return nil, errors.Wrap(err, "failed to display support bundle results")
		}
	}

	return results, nil
}

// validateSingleConfigLimit enforces 0 or 1 config limit for EC and KOTS linters
// Returns failed results for each path if multiple configs found
func validateSingleConfigLimit[T LintableResult](
	r *runners,
	paths []string,
	linterName string,
	singularName string,
	pluralName string,
	resultConstructor func(path string, errorMsg LintMessage) T,
) ([]T, bool) {
	// No error if 0 or 1 config
	if len(paths) <= 1 {
		return nil, false
	}

	// Create error message
	errorMsg := LintMessage{
		Severity: "ERROR",
		Message: fmt.Sprintf(
			"Multiple %s configs found (%d). Only 0 or 1 config per project is supported. Found configs: %s",
			linterName, len(paths), strings.Join(paths, ", "),
		),
	}

	// Build failed results for each path
	results := make([]T, 0, len(paths))
	for _, path := range paths {
		results = append(results, resultConstructor(path, errorMsg))
	}

	// Display if table format
	if r.outputFormat == "table" {
		lintableResults := make([]LintableResult, len(results))
		for i, result := range results {
			lintableResults[i] = result
		}
		if err := r.displayLintResults(linterName, singularName, pluralName, lintableResults); err != nil {
			// Non-fatal display error - log but continue
			fmt.Fprintf(r.w, "Warning: failed to display results: %v\n", err)
		}
	}

	return results, true
}

func (r *runners) lintEmbeddedClusterConfigs(cmd *cobra.Command, ecPaths []string, ecVersion string, appResolveErr error) (*EmbeddedClusterLintResults, error) {
	results := &EmbeddedClusterLintResults{
		Enabled: true,
		Configs: make([]EmbeddedClusterLintResult, 0, len(ecPaths)),
	}

	// If no embedded cluster configs found, that's not an error - they're optional
	if len(ecPaths) == 0 {
		return results, nil
	}

	// If app resolution failed, embedded-cluster linter cannot run
	// Return a failed result for each config instead of stopping the entire lint command
	if appResolveErr != nil {
		for _, configPath := range ecPaths {
			results.Configs = append(results.Configs, EmbeddedClusterLintResult{
				Path:    configPath,
				Success: false,
				Messages: []LintMessage{
					{
						Severity: "ERROR",
						Message:  fmt.Sprintf("Embedded cluster linting requires app context: %v", appResolveErr),
					},
				},
				Summary: ResourceSummary{
					ErrorCount:   1,
					WarningCount: 0,
					InfoCount:    0,
				},
			})
		}

		// Display results in table format
		if r.outputFormat == "table" {
			lintableResults := make([]LintableResult, len(results.Configs))
			for i, config := range results.Configs {
				lintableResults[i] = config
			}
			if err := r.displayLintResults("EMBEDDED CLUSTER", "embedded cluster config", "embedded cluster configs", lintableResults); err != nil {
				return nil, errors.Wrap(err, "failed to display embedded cluster results")
			}
		}

		return results, nil
	}

	// Validate: only 0 or 1 config allowed
	// If 2+ found, return failed results but don't block other linters
	if failedResults, hasError := validateSingleConfigLimit(
		r,
		ecPaths,
		"EMBEDDED CLUSTER",
		"embedded cluster config",
		"embedded cluster configs",
		func(path string, errorMsg LintMessage) EmbeddedClusterLintResult {
			return EmbeddedClusterLintResult{
				Path:    path,
				Success: false,
				Messages: []LintMessage{errorMsg},
				Summary: ResourceSummary{
					ErrorCount:   1,
					WarningCount: 0,
					InfoCount:    0,
				},
			}
		},
	); hasError {
		for _, result := range failedResults {
			results.Configs = append(results.Configs, result)
		}
		return results, nil
	}

	// Lint all embedded cluster configs and collect results
	for _, configPath := range ecPaths {
		lint2Result, err := lint2.LintEmbeddedCluster(cmd.Context(), configPath, ecVersion)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to lint embedded cluster config: %s", configPath)
		}

		// Convert to structured format
		ecResult := EmbeddedClusterLintResult{
			Path:     configPath,
			Success:  lint2Result.Success,
			Messages: convertLint2Messages(lint2Result.Messages),
			Summary:  calculateResourceSummary(lint2Result.Messages),
		}
		results.Configs = append(results.Configs, ecResult)
	}

	// Display results in table format (only if table output)
	if r.outputFormat == "table" {
		// Convert to []LintableResult for generic display
		lintableResults := make([]LintableResult, len(results.Configs))
		for i, config := range results.Configs {
			lintableResults[i] = config
		}
		if err := r.displayLintResults("EMBEDDED CLUSTER", "embedded cluster config", "embedded cluster configs", lintableResults); err != nil {
			return nil, errors.Wrap(err, "failed to display embedded cluster results")
		}
	}

	return results, nil
}

func (r *runners) lintKotsManifests(cmd *cobra.Command, kotsPaths []string, kotsVersion string) (*KotsLintResults, error) {
	results := &KotsLintResults{
		Enabled:   true,
		Manifests: make([]KotsLintResult, 0, len(kotsPaths)),
	}

	// If no KOTS manifests found, that's not an error - they're optional
	if len(kotsPaths) == 0 {
		return results, nil
	}

	// Validate: Only 0 or 1 KOTS Config allowed per project (like EC)
	// If 2+ found, return failed results but don't block other linters
	if failedResults, hasError := validateSingleConfigLimit(
		r,
		kotsPaths,
		"KOTS",
		"KOTS manifest",
		"KOTS manifests",
		func(path string, errorMsg LintMessage) KotsLintResult {
			return KotsLintResult{
				Path:    path,
				Success: false,
				Messages: []LintMessage{errorMsg},
				Summary: ResourceSummary{
					ErrorCount:   1,
					WarningCount: 0,
					InfoCount:    0,
				},
			}
		},
	); hasError {
		for _, result := range failedResults {
			results.Manifests = append(results.Manifests, result)
		}
		return results, nil
	}

	// Lint all KOTS manifests and collect results
	for _, manifestPath := range kotsPaths {
		lint2Result, err := lint2.LintKots(cmd.Context(), manifestPath, kotsVersion)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to lint KOTS manifest: %s", manifestPath)
		}

		// Convert to structured format
		kotsResult := KotsLintResult{
			Path:     manifestPath,
			Success:  lint2Result.Success,
			Messages: convertLint2Messages(lint2Result.Messages),
			Summary:  calculateResourceSummary(lint2Result.Messages),
		}
		results.Manifests = append(results.Manifests, kotsResult)
	}

	// Display results in table format (only if table output)
	if r.outputFormat == "table" {
		// Convert to []LintableResult for generic display
		lintableResults := make([]LintableResult, len(results.Manifests))
		for i, manifest := range results.Manifests {
			lintableResults[i] = manifest
		}
		if err := r.displayLintResults("KOTS", "KOTS manifest", "KOTS manifests", lintableResults); err != nil {
			return nil, errors.Wrap(err, "failed to display KOTS results")
		}
	}

	return results, nil
}

// Removed unused generic display helpers in favor of specific display functions

// initConfigForLint is a simplified version of init flow specifically for lint command
func (r *runners) initConfigForLint(cmd *cobra.Command) error {
	fmt.Fprintf(r.w, "Let's set up a basic linting configuration.\n\n")

	// Auto-detect resources
	detected, err := tools.AutoDetectResources(".")
	if err != nil {
		return errors.Wrap(err, "auto-detecting resources")
	}

	config := &tools.Config{}

	// Show what was detected
	if len(detected.Charts) > 0 {
		fmt.Fprintf(r.w, "Found %d Helm chart(s):\n", len(detected.Charts))
		for _, chart := range detected.Charts {
			fmt.Fprintf(r.w, "  - %s\n", chart)
		}
		fmt.Fprintf(r.w, "\n")

		// Add to config
		for _, chartPath := range detected.Charts {
			if !strings.HasPrefix(chartPath, ".") {
				chartPath = "./" + chartPath
			}
			config.Charts = append(config.Charts, tools.ChartConfig{
				Path: chartPath,
			})
		}
	}

	if len(detected.Preflights) > 0 {
		fmt.Fprintf(r.w, "Found %d preflight spec(s):\n", len(detected.Preflights))
		for _, preflight := range detected.Preflights {
			fmt.Fprintf(r.w, "  - %s\n", preflight)
		}
		fmt.Fprintf(r.w, "\n")

		// Add to config
		for _, preflightPath := range detected.Preflights {
			if !strings.HasPrefix(preflightPath, ".") {
				preflightPath = "./" + preflightPath
			}
			config.Preflights = append(config.Preflights, tools.PreflightConfig{
				Path: preflightPath,
			})
		}
	}

	if len(config.Charts) == 0 && len(config.Preflights) == 0 {
		fmt.Fprintf(r.w, "No Helm charts or preflight specs detected.\n")

		// Prompt for chart path
		chartPrompt := promptui.Prompt{
			Label:   "Chart path (leave empty to skip)",
			Default: "",
		}
		chartPath, _ := chartPrompt.Run()
		if chartPath != "" {
			config.Charts = append(config.Charts, tools.ChartConfig{Path: chartPath})
		}

		// Prompt for preflight path
		preflightPrompt := promptui.Prompt{
			Label:   "Preflight spec path (leave empty to skip)",
			Default: "",
		}
		preflightPath, _ := preflightPrompt.Run()
		if preflightPath != "" {
			config.Preflights = append(config.Preflights, tools.PreflightConfig{Path: preflightPath})
		}
	}

	// Apply defaults
	parser := tools.NewConfigParser()
	parser.ApplyDefaults(config)

	// Write config file
	configPath := filepath.Join(".", ".replicated")
	if err := tools.WriteConfigFile(config, configPath); err != nil {
		return errors.Wrap(err, "writing config file")
	}

	fmt.Fprintf(r.w, "Created %s\n", configPath)

	return nil
}

// extractImagesFromConfig extracts images from charts and returns structured results.
// Accepts already-extracted ChartsWithMetadata and HelmChartManifests to avoid redundant extraction.
func (r *runners) extractImagesFromCharts(ctx context.Context, charts []lint2.ChartWithMetadata, helmChartManifests map[string]*lint2.HelmChartManifest) (*ImageExtractResults, error) {
	extractor := imageextract.NewExtractor()

	if len(charts) == 0 {
		return &ImageExtractResults{
			Images:   []imageextract.ImageRef{},
			Warnings: []imageextract.Warning{},
			Summary:  ImageSummary{TotalImages: 0, UniqueImages: 0},
		}, nil
	}

	// Collect all images from all charts
	imageMap := make(map[string]imageextract.ImageRef) // For deduplication
	var allWarnings []imageextract.Warning

	for _, chart := range charts {
		// Create options for this chart
		opts := imageextract.Options{
			IncludeDuplicates: false,
			NoWarnings:        false,
		}

		// Look for matching HelmChart manifest and apply builder values
		if helmChartManifest := lint2.FindHelmChartManifest(chart.Name, chart.Version, helmChartManifests); helmChartManifest != nil {
			// Apply builder values from HelmChart manifest
			opts.HelmValues = helmChartManifest.BuilderValues
		}

		result, err := extractor.ExtractFromChart(ctx, chart.Path, opts)
		if err != nil {
			allWarnings = append(allWarnings, imageextract.Warning{
				Image:   chart.Path,
				Message: fmt.Sprintf("Failed to extract images: %v", err),
			})
			continue
		}

		// Add images to deduplicated map
		for _, img := range result.Images {
			if existing, ok := imageMap[img.Raw]; ok {
				// Merge sources
				existing.Sources = append(existing.Sources, img.Sources...)
				imageMap[img.Raw] = existing
			} else {
				imageMap[img.Raw] = img
			}
		}

		allWarnings = append(allWarnings, result.Warnings...)
	}

	// Convert map back to slice
	var allImages []imageextract.ImageRef
	for _, img := range imageMap {
		allImages = append(allImages, img)
	}

	return &ImageExtractResults{
		Images:   allImages,
		Warnings: allWarnings,
		Summary: ImageSummary{
			TotalImages:  len(allImages),
			UniqueImages: len(allImages),
		},
	}, nil
}

// displayImages displays image extraction results
func (r *runners) displayImages(results *ImageExtractResults) {
	if results == nil {
		return
	}

	// Print section header
	fmt.Fprintln(r.w, "════════════════════════════════════════════════════════════════════════════")
	fmt.Fprintln(r.w, "IMAGE EXTRACTION")
	fmt.Fprintln(r.w, "════════════════════════════════════════════════════════════════════════════")
	fmt.Fprintln(r.w)
	r.w.Flush()

	// Create a result for the print function
	printResult := &imageextract.Result{
		Images:   results.Images,
		Warnings: results.Warnings,
	}

	// Print images using existing print function
	if err := print.Images("table", r.w, printResult); err != nil {
		fmt.Fprintf(r.w, "Warning: Failed to display images: %v\n", err)
	}

	fmt.Fprintf(r.w, "\nFound %d unique images\n\n", results.Summary.UniqueImages)
	r.w.Flush()
}

// accumulateSummary adds results from a set of lintable resources to the summary.
// Leverages the LintableResult interface to provide generic accumulation across all resource types.
func accumulateSummary(summary *LintSummary, results []LintableResult) {
	for _, result := range results {
		summary.TotalResources++
		if result.GetSuccess() {
			summary.PassedResources++
		} else {
			summary.FailedResources++
		}
		s := result.GetSummary()
		summary.TotalErrors += s.ErrorCount
		summary.TotalWarnings += s.WarningCount
		summary.TotalInfo += s.InfoCount
	}
}

// calculateOverallSummary calculates the overall summary from all results
func (r *runners) calculateOverallSummary(output *JSONLintOutput) LintSummary {
	summary := LintSummary{}

	// Accumulate from Helm results
	if output.HelmResults != nil {
		results := make([]LintableResult, len(output.HelmResults.Charts))
		for i, chart := range output.HelmResults.Charts {
			results[i] = chart
		}
		accumulateSummary(&summary, results)
	}

	// Accumulate from Preflight results
	if output.PreflightResults != nil {
		results := make([]LintableResult, len(output.PreflightResults.Specs))
		for i, spec := range output.PreflightResults.Specs {
			results[i] = spec
		}
		accumulateSummary(&summary, results)
	}

	// Accumulate from Support Bundle results
	if output.SupportBundleResults != nil {
		results := make([]LintableResult, len(output.SupportBundleResults.Specs))
		for i, spec := range output.SupportBundleResults.Specs {
			results[i] = spec
		}
		accumulateSummary(&summary, results)
	}

	// Accumulate from Embedded Cluster results
	if output.EmbeddedClusterResults != nil {
		results := make([]LintableResult, len(output.EmbeddedClusterResults.Configs))
		for i, config := range output.EmbeddedClusterResults.Configs {
			results[i] = config
		}
		accumulateSummary(&summary, results)
	}

	// Accumulate from KOTS results
	if output.KotsResults != nil {
		results := make([]LintableResult, len(output.KotsResults.Manifests))
		for i, manifest := range output.KotsResults.Manifests {
			results[i] = manifest
		}
		accumulateSummary(&summary, results)
	}

	summary.OverallSuccess = summary.FailedResources == 0

	return summary
}

// displayLintResults is a generic function to display lint results for any lintable resource.
// This eliminates duplication across chart, preflight, and support bundle display functions.
func (r *runners) displayLintResults(
	sectionTitle string,
	itemName string, // e.g., "chart", "preflight spec", "support bundle spec"
	pluralName string, // e.g., "charts", "preflight specs", "support bundle specs"
	results []LintableResult,
) error {
	if len(results) == 0 {
		return nil
	}

	// Print section header
	fmt.Fprintln(r.w, "════════════════════════════════════════════════════════════════════════════")
	fmt.Fprintln(r.w, sectionTitle)
	fmt.Fprintln(r.w, "════════════════════════════════════════════════════════════════════════════")
	fmt.Fprintln(r.w)

	for _, result := range results {
		fmt.Fprintf(r.w, "==> Linting %s: %s\n\n", itemName, result.GetPath())

		if len(result.GetMessages()) == 0 {
			fmt.Fprintf(r.w, "No issues found\n")
		} else {
			for _, msg := range result.GetMessages() {
				if msg.Path != "" {
					fmt.Fprintf(r.w, "[%s] %s: %s\n", msg.Severity, msg.Path, msg.Message)
				} else {
					fmt.Fprintf(r.w, "[%s] %s\n", msg.Severity, msg.Message)
				}
			}
		}

		summary := result.GetSummary()
		fmt.Fprintf(r.w, "\nSummary for %s: %d error(s), %d warning(s), %d info\n",
			result.GetPath(), summary.ErrorCount, summary.WarningCount, summary.InfoCount)

		if result.GetSuccess() {
			fmt.Fprintf(r.w, "Status: Passed\n\n")
		} else {
			fmt.Fprintf(r.w, "Status: Failed\n\n")
		}
	}

	// Print overall summary if multiple resources
	if len(results) > 1 {
		totalErrors := 0
		totalWarnings := 0
		totalInfo := 0
		failedResources := 0

		for _, result := range results {
			summary := result.GetSummary()
			totalErrors += summary.ErrorCount
			totalWarnings += summary.WarningCount
			totalInfo += summary.InfoCount
			if !result.GetSuccess() {
				failedResources++
			}
		}

		fmt.Fprintf(r.w, "==> Overall Summary\n")
		fmt.Fprintf(r.w, "%s linted: %d\n", pluralName, len(results))
		fmt.Fprintf(r.w, "%s passed: %d\n", pluralName, len(results)-failedResources)
		fmt.Fprintf(r.w, "%s failed: %d\n", pluralName, failedResources)
		fmt.Fprintf(r.w, "Total errors: %d\n", totalErrors)
		fmt.Fprintf(r.w, "Total warnings: %d\n", totalWarnings)
		fmt.Fprintf(r.w, "Total info: %d\n", totalInfo)

		if failedResources > 0 {
			fmt.Fprintf(r.w, "\nOverall Status: Failed\n")
		} else {
			fmt.Fprintf(r.w, "\nOverall Status: Passed\n")
		}
	}

	return nil
}

// findConfigFilePath finds the .replicated config file path
func findConfigFilePath(startPath string) string {
	currentDir := startPath
	if currentDir == "" {
		var err error
		currentDir, err = os.Getwd()
		if err != nil {
			return ".replicated"
		}
	}

	for {
		// Try .replicated first, then .replicated.yaml
		candidates := []string{
			filepath.Join(currentDir, ".replicated"),
			filepath.Join(currentDir, ".replicated.yaml"),
		}

		for _, configPath := range candidates {
			if stat, err := os.Stat(configPath); err == nil && !stat.IsDir() {
				return configPath
			}
		}

		// Move up one directory
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// Reached root, return default
			return ".replicated"
		}
		currentDir = parentDir
	}
}
