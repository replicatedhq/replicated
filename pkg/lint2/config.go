package lint2

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
			matches, err := filepath.Glob(chartConfig.Path)
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
func containsGlob(path string) bool {
	return strings.ContainsAny(path, "*?[")
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
