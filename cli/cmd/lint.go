package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/imageextract"
	"github.com/replicatedhq/replicated/pkg/lint2"
	"github.com/replicatedhq/replicated/pkg/tools"
	"github.com/spf13/cobra"
)

func (r *runners) InitLint(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Lint Helm charts, Preflight specs, and Support Bundle specs",
		Long: `Lint Helm charts, Preflight specs, and Support Bundle specs defined in .replicated config file.

This command reads paths from the .replicated config and executes linting locally 
on each resource. Use --verbose to also display extracted container images.`,
		Example: `  # Lint with default table output
  replicated lint

  # Output JSON to stdout
  replicated lint --format json

  # Save results to file (writes to both stdout and file)
  replicated lint --output results.txt

  # Save JSON results to file
  replicated lint --format json --output results.json

  # Use in CI/CD pipelines
  replicated lint --format json | jq '.summary.overall_success'

  # Verbose mode with image extraction
  replicated lint --verbose --format json`,
		SilenceUsage: true,
	}

	cmd.Flags().BoolVarP(&r.args.lintVerbose, "verbose", "v", false, "Show detailed output including extracted container images")
	cmd.Flags().StringVar(&r.outputFormat, "format", "table", "The output format to use. One of: json|table")
	cmd.Flags().StringVarP(&r.args.lintOutputFile, "output", "o", "", "Write output to file at specified path")

	cmd.RunE = r.runLint

	parent.AddCommand(cmd)
	return cmd
}

