package lint2

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

// HelmChartManifest represents a parsed KOTS HelmChart custom resource.
// It contains the fields needed to match charts with their builder values
// for preflight template rendering.
type HelmChartManifest struct {
	Name          string                 // spec.chart.name - must match Chart.yaml name
	ChartVersion  string                 // spec.chart.chartVersion - must match Chart.yaml version
	BuilderValues map[string]interface{} // spec.builder - values for air gap bundle rendering (can be nil/empty)
	FilePath      string                 // Source file path for error reporting
}

// DuplicateHelmChartError is returned when multiple HelmChart manifests
// are found with the same name:chartVersion combination.
type DuplicateHelmChartError struct {
	ChartKey   string // "name:chartVersion"
	FirstFile  string
	SecondFile string
}

func (e *DuplicateHelmChartError) Error() string {
	return fmt.Sprintf(
		"duplicate HelmChart manifest found for chart %q\n"+
			"  First:  %s\n"+
			"  Second: %s\n"+
			"Each chart name:version pair must be unique",
		e.ChartKey, e.FirstFile, e.SecondFile,
	)
}

// DiscoverHelmChartManifests scans manifest glob patterns and extracts HelmChart custom resources.
// It returns a map keyed by "name:chartVersion" for efficient lookup during preflight rendering.
//
// Accepts HelmChart resources with any apiVersion (validation happens in the linter).
//
// Returns an error if:
//   - manifestGlobs is empty (required to find builder values for templated preflights)
//   - Duplicate name:chartVersion pairs are found (ambiguous builder values)
//   - Glob expansion fails
//
// Silently skips:
//   - Files that can't be read
//   - Files that aren't valid YAML
//   - Files that don't contain kind: HelmChart
//   - Hidden directories (.git, .github, etc.)
func DiscoverHelmChartManifests(manifestGlobs []string) (map[string]*HelmChartManifest, error) {
	if len(manifestGlobs) == 0 {
		// Error instead of returning empty map (unlike DiscoverSupportBundlesFromManifests)
		// because HelmChart discovery is only called when preflights have templated values,
		// so manifests are required to find builder values
		return nil, fmt.Errorf("no manifests configured - cannot discover HelmChart resources (required for templated preflights)")
	}

	helmCharts := make(map[string]*HelmChartManifest)
	seenFiles := make(map[string]bool) // Global deduplication across all patterns

	for _, pattern := range manifestGlobs {
		// Expand glob pattern to find YAML files
		matches, err := GlobFiles(pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to expand manifest pattern %s: %w", pattern, err)
		}

		for _, path := range matches {
			// Skip hidden paths (.git, .github, etc.)
			if isHiddenPath(path) {
				continue
			}

			// Skip if already processed (patterns can overlap)
			if seenFiles[path] {
				continue
			}
			seenFiles[path] = true

			// Check if this file contains a HelmChart resource
			isHelmChart, err := isHelmChartManifest(path)
			if err != nil {
				// Skip files we can't read or parse
				continue
			}
			if !isHelmChart {
				// Not a HelmChart - skip silently (allows mixed manifest directories)
				continue
			}

			// Parse the HelmChart manifest
			manifest, err := parseHelmChartManifest(path)
			if err != nil {
				// Skip malformed HelmCharts (missing required fields, etc.)
				continue
			}

			// Check for duplicates
			key := fmt.Sprintf("%s:%s", manifest.Name, manifest.ChartVersion)
			if existing, found := helmCharts[key]; found {
				return nil, &DuplicateHelmChartError{
					ChartKey:   key,
					FirstFile:  existing.FilePath,
					SecondFile: manifest.FilePath,
				}
			}

			helmCharts[key] = manifest
		}
	}

	return helmCharts, nil
}

// isHelmChartManifest checks if a YAML file contains a HelmChart kind.
// Handles multi-document YAML files properly using yaml.NewDecoder.
// Falls back to regex matching if the file has parse errors.
func isHelmChartManifest(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	// Use yaml.Decoder for proper multi-document YAML parsing
	decoder := yaml.NewDecoder(bytes.NewReader(data))

	// Iterate through all documents in the file
	for {
		var kindDoc struct {
			Kind string `yaml:"kind"`
		}

		err := decoder.Decode(&kindDoc)
		if err != nil {
			if err == io.EOF {
				// Reached end of file - no more documents
				break
			}
			// Parse error - file is malformed
			// Fall back to regex matching to detect if this looks like a HelmChart
			// This allows invalid YAML files to still be discovered (consistent with preflight/support bundle discovery)
			matched, _ := regexp.Match(`(?m)^kind:\s+HelmChart\s*$`, data)
			if matched {
				return true, nil
			}
			return false, nil
		}

		// Check if this document is a HelmChart
		if kindDoc.Kind == "HelmChart" {
			return true, nil
		}
	}

	return false, nil
}

// parseHelmChartManifest parses a HelmChart manifest and extracts the fields needed for preflight rendering.
// Accepts any apiVersion (validation happens in the linter).
//
// Returns an error if required fields are missing:
//   - spec.chart.name
//   - spec.chart.chartVersion
//
// The spec.builder field is optional (can be nil or empty).
func parseHelmChartManifest(path string) (*HelmChartManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse the full HelmChart structure
	// Support both v1beta1 and v1beta2 - they have the same structure for fields we need
	var helmChart struct {
		APIVersion string `yaml:"apiVersion"`
		Kind       string `yaml:"kind"`
		Spec       struct {
			Chart struct {
				Name         string `yaml:"name"`
				ChartVersion string `yaml:"chartVersion"`
			} `yaml:"chart"`
			Builder map[string]interface{} `yaml:"builder"`
		} `yaml:"spec"`
	}

	// Use yaml.NewDecoder to handle multi-document files
	decoder := yaml.NewDecoder(bytes.NewReader(data))

	// Find the first HelmChart document
	for {
		err := decoder.Decode(&helmChart)
		if err != nil {
			if err == io.EOF {
				return nil, fmt.Errorf("no HelmChart document found in file")
			}
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}

		if helmChart.Kind == "HelmChart" {
			break
		}
	}

	// Validate required fields
	if helmChart.Spec.Chart.Name == "" {
		return nil, fmt.Errorf("spec.chart.name is required but not found")
	}
	if helmChart.Spec.Chart.ChartVersion == "" {
		return nil, fmt.Errorf("spec.chart.chartVersion is required but not found")
	}

	// Note: We don't validate apiVersion here - discovery is permissive.
	// The preflight linter will validate apiVersion when it processes the HelmChart.
	// This allows future apiVersions to work without code changes.

	return &HelmChartManifest{
		Name:          helmChart.Spec.Chart.Name,
		ChartVersion:  helmChart.Spec.Chart.ChartVersion,
		BuilderValues: helmChart.Spec.Builder, // Can be nil or empty - that's valid
		FilePath:      path,
	}, nil
}
