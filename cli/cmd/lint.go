package cmd

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

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
		Use:          "lint",
		Short:        "Lint Helm charts, Preflight specs, and Support Bundle specs",
		Long:         `Lint Helm charts, Preflight specs, and Support Bundle specs defined in .replicated config file. This command reads paths from the .replicated config and executes linting locally on each resource. Use --verbose to also display extracted container images.`,
		SilenceUsage: true,
	}

	cmd.Flags().BoolVarP(&r.args.lintVerbose, "verbose", "v", false, "Show detailed output including extracted container images")

	cmd.RunE = r.runLint

	parent.AddCommand(cmd)
	return cmd
}

func (r *runners) runLint(cmd *cobra.Command, args []string) error {
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

	hasFailure := false

	// Extract and display images if verbose mode is enabled
	if r.args.lintVerbose {
		if err := r.extractAndDisplayImagesFromConfig(cmd.Context(), config); err != nil {
			// Log warning but don't fail the lint command
			fmt.Fprintf(r.w, "Warning: Failed to extract images: %v\n\n", err)
			r.w.Flush()
		}

		// Print separator
		fmt.Fprintln(r.w, "────────────────────────────────────────────────────────────────────────────")
		fmt.Fprintln(r.w, "\nRunning lint checks...")
		fmt.Fprintln(r.w)
		r.w.Flush()
	}

	// Lint Helm charts if enabled
	if config.ReplLint.Linters.Helm.IsEnabled() {
		if len(config.Charts) == 0 {
			fmt.Fprintf(r.w, "No Helm charts configured (skipping Helm linting)\n\n")
		} else {
			helmFailed, err := r.lintHelmCharts(cmd, config)
			if err != nil {
				return err
			}
			if helmFailed {
				hasFailure = true
			}
		}
	} else {
		fmt.Fprintf(r.w, "Helm linting is disabled in .replicated config\n\n")
	}

	// Lint Preflight specs if enabled
	if config.ReplLint.Linters.Preflight.IsEnabled() {
		if len(config.Preflights) == 0 {
			fmt.Fprintf(r.w, "No preflight specs configured (skipping preflight linting)\n\n")
		} else {
			preflightFailed, err := r.lintPreflightSpecs(cmd, config)
			if err != nil {
				return err
			}
			if preflightFailed {
				hasFailure = true
			}
		}
	} else {
		fmt.Fprintf(r.w, "Preflight linting is disabled in .replicated config\n\n")
	}

	// Lint Support Bundle specs if enabled
	if config.ReplLint.Linters.SupportBundle.IsEnabled() {
		sbFailed, err := r.lintSupportBundleSpecs(cmd, config)
		if err != nil {
			return err
		}
		if sbFailed {
			hasFailure = true
		}
	} else {
		fmt.Fprintf(r.w, "Support Bundle linting is disabled in .replicated config\n\n")
	}

	// Flush the tab writer
	if err := r.w.Flush(); err != nil {
		return errors.Wrap(err, "failed to flush output")
	}

	// Return error if any linting failed
	if hasFailure {
		return errors.New("linting failed")
	}

	return nil
}

func (r *runners) lintHelmCharts(cmd *cobra.Command, config *tools.Config) (bool, error) {
	// Get helm version from config
	helmVersion := tools.DefaultHelmVersion
	if config.ReplLint.Tools != nil {
		if v, ok := config.ReplLint.Tools[tools.ToolHelm]; ok {
			helmVersion = v
		}
	}

	// Check if there are any charts configured
	chartPaths, err := lint2.GetChartPathsFromConfig(config)
	if err != nil {
		return false, errors.Wrap(err, "failed to expand chart paths")
	}

	// Lint all charts and collect results
	var allResults []*lint2.LintResult
	var allPaths []string
	hasFailure := false

	for _, chartPath := range chartPaths {
		result, err := lint2.LintChart(cmd.Context(), chartPath, helmVersion)
		if err != nil {
			return false, errors.Wrapf(err, "failed to lint chart: %s", chartPath)
		}

		allResults = append(allResults, result)
		allPaths = append(allPaths, chartPath)

		if !result.Success {
			hasFailure = true
		}
	}

	// Display results for all charts
	if err := displayAllLintResults(r.w, "chart", allPaths, allResults); err != nil {
		return false, errors.Wrap(err, "failed to display lint results")
	}

	return hasFailure, nil
}

