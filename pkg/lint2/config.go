package lint2

import (
	"fmt"

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
	ChartName    string // Deprecated: no longer used
	ChartVersion string // Deprecated: no longer used
}

// GetPreflightWithValuesFromConfig extracts preflight paths with associated chart/values information
func GetPreflightWithValuesFromConfig(config *tools.Config) ([]PreflightWithValues, error) {
	if len(config.Preflights) == 0 {
		return nil, fmt.Errorf("no preflights found in .replicated config")
	}

	var results []PreflightWithValues

	for _, preflightConfig := range config.Preflights {
		// Discovery handles both explicit paths and glob patterns
		specPaths, err := DiscoverPreflightPaths(preflightConfig.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to discover preflights from %s: %w", preflightConfig.Path, err)
		}
		if len(specPaths) == 0 {
			return nil, fmt.Errorf("no preflight specs found matching: %s", preflightConfig.Path)
		}

		// Create PreflightWithValues for each discovered spec
		for _, specPath := range specPaths {
			result := PreflightWithValues{
				SpecPath:   specPath,
				ValuesPath: preflightConfig.ValuesPath, // Optional - can be empty
			}

			results = append(results, result)
		}
	}

	return results, nil
}
