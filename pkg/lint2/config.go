package lint2

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/replicatedhq/replicated/pkg/tools"
)

// GetChartPathsFromConfig extracts and expands chart paths from config
func GetChartPathsFromConfig(config *tools.Config) ([]string, error) {
	if len(config.Charts) == 0 {
		return nil, fmt.Errorf("no charts found in .replicated config")
	}

	return expandPaths(config.Charts, func(c tools.ChartConfig) string { return c.Path }, DiscoverChartPaths, "charts")
}

// expandPaths is a generic helper that expands resource paths from config.
// It takes a slice of configs, extracts the path from each using getPath,
// discovers resources using discoveryFunc, and validates that matches are found.
func expandPaths[T any](
	configs []T,
	getPath func(T) string,
	discoveryFunc func(string) ([]string, error),
	resourceName string,
) ([]string, error) {
	var paths []string

	for _, config := range configs {
		path := getPath(config)
		// Discovery function handles both explicit paths and glob patterns
		matches, err := discoveryFunc(path)
		if err != nil {
			return nil, fmt.Errorf("failed to discover %s from %s: %w", resourceName, path, err)
		}
		if len(matches) == 0 {
			return nil, fmt.Errorf("no %s found matching: %s", resourceName, path)
		}
		paths = append(paths, matches...)
	}

	return paths, nil
}

// ChartWithMetadata pairs a chart path with its metadata from Chart.yaml
type ChartWithMetadata struct {
	Path    string // Absolute path to the chart directory
	Name    string // Chart name from Chart.yaml
	Version string // Chart version from Chart.yaml
}

// GetChartsWithMetadataFromConfig extracts chart paths and their metadata from config
// This function combines GetChartPathsFromConfig with metadata extraction, reducing
// boilerplate for callers that need both path and metadata information (like image extraction).
func GetChartsWithMetadataFromConfig(config *tools.Config) ([]ChartWithMetadata, error) {
	chartPaths, err := GetChartPathsFromConfig(config)
	if err != nil {
		return nil, err
	}

	var results []ChartWithMetadata
	for _, chartPath := range chartPaths {
		metadata, err := GetChartMetadata(chartPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read chart metadata for %s: %w", chartPath, err)
		}

		results = append(results, ChartWithMetadata{
			Path:    chartPath,
			Name:    metadata.Name,
			Version: metadata.Version,
		})
	}

	return results, nil
}

// ==============================================================================
// Typed Errors for Chart Lookup
// ==============================================================================

// ChartNotFoundError is returned when a preflight references a chart that doesn't exist in the config.
type ChartNotFoundError struct {
	RequestedChart  string   // "name:version"
	AvailableCharts []string // List of available chart keys
}

func (e *ChartNotFoundError) Error() string {
	return fmt.Sprintf(
		"chart %q not found in charts configuration\n"+
			"Available charts: %v\n\n"+
			"Ensure the chart is listed in the 'charts' section of your .replicated config:\n"+
			"charts:\n"+
			"  - path: \"./path/to/chart\"",
		e.RequestedChart, e.AvailableCharts,
	)
}

// DuplicateChartError is returned when multiple charts have the same name:version.
type DuplicateChartError struct {
	ChartKey   string // "name:version"
	FirstPath  string
	SecondPath string
}

func (e *DuplicateChartError) Error() string {
	return fmt.Sprintf(
		"duplicate chart %q found in configuration\n"+
			"  First:  %s\n"+
			"  Second: %s\n\n"+
			"Each chart name:version pair must be unique. Consider:\n"+
			"- Renaming one chart in Chart.yaml\n"+
			"- Changing the version in Chart.yaml\n"+
			"- Removing one chart from the configuration",
		e.ChartKey, e.FirstPath, e.SecondPath,
	)
}

// ==============================================================================
// Chart Lookup Utilities
// ==============================================================================

// getChartKeys extracts chart keys from a lookup map.
// Helper function for error messages listing available charts.
func getChartKeys(lookup map[string]*ChartWithMetadata) []string {
	keys := make([]string, 0, len(lookup))
	for k := range lookup {
		keys = append(keys, k)
	}
	return keys
}