func (r *runners) runLint(cmd *cobra.Command, args []string) error {
	// Validate format
	if r.outputFormat != "table" && r.outputFormat != "json" {
		return errors.Errorf("invalid format: %s. Supported formats: json, table", r.outputFormat)
	}

	// Load .replicated config using tools parser (supports monorepos)
	parser := tools.NewConfigParser()
	config, err := parser.FindAndParseConfig(".")
	if err != nil {
		// Check if error is "no config found"
		if strings.Contains(err.Error(), "no .replicated config file found") {
			// Offer to create one
			if tools.IsNonInteractive() {
				return errors.New("no .replicated config found. Run 'replicated config init' to create one, or use --chart-dir flag to specify charts directly")
			}

			fmt.Fprintf(r.w, "No .replicated config file found.\n\n")

			// Ask if they want to create one
			prompt := promptui.Select{
				Label: "Would you like to create a .replicated config now?",
				Items: []string{"Yes", "No"},
			}

			_, result, err := prompt.Run()
			if err != nil || result == "No" {
				return errors.New("config file required. Run 'replicated config init' to create one")
			}

			// Run init flow inline
			if err := r.initConfigForLint(cmd); err != nil {
				return errors.Wrap(err, "failed to initialize config")
			}

			// Try loading config again
			config, err = parser.FindAndParseConfig(".")
			if err != nil {
				return errors.Wrap(err, "failed to load newly created config")
			}

			fmt.Fprintf(r.w, "\n")
		} else {
			return errors.Wrap(err, "failed to load .replicated config")
		}
	}

	// Initialize JSON output structure
	output := &JSONLintOutput{}

	// Get Helm version from config
	helmVersion := tools.DefaultHelmVersion
	if config.ReplLint.Tools != nil {
		if v, ok := config.ReplLint.Tools[tools.ToolHelm]; ok {
			helmVersion = v
		}
	}

	// Populate metadata
	configPath := findConfigFilePath(".")
	output.Metadata = newLintMetadata(configPath, helmVersion, "v0.90.0") // TODO: Get actual CLI version

	// Extract and display images if verbose mode is enabled
	if r.args.lintVerbose {
		imageResults, err := r.extractImagesFromConfig(cmd.Context(), config)
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

	// Display tool versions if verbose mode is enabled
	if r.args.lintVerbose {
		fmt.Fprintln(r.w, "Tool versions:")

		// Resolve and display Helm version
		helmVersion := "latest"
		if config.ReplLint.Tools != nil {
			if v, ok := config.ReplLint.Tools[tools.ToolHelm]; ok {
				helmVersion = v
			}
		}
		if helmVersion == "latest" || helmVersion == "" {
			resolver := tools.NewResolver()
			resolvedVersion, err := resolver.ResolveLatestVersion(cmd.Context(), tools.ToolHelm)
			if err == nil {
				helmVersion = resolvedVersion
			}
		}
		fmt.Fprintf(r.w, "  Helm: %s\n", helmVersion)

		// Resolve and display Preflight version
		preflightVersion := "latest"
		if config.ReplLint.Tools != nil {
			if v, ok := config.ReplLint.Tools[tools.ToolPreflight]; ok {
				preflightVersion = v
			}
		}
		if preflightVersion == "latest" || preflightVersion == "" {
			resolver := tools.NewResolver()
			resolvedVersion, err := resolver.ResolveLatestVersion(cmd.Context(), tools.ToolPreflight)
			if err == nil {
				preflightVersion = resolvedVersion
			}
		}
		fmt.Fprintf(r.w, "  Preflight: %s\n", preflightVersion)

		// Resolve and display Support Bundle version
		sbVersion := "latest"
		if config.ReplLint.Tools != nil {
			if v, ok := config.ReplLint.Tools[tools.ToolSupportBundle]; ok {
				sbVersion = v
			}
		}
		if sbVersion == "latest" || sbVersion == "" {
			resolver := tools.NewResolver()
			resolvedVersion, err := resolver.ResolveLatestVersion(cmd.Context(), tools.ToolSupportBundle)
			if err == nil {
				sbVersion = resolvedVersion
			}
		}
		fmt.Fprintf(r.w, "  Support Bundle: %s\n", sbVersion)

		fmt.Fprintln(r.w)
		r.w.Flush()
	}

	// Lint Helm charts if enabled
	if config.ReplLint.Linters.Helm.IsEnabled() {
		if len(config.Charts) == 0 {
			output.HelmResults = &HelmLintResults{Enabled: true, Charts: []ChartLintResult{}}
			if r.outputFormat == "table" {
				fmt.Fprintf(r.w, "No Helm charts configured (skipping Helm linting)\n\n")
			}
		} else {
			helmResults, err := r.lintHelmCharts(cmd, config)
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
		if len(config.Preflights) == 0 {
			output.PreflightResults = &PreflightLintResults{Enabled: true, Specs: []PreflightLintResult{}}
			if r.outputFormat == "table" {
				fmt.Fprintf(r.w, "No preflight specs configured (skipping preflight linting)\n\n")
			}
		} else {
			preflightResults, err := r.lintPreflightSpecs(cmd, config)
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
		sbResults, err := r.lintSupportBundleSpecs(cmd, config)
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

	// Calculate overall summary
	output.Summary = r.calculateOverallSummary(output)

	// Check if output file already exists
	if r.args.lintOutputFile != "" {
		if _, err := os.Stat(r.args.lintOutputFile); err == nil {
			return errors.Errorf("file already exists: %s. Please specify a different path or remove the existing file", r.args.lintOutputFile)
		} else if !os.IsNotExist(err) {
			return errors.Wrapf(err, "failed to check if file exists: %s", r.args.lintOutputFile)
		}
	}

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

	// Output to file if specified
	if r.args.lintOutputFile != "" {
		if err := r.writeOutputToFile(output); err != nil {
			return errors.Wrapf(err, "failed to write output to file: %s", r.args.lintOutputFile)
		}
	}

	// Return error if any linting failed
	if !output.Summary.OverallSuccess {
		return errors.New("linting failed")
	}

	return nil
}

func (r *runners) lintHelmCharts(cmd *cobra.Command, config *tools.Config) (bool, error) {
	// Get helm version from config, default to "latest" if not specified
	helmVersion := "latest"
	if config.ReplLint.Tools != nil {
		if v, ok := config.ReplLint.Tools[tools.ToolHelm]; ok {
			helmVersion = v
		}
	}

	// Check if there are any charts configured
	chartPaths, err := lint2.GetChartPathsFromConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to expand chart paths")
	}

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
		if err := r.displayHelmResults(results); err != nil {
			return nil, errors.Wrap(err, "failed to display helm results")
		}
	}

	return results, nil
}

func (r *runners) lintPreflightSpecs(cmd *cobra.Command, config *tools.Config) (bool, error) {
	// Get preflight version from config, default to "latest" if not specified
	preflightVersion := "latest"
	if config.ReplLint.Tools != nil {
		if v, ok := config.ReplLint.Tools[tools.ToolPreflight]; ok {
			preflightVersion = v
		}
	}

	// Check if there are any preflight specs configured
	preflightPaths, err := lint2.GetPreflightPathsFromConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to expand preflight paths")
	}

	results := &PreflightLintResults{
		Enabled: true,
		Specs:   make([]PreflightLintResult, 0, len(preflightPaths)),
	}

	// Lint all preflight specs and collect results
	for _, specPath := range preflightPaths {
		lint2Result, err := lint2.LintPreflight(cmd.Context(), specPath, preflightVersion)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to lint preflight spec: %s", specPath)
		}

		// Convert to structured format
		preflightResult := PreflightLintResult{
			Path:     specPath,
			Success:  lint2Result.Success,
			Messages: convertLint2Messages(lint2Result.Messages),
			Summary:  calculateResourceSummary(lint2Result.Messages),
		}
		results.Specs = append(results.Specs, preflightResult)
	}

	// Display results in table format (only if table output)
	if r.outputFormat == "table" {
		if err := r.displayPreflightResults(results); err != nil {
			return nil, errors.Wrap(err, "failed to display preflight results")
		}
	}

	return results, nil
}

