package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

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
		Short: "Lint Helm charts and Preflight specs",
		Long: `Lint Helm charts and Preflight specs defined in .replicated config file.

This command reads paths from the .replicated config and executes linting locally 
on each resource. Use --verbose to also display extracted container images.`,
		Example: `  # Lint with default table output
  replicated lint

  # Output JSON to stdout
  replicated lint --output json

  # Save results to file (writes to both stdout and file)
  replicated lint --output-file results.txt

  # Save JSON results to file
  replicated lint --output json --output-file results.json

  # Use in CI/CD pipelines
  replicated lint -o json | jq '.summary.overall_success'

  # Verbose mode with image extraction
  replicated lint --verbose`,
		SilenceUsage: true,
	}

	cmd.Flags().BoolVarP(&r.args.lintVerbose, "verbose", "v", false, "Show detailed output including extracted container images")
	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "The output format to use. One of: json|table")
	cmd.Flags().StringVar(&r.args.lintOutputFile, "output-file", "", "Write output to file at specified path")

	cmd.RunE = r.runLint

	parent.AddCommand(cmd)
	return cmd
}

func (r *runners) runLint(cmd *cobra.Command, args []string) error {
	// Validate output format
	if r.outputFormat != "table" && r.outputFormat != "json" {
		return errors.Errorf("invalid output format: %s. Supported formats: json, table", r.outputFormat)
	}

	// Load .replicated config using tools parser (supports monorepos)
	parser := tools.NewConfigParser()
	config, err := parser.FindAndParseConfig(".")
	if err != nil {
		return errors.Wrap(err, "failed to load .replicated config")
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

	// Lint Helm charts if enabled
	if config.ReplLint.Linters.Helm.IsEnabled() {
		helmResults, err := r.lintHelmCharts(cmd, config)
		if err != nil {
			return err
		}
		output.HelmResults = helmResults
	} else {
		output.HelmResults = &HelmLintResults{Enabled: false, Charts: []ChartLintResult{}}
		if r.outputFormat == "table" {
			fmt.Fprintf(r.w, "Helm linting is disabled in .replicated config\n\n")
		}
	}

	// Lint Preflight specs if enabled
	if config.ReplLint.Linters.Preflight.IsEnabled() {
		preflightResults, err := r.lintPreflightSpecs(cmd, config)
		if err != nil {
			return err
		}
		output.PreflightResults = preflightResults
	} else {
		output.PreflightResults = &PreflightLintResults{Enabled: false, Specs: []PreflightLintResult{}}
		if r.outputFormat == "table" {
			fmt.Fprintf(r.w, "Preflight linting is disabled in .replicated config\n\n")
		}
	}

	// Calculate overall summary
	output.Summary = r.calculateOverallSummary(output)

	// Check if output file already exists
	if r.args.lintOutputFile != "" {
		if _, err := os.Stat(r.args.lintOutputFile); err == nil {
			return errors.Errorf("File already exists: %s. Please specify a different path or remove the existing file.", r.args.lintOutputFile)
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

func (r *runners) lintHelmCharts(cmd *cobra.Command, config *tools.Config) (*HelmLintResults, error) {
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

func (r *runners) lintPreflightSpecs(cmd *cobra.Command, config *tools.Config) (*PreflightLintResults, error) {
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

	// Display disabled linters messages
	if output.HelmResults != nil && !output.HelmResults.Enabled {
		fmt.Fprintf(w, "Helm linting is disabled in .replicated config\n\n")
	}
	if output.PreflightResults != nil && !output.PreflightResults.Enabled {
		fmt.Fprintf(w, "Preflight linting is disabled in .replicated config\n\n")
	}

	// Flush and close
	if err := w.Flush(); err != nil {
		return errors.Wrap(err, "failed to flush output to file")
	}

	return nil
}
