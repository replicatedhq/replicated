package lint2

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/replicatedhq/replicated/pkg/tools"
	"gopkg.in/yaml.v3"
)

// GetChartPathsFromConfig extracts and expands chart paths from config
func GetChartPathsFromConfig(config *tools.Config) ([]string, error) {
	if len(config.Charts) == 0 {
		return nil, fmt.Errorf("no charts found in .replicated config")
	}

	return expandChartPaths(config.Charts)
}

// expandChartPaths expands glob patterns in chart paths and returns a list of concrete paths
func expandChartPaths(chartConfigs []tools.ChartConfig) ([]string, error) {
	var paths []string

	for _, chartConfig := range chartConfigs {
		// Check if path contains glob pattern
		if containsGlob(chartConfig.Path) {
			// Use content-aware discovery to find charts
			matches, err := discoverChartPaths(chartConfig.Path)
			if err != nil {
				return nil, fmt.Errorf("failed to discover charts from pattern %s: %w", chartConfig.Path, err)
			}
			if len(matches) == 0 {
				// Error if pattern matches nothing - this catches typos and invalid patterns
				return nil, fmt.Errorf("no charts found matching pattern: %s", chartConfig.Path)
			}
			// All matches are already validated by discoverChartPaths
			paths = append(paths, matches...)
		} else {
			// Validate single path
			if err := validateChartPath(chartConfig.Path); err != nil {
				return nil, fmt.Errorf("invalid chart path %s: %w", chartConfig.Path, err)
			}
			paths = append(paths, chartConfig.Path)
		}
	}

	return paths, nil
}

// containsGlob checks if a path contains glob wildcards
// Calls exported ContainsGlob for consistency
func containsGlob(path string) bool {
	return ContainsGlob(path)
}

// validateChartPath checks if a path is a valid Helm chart directory
func validateChartPath(path string) error {
	// Check if path exists and is a directory
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist")
		}
		return fmt.Errorf("failed to stat path: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory")
	}

	// Check if Chart.yaml exists in the directory
	chartYaml := filepath.Join(path, "Chart.yaml")
	if _, err := os.Stat(chartYaml); err != nil {
		// Try Chart.yml as fallback
		chartYml := filepath.Join(path, "Chart.yml")
		if _, err := os.Stat(chartYml); err != nil {
			return fmt.Errorf("Chart.yaml or Chart.yml not found (not a valid Helm chart)")
		}
	}

	return nil
}

// GetPreflightPathsFromConfig extracts and expands preflight spec paths from config
func GetPreflightPathsFromConfig(config *tools.Config) ([]string, error) {
	if len(config.Preflights) == 0 {
		return nil, fmt.Errorf("no preflights found in .replicated config")
	}

	return expandPreflightPaths(config.Preflights)
}

// expandPreflightPaths expands glob patterns in preflight paths and returns a list of concrete file paths
func expandPreflightPaths(preflightConfigs []tools.PreflightConfig) ([]string, error) {
	var paths []string

	for _, preflightConfig := range preflightConfigs {
		// Check if path contains glob pattern
		if containsGlob(preflightConfig.Path) {
			// Use content-aware discovery to find preflights
			matches, err := discoverPreflightPaths(preflightConfig.Path)
			if err != nil {
				return nil, fmt.Errorf("failed to discover preflights from pattern %s: %w", preflightConfig.Path, err)
			}
			if len(matches) == 0 {
				// Error if pattern matches nothing - this catches typos and invalid patterns
				return nil, fmt.Errorf("no preflight specs found matching pattern: %s", preflightConfig.Path)
			}
			// All matches are already validated by discoverPreflightPaths
			paths = append(paths, matches...)
		} else {
			// Validate single path
			if err := validatePreflightPath(preflightConfig.Path); err != nil {
				return nil, fmt.Errorf("invalid preflight spec path %s: %w", preflightConfig.Path, err)
			}
			paths = append(paths, preflightConfig.Path)
		}
	}

	return paths, nil
}