func (r *runners) lintSupportBundleSpecs(cmd *cobra.Command, config *tools.Config) (bool, error) {
	// Get support-bundle version from config, default to "latest" if not specified
	sbVersion := "latest"
	if config.ReplLint.Tools != nil {
		if v, ok := config.ReplLint.Tools[tools.ToolSupportBundle]; ok {
			sbVersion = v
		}
	}

	// Discover support bundle specs from manifests
	// Support bundles are co-located with other Kubernetes manifests,
	// unlike preflights which are moving to a separate location
	sbPaths, err := lint2.DiscoverSupportBundlesFromManifests(config.Manifests)
	if err != nil {
		return nil, errors.Wrap(err, "failed to discover support bundle specs from manifests")
	}

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
		if err := r.displaySupportBundleResults(results); err != nil {
			return nil, errors.Wrap(err, "failed to display support bundle results")
		}
	}

	return results, nil
}

type resourceSummary struct {
	errorCount   int
	warningCount int
	infoCount    int
}

func displayAllLintResults(w io.Writer, resourceType string, resourcePaths []string, results []*lint2.LintResult) error {
	totalErrors := 0
	totalWarnings := 0
	totalInfo := 0
	totalResourcesFailed := 0

	// Display results for each resource
	for i, result := range results {
		resourcePath := resourcePaths[i]
		summary := displaySingleResourceResult(w, resourceType, resourcePath, result)

		totalErrors += summary.errorCount
		totalWarnings += summary.warningCount
		totalInfo += summary.infoCount

		if !result.Success {
			totalResourcesFailed++
		}
	}

	// Print overall summary if multiple resources
	if len(results) > 1 {
		displayOverallSummary(w, resourceType, len(results), totalResourcesFailed, totalErrors, totalWarnings, totalInfo)
	}

	return nil
}

func displaySingleResourceResult(w io.Writer, resourceType string, resourcePath string, result *lint2.LintResult) resourceSummary {
	// Print header for this resource
	fmt.Fprintf(w, "==> Linting %s: %s\n\n", resourceType, resourcePath)

	// Print messages
	if len(result.Messages) == 0 {
		fmt.Fprintf(w, "No issues found\n")
	} else {
		for _, msg := range result.Messages {
			displayLintMessage(w, msg)
		}
	}

	// Count messages by severity
	summary := countMessagesBySeverity(result.Messages)

	// Print per-resource summary
	fmt.Fprintf(w, "\nSummary for %s: %d error(s), %d warning(s), %d info\n",
		resourcePath, summary.errorCount, summary.warningCount, summary.infoCount)

	// Print per-resource status
	if result.Success {
		fmt.Fprintf(w, "Status: Passed\n\n")
	} else {
		fmt.Fprintf(w, "Status: Failed\n\n")
	}

	return summary
}

func displayLintMessage(w io.Writer, msg lint2.LintMessage) {
	if msg.Path != "" {
		fmt.Fprintf(w, "[%s] %s: %s\n", msg.Severity, msg.Path, msg.Message)
	} else {
		fmt.Fprintf(w, "[%s] %s\n", msg.Severity, msg.Message)
	}
}

func countMessagesBySeverity(messages []lint2.LintMessage) resourceSummary {
	summary := resourceSummary{}
	for _, msg := range messages {
		switch msg.Severity {
		case "ERROR":
			summary.errorCount++
		case "WARNING":
			summary.warningCount++
		case "INFO":
			summary.infoCount++
		}
	}
	return summary
}