func (r *runners) lintPreflightSpecs(cmd *cobra.Command, config *tools.Config) (bool, error) {
	// Get preflight version from config
	preflightVersion := tools.DefaultPreflightVersion
	if config.ReplLint.Tools != nil {
		if v, ok := config.ReplLint.Tools[tools.ToolPreflight]; ok {
			preflightVersion = v
		}
	}

	// Check if there are any preflight specs configured
	preflightPaths, err := lint2.GetPreflightPathsFromConfig(config)
	if err != nil {
		return false, errors.Wrap(err, "failed to expand preflight paths")
	}

	// Lint all preflight specs and collect results
	var allResults []*lint2.LintResult
	var allPaths []string
	hasFailure := false

	for _, specPath := range preflightPaths {
		result, err := lint2.LintPreflight(cmd.Context(), specPath, preflightVersion)
		if err != nil {
			return false, errors.Wrapf(err, "failed to lint preflight spec: %s", specPath)
		}

		allResults = append(allResults, result)
		allPaths = append(allPaths, specPath)

		if !result.Success {
			hasFailure = true
		}
	}

	// Display results for all preflight specs
	if err := displayAllLintResults(r.w, "preflight spec", allPaths, allResults); err != nil {
		return false, errors.Wrap(err, "failed to display lint results")
	}

	return hasFailure, nil
}

func (r *runners) lintSupportBundleSpecs(cmd *cobra.Command, config *tools.Config) (bool, error) {
	// Get support-bundle version from config
	sbVersion := tools.DefaultSupportBundleVersion
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
		return false, errors.Wrap(err, "failed to discover support bundle specs from manifests")
	}

	// If no support bundles found, that's not an error - they're optional
	if len(sbPaths) == 0 {
		return false, nil
	}

	// Lint all support bundle specs and collect results
	var allResults []*lint2.LintResult
	var allPaths []string
	hasFailure := false

	for _, specPath := range sbPaths {
		result, err := lint2.LintSupportBundle(cmd.Context(), specPath, sbVersion)
		if err != nil {
			return false, errors.Wrapf(err, "failed to lint support bundle spec: %s", specPath)
		}

		allResults = append(allResults, result)
		allPaths = append(allPaths, specPath)

		if !result.Success {
			hasFailure = true
		}
	}

	// Display results for all support bundle specs
	if err := displayAllLintResults(r.w, "support bundle spec", allPaths, allResults); err != nil {
		return false, errors.Wrap(err, "failed to display lint results")
	}

	return hasFailure, nil
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

func (r *runners) extractAndDisplayImagesFromConfig(ctx context.Context, config *tools.Config) error {
	extractor := imageextract.NewExtractor()

	opts := imageextract.Options{
		IncludeDuplicates: false,
		NoWarnings:        true, // Suppress warnings in lint context
	}

	fmt.Fprintln(r.w, "Extracting images from Helm charts...")
	fmt.Fprintln(r.w)
	r.w.Flush()

	// Get chart paths from config
	chartPaths, err := lint2.GetChartPathsFromConfig(config)
	if err != nil {
		return errors.Wrap(err, "failed to get chart paths from config")
	}

	if len(chartPaths) == 0 {
		fmt.Fprintln(r.w, "No Helm charts found in .replicated config")
		fmt.Fprintln(r.w)
		r.w.Flush()
		return nil
	}

	// Collect all images from all charts
	var allImages []imageextract.ImageRef
	imageMap := make(map[string]imageextract.ImageRef) // For deduplication

	for _, chartPath := range chartPaths {
		result, err := extractor.ExtractFromChart(ctx, chartPath, opts)
		if err != nil {
			fmt.Fprintf(r.w, "Warning: Failed to extract images from %s: %v\n", chartPath, err)
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
	}

	// Convert map back to slice
	for _, img := range imageMap {
		allImages = append(allImages, img)
	}

	// Create a result with all images
	combinedResult := &imageextract.Result{
		Images: allImages,
	}

	// Print images using existing print function
	if err := print.Images("table", r.w, combinedResult); err != nil {
		return err
	}

	fmt.Fprintf(r.w, "\nFound %d unique images across %d chart(s)\n\n", len(allImages), len(chartPaths))
	return r.w.Flush()
}
