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

	return expandChartPaths(config.Charts)
}

// expandChartPaths expands glob patterns in chart paths and returns a list of concrete paths
func expandChartPaths(chartConfigs []tools.ChartConfig) ([]string, error) {
	var paths []string

	for _, chartConfig := range chartConfigs {
		// Check if path contains glob pattern
		if containsGlob(chartConfig.Path) {
			matches, err := GlobDirs(chartConfig.Path)
			if err != nil {
				return nil, fmt.Errorf("failed to expand glob pattern %s: %w", chartConfig.Path, err)
			}
			if len(matches) == 0 {
				return nil, fmt.Errorf("no charts found matching pattern: %s", chartConfig.Path)
			}
			// Validate each matched path
			for _, match := range matches {
				if err := validateChartPath(match); err != nil {
					return nil, fmt.Errorf("invalid chart path %s: %w", match, err)
				}
			}
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
			matches, err := GlobFiles(preflightConfig.Path)
			if err != nil {
				return nil, fmt.Errorf("failed to expand glob pattern %s: %w", preflightConfig.Path, err)
			}
			if len(matches) == 0 {
				return nil, fmt.Errorf("no preflight specs found matching pattern: %s", preflightConfig.Path)
			}
			// Validate each matched path
			for _, match := range matches {
				if err := validatePreflightPath(match); err != nil {
					return nil, fmt.Errorf("invalid preflight spec path %s: %w", match, err)
				}
			}
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

	return nil
}