// BuildChartLookup creates a name:version lookup map for charts.
// Returns error if duplicate chart name:version pairs are found.
// This utility is reusable by any code that needs to resolve charts by name:version.
func BuildChartLookup(config *tools.Config) (map[string]*ChartWithMetadata, error) {
	if len(config.Charts) == 0 {
		return nil, fmt.Errorf("no charts found in .replicated config")
	}

	// Get all charts with metadata
	charts, err := GetChartsWithMetadataFromConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to get charts from config: %w", err)
	}

	// Build lookup map
	chartLookup := make(map[string]*ChartWithMetadata)
	for i := range charts {
		key := fmt.Sprintf("%s:%s", charts[i].Name, charts[i].Version)
		if existing, exists := chartLookup[key]; exists {
			return nil, &DuplicateChartError{
				ChartKey:   key,
				FirstPath:  existing.Path,
				SecondPath: charts[i].Path,
			}
		}
		chartLookup[key] = &charts[i]
	}

	return chartLookup, nil
}

// findValuesFile finds values.yaml or values.yml in a chart directory.
// Checks for values.yaml first (most common), then falls back to values.yml.
// Returns the path if found, or an error if neither exists.
//
// This follows the same pattern as Chart.yaml/Chart.yml detection in GetChartMetadata.
func findValuesFile(chartPath string) (string, error) {
	// Check values.yaml first (most common)
	valuesYaml := filepath.Join(chartPath, "values.yaml")
	if _, err := os.Stat(valuesYaml); err == nil {
		return valuesYaml, nil
	}

	// Check values.yml as fallback
	valuesYml := filepath.Join(chartPath, "values.yml")
	if _, err := os.Stat(valuesYml); err == nil {
		return valuesYml, nil
	}

	return "", fmt.Errorf("no values.yaml or values.yml found in chart directory %s", chartPath)
}

// GetPreflightPathsFromConfig extracts and expands preflight spec paths from config
func GetPreflightPathsFromConfig(config *tools.Config) ([]string, error) {
	if len(config.Preflights) == 0 {
		return nil, fmt.Errorf("no preflights found in .replicated config")
	}

	return expandPaths(config.Preflights, func(p tools.PreflightConfig) string { return p.Path }, DiscoverPreflightPaths, "preflight specs")
}

// PreflightWithValues contains preflight spec path and optional values file
type PreflightWithValues struct {
	SpecPath     string // Path to the preflight spec file
	ValuesPath   string // Path to values.yaml (optional - passed to preflight lint if provided)
	ChartName    string // Chart name from Chart.yaml (used to look up HelmChart manifest for builder values)
	ChartVersion string // Chart version from Chart.yaml (used to look up HelmChart manifest for builder values)
}