// validatePreflightPath checks if a path is a valid preflight spec file
func validatePreflightPath(path string) error {
	// Check if path exists and is a file
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist")
		}
		return fmt.Errorf("failed to stat path: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("path is a directory, expected a file")
	}

	// Check if file has .yaml or .yml extension
	ext := filepath.Ext(path)
	if ext != ".yaml" && ext != ".yml" {
		return fmt.Errorf("file must have .yaml or .yml extension")
	}

	// Check if file actually contains kind: Preflight
	isPreflight, err := isPreflightSpec(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	if !isPreflight {
		return fmt.Errorf("file does not contain kind: Preflight (not a valid Preflight spec)")
	}

	return nil
}

// PreflightWithValues contains preflight spec path and associated chart/values information
// All fields are required - every preflight must have an associated chart structure
type PreflightWithValues struct {
	SpecPath     string // Path to the preflight spec file
	ValuesPath   string // Path to values.yaml for template rendering (required)
	ChartName    string // Chart name from Chart.yaml (required)
	ChartVersion string // Chart version from Chart.yaml (required)
}

// ChartMetadata represents the minimal Chart.yaml structure needed for matching
type ChartMetadata struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

// parseChartYaml reads and parses a Chart.yaml file
func parseChartYaml(path string) (*ChartMetadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read Chart.yaml: %w", err)
	}

	var chart ChartMetadata
	if err := yaml.Unmarshal(data, &chart); err != nil {
		return nil, fmt.Errorf("failed to parse Chart.yaml: %w", err)
	}

	if chart.Name == "" {
		return nil, fmt.Errorf("Chart.yaml missing required field: name")
	}
	if chart.Version == "" {
		return nil, fmt.Errorf("Chart.yaml missing required field: version")
	}

	return &chart, nil
}

// GetPreflightWithValuesFromConfig extracts preflight paths with associated chart/values information
func GetPreflightWithValuesFromConfig(config *tools.Config) ([]PreflightWithValues, error) {
	if len(config.Preflights) == 0 {
		return nil, fmt.Errorf("no preflights found in .replicated config")
	}

	var results []PreflightWithValues

	for _, preflightConfig := range config.Preflights {
		// Handle glob patterns in preflight path
		var specPaths []string
		if containsGlob(preflightConfig.Path) {
			matches, err := discoverPreflightPaths(preflightConfig.Path)
			if err != nil {
				return nil, fmt.Errorf("failed to discover preflights from pattern %s: %w", preflightConfig.Path, err)
			}
			if len(matches) == 0 {
				return nil, fmt.Errorf("no preflight specs found matching pattern: %s", preflightConfig.Path)
			}
			specPaths = matches
		} else {
			if err := validatePreflightPath(preflightConfig.Path); err != nil {
				return nil, fmt.Errorf("invalid preflight spec path %s: %w", preflightConfig.Path, err)
			}
			specPaths = []string{preflightConfig.Path}
		}

		// Create PreflightWithValues for each discovered spec
		for _, specPath := range specPaths {
			// valuesPath is REQUIRED - error if missing
			if preflightConfig.ValuesPath == "" {
				return nil, fmt.Errorf("preflight (%s) missing required field 'valuesPath'\n"+
					"All preflights must specify a valuesPath pointing to chart values.yaml", specPath)
			}

			result := PreflightWithValues{
				SpecPath:   specPath,
				ValuesPath: preflightConfig.ValuesPath,
			}

			// Extract chart metadata (always required)
			valuesDir := filepath.Dir(preflightConfig.ValuesPath)
			chartYamlPath := filepath.Join(valuesDir, "Chart.yaml")

			// Try Chart.yml as fallback
			if _, err := os.Stat(chartYamlPath); err != nil {
				chartYmlPath := filepath.Join(valuesDir, "Chart.yml")
				if _, err := os.Stat(chartYmlPath); err == nil {
					chartYamlPath = chartYmlPath
				} else {
					return nil, fmt.Errorf("Chart.yaml not found for preflight\nExpected at: %s\nPreflight: %s", chartYamlPath, specPath)
				}
			}

			// Parse Chart.yaml to get name and version
			chart, err := parseChartYaml(chartYamlPath)
			if err != nil {
				return nil, fmt.Errorf("failed to parse Chart.yaml for preflight %s: %w", specPath, err)
			}

			result.ChartName = chart.Name
			result.ChartVersion = chart.Version

			results = append(results, result)
		}
	}

	return results, nil
}
