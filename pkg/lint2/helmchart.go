package lint2

import (
	"bytes"
	"fmt"
	"io"
	"os"

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

// FindHelmChartManifest looks up a HelmChart manifest by chart name and version.
// The matching key format is "name:version" which must exactly match both the chart
// metadata and the HelmChart manifest's spec.chart.name and spec.chart.chartVersion.
// Returns nil if no matching manifest is found.
func FindHelmChartManifest(chartName, chartVersion string, manifests map[string]*HelmChartManifest) *HelmChartManifest {
	key := fmt.Sprintf("%s:%s", chartName, chartVersion)
	return manifests[key]
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
		// Return empty map - validation layer will handle this
		return make(map[string]*HelmChartManifest), nil
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
			isHelmChart, err := hasKind(path, "HelmChart")
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

	// Discover EC Config helm charts and merge
	ecHelmCharts, err := DiscoverECConfigHelmCharts(manifestGlobs)
	if err != nil {
		return nil, err
	}
	for key, manifest := range ecHelmCharts {
		if existing, found := helmCharts[key]; found {
			return nil, &DuplicateHelmChartError{
				ChartKey:   key,
				FirstFile:  existing.FilePath,
				SecondFile: manifest.FilePath,
			}
		}
		helmCharts[key] = manifest
	}

	// Return empty map if no HelmCharts found - validation layer will check if charts need HelmCharts
	// Discovery is lenient - validation happens later in the flow
	if len(helmCharts) == 0 {
		return make(map[string]*HelmChartManifest), nil
	}

	return helmCharts, nil
}

// isHelmChartManifest checks if a YAML file contains a HelmChart kind.
// This is a thin wrapper around hasKind for backward compatibility.
func isHelmChartManifest(path string) (bool, error) {
	return hasKind(path, "HelmChart")
}

// DiscoverECConfigHelmCharts scans manifest glob patterns and extracts helm chart
// declarations from embeddedcluster.replicated.com/v1beta1 Config manifests.
// It returns a map keyed by "name:chartVersion" for efficient lookup during validation.
//
// Silently skips:
//   - Files that can't be read
//   - Files that aren't valid YAML
//   - Files that don't contain an EC Config with extensions.helmCharts
//   - Hidden directories (.git, .github, etc.)
func DiscoverECConfigHelmCharts(manifestGlobs []string) (map[string]*HelmChartManifest, error) {
	if len(manifestGlobs) == 0 {
		return make(map[string]*HelmChartManifest), nil
	}

	helmCharts := make(map[string]*HelmChartManifest)
	seenFiles := make(map[string]bool)

	for _, pattern := range manifestGlobs {
		matches, err := GlobFiles(pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to expand manifest pattern %s: %w", pattern, err)
		}

		for _, path := range matches {
			if isHiddenPath(path) {
				continue
			}
			if seenFiles[path] {
				continue
			}
			seenFiles[path] = true

			// Quick apiVersion+kind check before parsing to avoid matching KOTS Config
			isECConfig, err := hasAPIVersionKind(path, ecConfigAPIVersion, "Config")
			if err != nil || !isECConfig {
				continue
			}

			// Parse EC Config helm charts
			manifests, err := parseECConfigHelmCharts(path)
			if err != nil {
				continue
			}

			for _, manifest := range manifests {
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
	}

	if len(helmCharts) == 0 {
		return make(map[string]*HelmChartManifest), nil
	}

	return helmCharts, nil
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

	// Use yaml.NewDecoder to handle multi-document files
	decoder := yaml.NewDecoder(bytes.NewReader(data))

	// Find the first HelmChart document
	for {
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

		err := decoder.Decode(&helmChart)
		if err != nil {
			if err == io.EOF {
				return nil, fmt.Errorf("no HelmChart document found in file")
			}
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}

		if helmChart.Kind == "HelmChart" {
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
	}
}

// parseECConfigHelmCharts reads a YAML file and extracts helm chart declarations from
// any embeddedcluster.replicated.com/v1beta1 Config documents.
// It returns a slice of HelmChartManifest (without BuilderValues, as EC Config doesn't have them).
func parseECConfigHelmCharts(path string) ([]*HelmChartManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	var manifests []*HelmChartManifest

	for {
		var ecConfig struct {
			APIVersion string `yaml:"apiVersion"`
			Kind       string `yaml:"kind"`
			Spec       struct {
				Extensions struct {
					HelmCharts []struct {
						Chart struct {
							Name         string `yaml:"name"`
							ChartVersion string `yaml:"chartVersion"`
						} `yaml:"chart"`
					} `yaml:"helmCharts"`
				} `yaml:"extensions"`
			} `yaml:"spec"`
		}

		err := decoder.Decode(&ecConfig)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}

		if ecConfig.Kind != "Config" || ecConfig.APIVersion != ecConfigAPIVersion {
			continue
		}

		for _, hc := range ecConfig.Spec.Extensions.HelmCharts {
			if hc.Chart.Name == "" || hc.Chart.ChartVersion == "" {
				continue
			}
			manifests = append(manifests, &HelmChartManifest{
				Name:         hc.Chart.Name,
				ChartVersion: hc.Chart.ChartVersion,
				FilePath:     path,
			})
		}
	}

	return manifests, nil
}