func displayOverallSummary(w io.Writer, resourceType string, totalResources, failedResources, totalErrors, totalWarnings, totalInfo int) {
	// Pluralize resource type (simple 's' suffix)
	// Note: Works for current resource types (chart→charts, preflight spec→preflight specs)
	// Would need enhancement for irregular plurals if new resource types are added
	resourceTypePlural := resourceType + "s"

	fmt.Fprintf(w, "==> Overall Summary\n")
	fmt.Fprintf(w, "%s linted: %d\n", resourceTypePlural, totalResources)
	fmt.Fprintf(w, "%s passed: %d\n", resourceTypePlural, totalResources-failedResources)
	fmt.Fprintf(w, "%s failed: %d\n", resourceTypePlural, failedResources)
	fmt.Fprintf(w, "Total errors: %d\n", totalErrors)
	fmt.Fprintf(w, "Total warnings: %d\n", totalWarnings)
	fmt.Fprintf(w, "Total info: %d\n", totalInfo)

	if failedResources > 0 {
		fmt.Fprintf(w, "\nOverall Status: Failed\n")
	} else {
		fmt.Fprintf(w, "\nOverall Status: Passed\n")
	}
}

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

// extractImagesFromConfig extracts images from charts and returns structured results
func (r *runners) extractImagesFromConfig(ctx context.Context, config *tools.Config) (*ImageExtractResults, error) {
	extractor := imageextract.NewExtractor()

	opts := imageextract.Options{
		IncludeDuplicates: false,
		NoWarnings:        false,
	}

	// Get chart paths from config
	chartPaths, err := lint2.GetChartPathsFromConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get chart paths from config")
	}

	if len(chartPaths) == 0 {
		return &ImageExtractResults{
			Images:   []imageextract.ImageRef{},
			Warnings: []imageextract.Warning{},
			Summary:  ImageSummary{TotalImages: 0, UniqueImages: 0},
		}, nil
	}

	// Collect all images from all charts
	imageMap := make(map[string]imageextract.ImageRef) // For deduplication
	var allWarnings []imageextract.Warning

	for _, chartPath := range chartPaths {
		result, err := extractor.ExtractFromChart(ctx, chartPath, opts)
		if err != nil {
			allWarnings = append(allWarnings, imageextract.Warning{
				Image:   chartPath,
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

	fmt.Fprintln(r.w, "Extracting images from Helm charts...")
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

// calculateOverallSummary calculates the overall summary from all results
func (r *runners) calculateOverallSummary(output *JSONLintOutput) LintSummary {
	summary := LintSummary{}

	// Count from Helm results
	if output.HelmResults != nil {
		for _, chart := range output.HelmResults.Charts {
			summary.TotalResources++
			if chart.Success {
				summary.PassedResources++
			} else {
				summary.FailedResources++
			}
			summary.TotalErrors += chart.Summary.ErrorCount
			summary.TotalWarnings += chart.Summary.WarningCount
			summary.TotalInfo += chart.Summary.InfoCount
		}
	}

	// Count from Preflight results
	if output.PreflightResults != nil {
		for _, spec := range output.PreflightResults.Specs {
			summary.TotalResources++
			if spec.Success {
				summary.PassedResources++
			} else {
				summary.FailedResources++
			}
			summary.TotalErrors += spec.Summary.ErrorCount
			summary.TotalWarnings += spec.Summary.WarningCount
			summary.TotalInfo += spec.Summary.InfoCount
		}
	}

	// Count from Support Bundle results
	if output.SupportBundleResults != nil {
		for _, spec := range output.SupportBundleResults.Specs {
			summary.TotalResources++
			if spec.Success {
				summary.PassedResources++
			} else {
				summary.FailedResources++
			}
			summary.TotalErrors += spec.Summary.ErrorCount
			summary.TotalWarnings += spec.Summary.WarningCount
			summary.TotalInfo += spec.Summary.InfoCount
		}
	}

	summary.OverallSuccess = summary.FailedResources == 0

	return summary
}

// displayHelmResults displays Helm lint results in table format
func (r *runners) displayHelmResults(results *HelmLintResults) error {
	if results == nil || len(results.Charts) == 0 {
		return nil
	}

	for _, chart := range results.Charts {
		fmt.Fprintf(r.w, "==> Linting chart: %s\n\n", chart.Path)

		if len(chart.Messages) == 0 {
			fmt.Fprintf(r.w, "No issues found\n")
		} else {
			for _, msg := range chart.Messages {
				if msg.Path != "" {
					fmt.Fprintf(r.w, "[%s] %s: %s\n", msg.Severity, msg.Path, msg.Message)
				} else {
					fmt.Fprintf(r.w, "[%s] %s\n", msg.Severity, msg.Message)
				}
			}
		}

		fmt.Fprintf(r.w, "\nSummary for %s: %d error(s), %d warning(s), %d info\n",
			chart.Path, chart.Summary.ErrorCount, chart.Summary.WarningCount, chart.Summary.InfoCount)

		if chart.Success {
			fmt.Fprintf(r.w, "Status: Passed\n\n")
		} else {
			fmt.Fprintf(r.w, "Status: Failed\n\n")
		}
	}

	// Print overall summary if multiple charts
	if len(results.Charts) > 1 {
		totalErrors := 0
		totalWarnings := 0
		totalInfo := 0
		failedCharts := 0

		for _, chart := range results.Charts {
			totalErrors += chart.Summary.ErrorCount
			totalWarnings += chart.Summary.WarningCount
			totalInfo += chart.Summary.InfoCount
			if !chart.Success {
				failedCharts++
			}
		}

		fmt.Fprintf(r.w, "==> Overall Summary\n")
		fmt.Fprintf(r.w, "charts linted: %d\n", len(results.Charts))
		fmt.Fprintf(r.w, "charts passed: %d\n", len(results.Charts)-failedCharts)
		fmt.Fprintf(r.w, "charts failed: %d\n", failedCharts)
		fmt.Fprintf(r.w, "Total errors: %d\n", totalErrors)
		fmt.Fprintf(r.w, "Total warnings: %d\n", totalWarnings)
		fmt.Fprintf(r.w, "Total info: %d\n", totalInfo)

		if failedCharts > 0 {
			fmt.Fprintf(r.w, "\nOverall Status: Failed\n")
		} else {
			fmt.Fprintf(r.w, "\nOverall Status: Passed\n")
		}
	}

	return nil
}

// displayPreflightResults displays Preflight lint results in table format
func (r *runners) displayPreflightResults(results *PreflightLintResults) error {
	if results == nil || len(results.Specs) == 0 {
		return nil
	}

	for _, spec := range results.Specs {
		fmt.Fprintf(r.w, "==> Linting preflight spec: %s\n\n", spec.Path)

		if len(spec.Messages) == 0 {
			fmt.Fprintf(r.w, "No issues found\n")
		} else {
			for _, msg := range spec.Messages {
				if msg.Path != "" {
					fmt.Fprintf(r.w, "[%s] %s: %s\n", msg.Severity, msg.Path, msg.Message)
				} else {
					fmt.Fprintf(r.w, "[%s] %s\n", msg.Severity, msg.Message)
				}
			}
		}

		fmt.Fprintf(r.w, "\nSummary for %s: %d error(s), %d warning(s), %d info\n",
			spec.Path, spec.Summary.ErrorCount, spec.Summary.WarningCount, spec.Summary.InfoCount)

		if spec.Success {
			fmt.Fprintf(r.w, "Status: Passed\n\n")
		} else {
			fmt.Fprintf(r.w, "Status: Failed\n\n")
		}
	}

	// Print overall summary if multiple specs
	if len(results.Specs) > 1 {
		totalErrors := 0
		totalWarnings := 0
		totalInfo := 0
		failedSpecs := 0

		for _, spec := range results.Specs {
			totalErrors += spec.Summary.ErrorCount
			totalWarnings += spec.Summary.WarningCount
			totalInfo += spec.Summary.InfoCount
			if !spec.Success {
				failedSpecs++
			}
		}

		fmt.Fprintf(r.w, "==> Overall Summary\n")
		fmt.Fprintf(r.w, "preflight specs linted: %d\n", len(results.Specs))
		fmt.Fprintf(r.w, "preflight specs passed: %d\n", len(results.Specs)-failedSpecs)
		fmt.Fprintf(r.w, "preflight specs failed: %d\n", failedSpecs)
		fmt.Fprintf(r.w, "Total errors: %d\n", totalErrors)
		fmt.Fprintf(r.w, "Total warnings: %d\n", totalWarnings)
		fmt.Fprintf(r.w, "Total info: %d\n", totalInfo)

		if failedSpecs > 0 {
			fmt.Fprintf(r.w, "\nOverall Status: Failed\n")
		} else {
			fmt.Fprintf(r.w, "\nOverall Status: Passed\n")
		}
	}

	return nil
}

// displaySupportBundleResults displays Support Bundle lint results in table format
func (r *runners) displaySupportBundleResults(results *SupportBundleLintResults) error {
	if results == nil || len(results.Specs) == 0 {
		return nil
	}

	for _, spec := range results.Specs {
		fmt.Fprintf(r.w, "==> Linting support bundle spec: %s\n\n", spec.Path)

		if len(spec.Messages) == 0 {
			fmt.Fprintf(r.w, "No issues found\n")
		} else {
			for _, msg := range spec.Messages {
				if msg.Path != "" {
					fmt.Fprintf(r.w, "[%s] %s: %s\n", msg.Severity, msg.Path, msg.Message)
				} else {
					fmt.Fprintf(r.w, "[%s] %s\n", msg.Severity, msg.Message)
				}
			}
		}

		fmt.Fprintf(r.w, "\nSummary for %s: %d error(s), %d warning(s), %d info\n",
			spec.Path, spec.Summary.ErrorCount, spec.Summary.WarningCount, spec.Summary.InfoCount)

		if spec.Success {
			fmt.Fprintf(r.w, "Status: Passed\n\n")
		} else {
			fmt.Fprintf(r.w, "Status: Failed\n\n")
		}
	}

	// Print overall summary if multiple specs
	if len(results.Specs) > 1 {
		totalErrors := 0
		totalWarnings := 0
		totalInfo := 0
		failedSpecs := 0

		for _, spec := range results.Specs {
			totalErrors += spec.Summary.ErrorCount
			totalWarnings += spec.Summary.WarningCount
			totalInfo += spec.Summary.InfoCount
			if !spec.Success {
				failedSpecs++
			}
		}

		fmt.Fprintf(r.w, "==> Overall Summary\n")
		fmt.Fprintf(r.w, "support bundle specs linted: %d\n", len(results.Specs))
		fmt.Fprintf(r.w, "support bundle specs passed: %d\n", len(results.Specs)-failedSpecs)
		fmt.Fprintf(r.w, "support bundle specs failed: %d\n", failedSpecs)
		fmt.Fprintf(r.w, "Total errors: %d\n", totalErrors)
		fmt.Fprintf(r.w, "Total warnings: %d\n", totalWarnings)
		fmt.Fprintf(r.w, "Total info: %d\n", totalInfo)

		if failedSpecs > 0 {
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

// writeOutputToFile writes lint output to a file
func (r *runners) writeOutputToFile(output *JSONLintOutput) error {
	// Create the file
	file, err := os.Create(r.args.lintOutputFile)
	if err != nil {
		return errors.Wrap(err, "failed to create output file")
	}
	defer file.Close()

	// For JSON format, write JSON directly
	if r.outputFormat == "json" {
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(output); err != nil {
			return errors.Wrap(err, "failed to write JSON to file")
		}
		return nil
	}

	// For table format, we need to recreate the table output
	// Create a tabwriter for the file
	w := tabwriter.NewWriter(file, minWidth, tabWidth, padding, padChar, tabwriter.TabIndent)

	// Re-display helm results
	if output.HelmResults != nil && output.HelmResults.Enabled {
		for _, chart := range output.HelmResults.Charts {
			fmt.Fprintf(w, "==> Linting chart: %s\n\n", chart.Path)

			if len(chart.Messages) == 0 {
				fmt.Fprintf(w, "No issues found\n")
			} else {
				for _, msg := range chart.Messages {
					if msg.Path != "" {
						fmt.Fprintf(w, "[%s] %s: %s\n", msg.Severity, msg.Path, msg.Message)
					} else {
						fmt.Fprintf(w, "[%s] %s\n", msg.Severity, msg.Message)
					}
				}
			}

			fmt.Fprintf(w, "\nSummary for %s: %d error(s), %d warning(s), %d info\n",
				chart.Path, chart.Summary.ErrorCount, chart.Summary.WarningCount, chart.Summary.InfoCount)

			if chart.Success {
				fmt.Fprintf(w, "Status: Passed\n\n")
			} else {
				fmt.Fprintf(w, "Status: Failed\n\n")
			}
		}

		// Print overall summary if multiple charts
		if len(output.HelmResults.Charts) > 1 {
			totalErrors := 0
			totalWarnings := 0
			totalInfo := 0
			failedCharts := 0

			for _, chart := range output.HelmResults.Charts {
				totalErrors += chart.Summary.ErrorCount
				totalWarnings += chart.Summary.WarningCount
				totalInfo += chart.Summary.InfoCount
				if !chart.Success {
					failedCharts++
				}
			}

			fmt.Fprintf(w, "==> Overall Summary\n")
			fmt.Fprintf(w, "charts linted: %d\n", len(output.HelmResults.Charts))
			fmt.Fprintf(w, "charts passed: %d\n", len(output.HelmResults.Charts)-failedCharts)
			fmt.Fprintf(w, "charts failed: %d\n", failedCharts)
			fmt.Fprintf(w, "Total errors: %d\n", totalErrors)
			fmt.Fprintf(w, "Total warnings: %d\n", totalWarnings)
			fmt.Fprintf(w, "Total info: %d\n", totalInfo)

			if failedCharts > 0 {
				fmt.Fprintf(w, "\nOverall Status: Failed\n")
			} else {
				fmt.Fprintf(w, "\nOverall Status: Passed\n")
			}
		}
	}

	// Display preflight results
	if output.PreflightResults != nil && output.PreflightResults.Enabled {
		for _, spec := range output.PreflightResults.Specs {
			fmt.Fprintf(w, "==> Linting preflight spec: %s\n\n", spec.Path)

			if len(spec.Messages) == 0 {
				fmt.Fprintf(w, "No issues found\n")
			} else {
				for _, msg := range spec.Messages {
					if msg.Path != "" {
						fmt.Fprintf(w, "[%s] %s: %s\n", msg.Severity, msg.Path, msg.Message)
					} else {
						fmt.Fprintf(w, "[%s] %s\n", msg.Severity, msg.Message)
					}
				}
			}

			fmt.Fprintf(w, "\nSummary for %s: %d error(s), %d warning(s), %d info\n",
				spec.Path, spec.Summary.ErrorCount, spec.Summary.WarningCount, spec.Summary.InfoCount)

			if spec.Success {
				fmt.Fprintf(w, "Status: Passed\n\n")
			} else {
				fmt.Fprintf(w, "Status: Failed\n\n")
			}
		}
	}

	// Display support bundle results
	if output.SupportBundleResults != nil && output.SupportBundleResults.Enabled {
		for _, spec := range output.SupportBundleResults.Specs {
			fmt.Fprintf(w, "==> Linting support bundle spec: %s\n\n", spec.Path)

			if len(spec.Messages) == 0 {
				fmt.Fprintf(w, "No issues found\n")
			} else {
				for _, msg := range spec.Messages {
					if msg.Path != "" {
						fmt.Fprintf(w, "[%s] %s: %s\n", msg.Severity, msg.Path, msg.Message)
					} else {
						fmt.Fprintf(w, "[%s] %s\n", msg.Severity, msg.Message)
					}
				}
			}

			fmt.Fprintf(w, "\nSummary for %s: %d error(s), %d warning(s), %d info\n",
				spec.Path, spec.Summary.ErrorCount, spec.Summary.WarningCount, spec.Summary.InfoCount)

			if spec.Success {
				fmt.Fprintf(w, "Status: Passed\n\n")
			} else {
				fmt.Fprintf(w, "Status: Failed\n\n")
			}
		}
	}

	// Display disabled linters messages
	if output.HelmResults != nil && !output.HelmResults.Enabled {
		fmt.Fprintf(w, "Helm linting is disabled in .replicated config\n\n")
	}
	if output.PreflightResults != nil && !output.PreflightResults.Enabled {
		fmt.Fprintf(w, "Preflight linting is disabled in .replicated config\n\n")
	}
	if output.SupportBundleResults != nil && !output.SupportBundleResults.Enabled {
		fmt.Fprintf(w, "Support Bundle linting is disabled in .replicated config\n\n")
	}

	// Flush and close
	if err := w.Flush(); err != nil {
		return errors.Wrap(err, "failed to flush output to file")
	}

	return nil
}
