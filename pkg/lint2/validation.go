package lint2

import (
	"fmt"
	"strings"
)

// ChartMissingHelmChartInfo contains information about a chart missing its HelmChart manifest.
type ChartMissingHelmChartInfo struct {
	ChartPath    string // Absolute path to chart directory
	ChartName    string // Chart name from Chart.yaml
	ChartVersion string // Chart version from Chart.yaml
}

// MultipleChartsMissingHelmChartsError is returned when multiple charts are missing HelmChart manifests.
type MultipleChartsMissingHelmChartsError struct {
	MissingCharts []ChartMissingHelmChartInfo
}

func (e *MultipleChartsMissingHelmChartsError) Error() string {
	var b strings.Builder

	if len(e.MissingCharts) == 1 {
		// Single chart - simple message
		c := e.MissingCharts[0]
		fmt.Fprintf(&b, "Missing HelmChart manifest for chart at %q (%s:%s)\n\n",
			c.ChartPath, c.ChartName, c.ChartVersion)
	} else {
		// Multiple charts - list all
		fmt.Fprintf(&b, "Chart validation failed - %d charts missing HelmChart manifests:\n", len(e.MissingCharts))
		for _, c := range e.MissingCharts {
			fmt.Fprintf(&b, "  - %s (%s:%s)\n", c.ChartPath, c.ChartName, c.ChartVersion)
		}
		fmt.Fprintln(&b)
	}

	fmt.Fprint(&b, "Each Helm chart requires a corresponding HelmChart manifest (kind: HelmChart).\n")
	fmt.Fprint(&b, "Ensure the manifests are in paths specified in the 'manifests' section of .replicated config.")

	return b.String()
}

// ChartToHelmChartValidationResult contains validation results and warnings
type ChartToHelmChartValidationResult struct {
	Warnings []string // Non-fatal issues (orphaned HelmChart manifests)
}

// ValidateChartToHelmChartMapping validates that every chart has a corresponding HelmChart manifest.
//
// Requirements:
//   - Every chart in charts must have a matching HelmChart manifest in helmChartManifests
//   - Matching key format: "chartName:chartVersion"
//
// Returns:
//   - result: Contains warnings (orphaned HelmChart manifests)
//   - error: Hard error if any chart is missing its HelmChart manifest (batch reports all missing)
//
// Behavior:
//   - Hard error: Chart exists but no matching HelmChart manifest (collects ALL missing, reports together)
//   - Warning: HelmChart manifest exists but no matching chart configured
//   - Duplicate HelmChart manifests are detected in DiscoverHelmChartManifests()
func ValidateChartToHelmChartMapping(
	charts []ChartWithMetadata,
	helmChartManifests map[string]*HelmChartManifest,
) (*ChartToHelmChartValidationResult, error) {
	result := &ChartToHelmChartValidationResult{
		Warnings: []string{},
	}

	// Track which HelmChart manifests are matched (to detect orphans)
	matchedManifests := make(map[string]bool)

	// Collect all missing charts (batch reporting)
	var missingCharts []ChartMissingHelmChartInfo

	// Check each chart has a corresponding HelmChart manifest
	for _, chart := range charts {
		key := fmt.Sprintf("%s:%s", chart.Name, chart.Version)

		if _, found := helmChartManifests[key]; !found {
			// Collect missing chart info (don't return yet)
			missingCharts = append(missingCharts, ChartMissingHelmChartInfo{
				ChartPath:    chart.Path,
				ChartName:    chart.Name,
				ChartVersion: chart.Version,
			})
		} else {
			// Mark manifest as matched
			matchedManifests[key] = true
		}
	}

	// Return all missing charts at once
	if len(missingCharts) > 0 {
		return nil, &MultipleChartsMissingHelmChartsError{
			MissingCharts: missingCharts,
		}
	}

	// Check for orphaned HelmChart manifests (warn but don't error)
	for key, manifest := range helmChartManifests {
		if !matchedManifests[key] {
			warning := fmt.Sprintf(
				"HelmChart manifest at %q (%s) has no corresponding chart configured",
				manifest.FilePath, key,
			)
			result.Warnings = append(result.Warnings, warning)
		}
	}

	return result, nil
}