// resolvePreflightWithChart resolves a single preflight spec with an explicit chart reference.
// Uses v1beta3 detection to determine strict (error on failures) vs lenient (continue with empty values) behavior.
//
// Parameters:
//   - specPath: path to the preflight spec file
//   - chartName: explicit chart name reference
//   - chartVersion: explicit chart version reference
//   - chartLookup: pre-built name:version lookup map (or nil if chartLookupErr is set)
//   - chartLookupErr: error from BuildChartLookup (or nil if chartLookup is valid)
//
// Returns:
//   - PreflightWithValues with resolved chart/values information
//   - error if strict validation fails (v1beta3) or if critical errors occur
//
// Behavior:
//   - v1beta3 specs: errors if chart lookup fails, chart not found, or values file missing (strict)
//   - v1beta2 specs: continues with empty values if lookups fail (lenient)
func resolvePreflightWithChart(
	specPath string,
	chartName string,
	chartVersion string,
	chartLookup map[string]*ChartWithMetadata,
	chartLookupErr error,
) (PreflightWithValues, error) {
	// Check if this is v1beta3 (determines strict vs lenient behavior)
	isV1Beta3, err := isPreflightV1Beta3(specPath)
	if err != nil {
		return PreflightWithValues{}, fmt.Errorf("failed to check preflight version for %s: %w", specPath, err)
	}

	// Handle chart lookup failure
	if chartLookupErr != nil {
		if isV1Beta3 {
			// v1beta3 requires charts to be configured
			return PreflightWithValues{}, fmt.Errorf("v1beta3 preflight %s requires charts configuration: %w", specPath, chartLookupErr)
		}
		// v1beta2 can continue without charts (lenient)
		return PreflightWithValues{
			SpecPath:     specPath,
			ValuesPath:   "",
			ChartName:    "",
			ChartVersion: "",
		}, nil
	}

	// Look up chart by name:version
	key := fmt.Sprintf("%s:%s", chartName, chartVersion)
	chart, found := chartLookup[key]
	if !found {
		if isV1Beta3 {
			// v1beta3 requires the chart to exist (strict)
			return PreflightWithValues{}, &ChartNotFoundError{
				RequestedChart:  key,
				AvailableCharts: getChartKeys(chartLookup),
			}
		}
		// v1beta2 can continue without the specific chart (lenient)
		return PreflightWithValues{
			SpecPath:     specPath,
			ValuesPath:   "",
			ChartName:    chartName,
			ChartVersion: chartVersion,
		}, nil
	}

	// Find values file
	valuesPath, err := findValuesFile(chart.Path)
	if err != nil {
		if isV1Beta3 {
			// v1beta3 requires values file to exist (strict)
			return PreflightWithValues{}, fmt.Errorf("chart %q: %w\nEnsure the chart directory contains values.yaml or values.yml", key, err)
		}
		// v1beta2 can continue without values file (lenient)
		return PreflightWithValues{
			SpecPath:     specPath,
			ValuesPath:   "",
			ChartName:    chart.Name,
			ChartVersion: chart.Version,
		}, nil
	}

	// Success: all lookups worked
	return PreflightWithValues{
		SpecPath:     specPath,
		ValuesPath:   valuesPath,
		ChartName:    chart.Name,
		ChartVersion: chart.Version,
	}, nil
}

// GetPreflightWithValuesFromConfig extracts preflight paths with optional chart/values information.
// If chartName/chartVersion are provided, uses explicit chart references.
// If not provided, returns empty values and lets the linter decide requirements.
func GetPreflightWithValuesFromConfig(config *tools.Config) ([]PreflightWithValues, error) {
	if len(config.Preflights) == 0 {
		return nil, fmt.Errorf("no preflights found in .replicated config")
	}

	// Build chart lookup once (lazy initialization - only when first needed)
	// This is shared across all preflights to avoid redundant rebuilding
	var chartLookup map[string]*ChartWithMetadata
	var chartLookupErr error
	var chartLookupBuilt bool

	var results []PreflightWithValues

	for _, preflightConfig := range config.Preflights {
		// Discover preflight spec paths (handles globs)
		specPaths, err := DiscoverPreflightPaths(preflightConfig.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to discover preflights from %s: %w", preflightConfig.Path, err)
		}
		if len(specPaths) == 0 {
			return nil, fmt.Errorf("no preflight specs found matching: %s", preflightConfig.Path)
		}

		// BRANCH 1: Chart reference provided (explicit reference pattern)
		if preflightConfig.ChartName != "" {
			// Validate chartVersion also provided
			if preflightConfig.ChartVersion == "" {
				return nil, fmt.Errorf("preflight %s: chartVersion required when chartName is specified", preflightConfig.Path)
			}

			// Build chart lookup on first use (lazy initialization for performance)
			if !chartLookupBuilt {
				chartLookup, chartLookupErr = BuildChartLookup(config)
				chartLookupBuilt = true
			}

			// Process each discovered spec using helper function
			for _, specPath := range specPaths {
				result, err := resolvePreflightWithChart(
					specPath,
					preflightConfig.ChartName,
					preflightConfig.ChartVersion,
					chartLookup,
					chartLookupErr,
				)
				if err != nil {
					return nil, err
				}
				results = append(results, result)
			}
		} else {
			// BRANCH 2: No chart reference provided (linter decides)
			for _, specPath := range specPaths {
				results = append(results, PreflightWithValues{
					SpecPath:     specPath,
					ValuesPath:   "",
					ChartName:    "",
					ChartVersion: "",
				})
			}
		}
	}

	return results, nil
}
