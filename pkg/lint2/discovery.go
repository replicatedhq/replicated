package lint2

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// DiscoverSupportBundlesFromManifests discovers support bundle spec files from manifest glob patterns.
// It expands the glob patterns, reads each YAML file, and identifies files containing kind: SupportBundle.
// This allows support bundles to be co-located with other Kubernetes manifests without explicit configuration.
func DiscoverSupportBundlesFromManifests(manifestGlobs []string) ([]string, error) {
	if len(manifestGlobs) == 0 {
		// No manifests configured - return empty list, not an error
		return []string{}, nil
	}

	var allPaths []string
	seenPaths := make(map[string]bool) // Global deduplication across all patterns

	for _, pattern := range manifestGlobs {
		// Use smart pattern discovery
		paths, err := DiscoverSupportBundlePaths(pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to discover support bundles from pattern %s: %w", pattern, err)
		}

		// Deduplicate across patterns
		for _, path := range paths {
			if !seenPaths[path] {
				seenPaths[path] = true
				allPaths = append(allPaths, path)
			}
		}
	}

	return allPaths, nil
}

// isHiddenPath checks if a path contains any hidden directory components (starting with .)
// Returns true for paths like .git, .github, foo/.hidden/bar, etc.
// Does not consider . or .. as hidden (current/parent directory references).
func isHiddenPath(path string) bool {
	parts := strings.Split(filepath.ToSlash(path), "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ".") && part != "." && part != ".." {
			return true
		}
	}
	return false
}

// isChartDirectory checks if a directory contains a Chart.yaml or Chart.yml file.
// Returns true if the directory is a valid Helm chart directory.
func isChartDirectory(dirPath string) (bool, error) {
	chartYaml := filepath.Join(dirPath, "Chart.yaml")
	chartYml := filepath.Join(dirPath, "Chart.yml")

	// Check Chart.yaml
	if _, err := os.Stat(chartYaml); err == nil {
		return true, nil
	}

	// Check Chart.yml
	if _, err := os.Stat(chartYml); err == nil {
		return true, nil
	}

	return false, nil
}

// discoverChartPaths discovers Helm chart directories from a glob pattern.
// This is a thin wrapper around discoverDirsByMarkerFile for backward compatibility.
//
// Supports patterns like:
//   - "./charts/**"              (finds all charts recursively)
//   - "./charts/{app,api}/**"    (finds charts in specific subdirectories)
//   - "./pkg/**/Chart.yaml"      (explicit Chart.yaml pattern)
//   - "./my-chart"               (explicit directory path - validated strictly)
func DiscoverChartPaths(pattern string) ([]string, error) {
	return discoverDirsByMarkerFile(pattern, []string{"Chart.yaml", "Chart.yml"}, "Helm chart")
}

// hasKind checks if a YAML file contains a specific kind.
// Handles multi-document YAML files properly using yaml.NewDecoder, which correctly
// handles document separators (---) even when they appear inside strings or block scalars.
// For files with syntax errors, falls back to simple regex matching to detect the kind.
//
// Pass the kind name (e.g., "Preflight", "SupportBundle", "HelmChart") to check for.
func hasKind(path string, kind string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	// Use yaml.Decoder for proper multi-document YAML parsing
	// This correctly handles --- separators according to the YAML spec
	decoder := yaml.NewDecoder(bytes.NewReader(data))

	// Iterate through all documents in the file
	for {
		// Parse just the kind field (lightweight)
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
			// Fall back to regex matching to detect if this looks like the target kind
			// This allows invalid YAML files to still be discovered and linted
			// Use regex to match "kind: <kind>" as a complete line (not in comments/strings)
			pattern := fmt.Sprintf(`(?m)^kind:\s+%s\s*$`, regexp.QuoteMeta(kind))
			matched, _ := regexp.Match(pattern, data)
			return matched, nil
		}

		// Check if this document matches the target kind
		if kindDoc.Kind == kind {
			return true, nil
		}
	}

	return false, nil
}

// discoverPreflightPaths discovers Preflight spec files from a glob pattern.
// This is a thin wrapper around discoverYAMLsByKind for backward compatibility.
//
// Supports patterns like:
//   - "./preflights/**"           (finds all Preflight specs recursively)
//   - "./preflights/**/*.yaml"    (explicit YAML extension)
//   - "./k8s/{dev,prod}/**/*.yaml" (environment-specific)
//   - "./preflight.yaml"          (explicit file path - validated strictly)
func DiscoverPreflightPaths(pattern string) ([]string, error) {
	return discoverYAMLsByKind(pattern, "Preflight", "preflight spec")
}

// DiscoverHelmChartPaths discovers HelmChart manifest files from a glob pattern.
// This is a thin wrapper around discoverYAMLsByKind for backward compatibility.
//
// Supports patterns like:
//   - "./manifests/**"           (finds all HelmChart manifests recursively)
//   - "./manifests/**/*.yaml"    (explicit YAML extension)
//   - "./k8s/{dev,prod}/**/*.yaml" (environment-specific)
//   - "./helmchart.yaml"         (explicit file path - validated strictly)
func DiscoverHelmChartPaths(pattern string) ([]string, error) {
	return discoverYAMLsByKind(pattern, "HelmChart", "HelmChart manifest")
}

// (duplicate isPreflightSpec removed)
// (duplicate isSupportBundleSpec removed)

// ==============================================================================
// Generic Discovery Functions
// ==============================================================================
//
// The functions below provide generic, reusable discovery logic for both
// YAML files (by kind) and directories (by marker files). They eliminate
// duplication across chart, preflight, and support bundle discovery.

// buildYAMLPatterns classifies a pattern and builds search patterns for YAML files.
// Handles: explicit .yaml/.yml, /*, /**, brace expansion, etc.
func buildYAMLPatterns(pattern string) ([]string, error) {
	if strings.HasSuffix(pattern, ".yaml") || strings.HasSuffix(pattern, ".yml") {
		return []string{pattern}, nil
	}

	if strings.HasSuffix(pattern, "/*") {
		basePattern := strings.TrimSuffix(pattern, "/*")
		return []string{
			filepath.Join(basePattern, "*.yaml"),
			filepath.Join(basePattern, "*.yml"),
		}, nil
	}

	if strings.HasSuffix(pattern, "**") || strings.Contains(pattern, "{") {
		return []string{
			filepath.Join(pattern, "*.yaml"),
			filepath.Join(pattern, "*.yml"),
		}, nil
	}

	// Check if it's a literal file path
	ext := filepath.Ext(pattern)
	if ext == ".yaml" || ext == ".yml" {
		return []string{pattern}, nil
	}

	return nil, fmt.Errorf("pattern must end with .yaml, .yml, *, or **")
}

// validateExplicitYAMLFile validates a single YAML file path and checks its kind.
// Returns the path in a slice for consistency with discovery functions.
// Returns error if file doesn't exist, isn't a file, has wrong extension, or doesn't contain the kind.
func validateExplicitYAMLFile(path, kind, resourceName string) ([]string, error) {
	// Check extension
	ext := filepath.Ext(path)
	if ext != ".yaml" && ext != ".yml" {
		return nil, fmt.Errorf("file must have .yaml or .yml extension")
	}

	// Check exists and is file
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("path does not exist")
		}
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory, expected a file")
	}

	// Check kind
	hasTargetKind, err := hasKind(path, kind)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	if !hasTargetKind {
		return nil, fmt.Errorf("file does not contain kind: %s (not a valid %s)", kind, resourceName)
	}

	return []string{path}, nil
}

// filterYAMLFilesByKind expands glob patterns and filters to files with matching kind.
// Silently skips files that can't be read or don't have the target kind.
func filterYAMLFilesByKind(patterns []string, originalPattern, kind string) ([]string, error) {
	var resultPaths []string
	seenPaths := make(map[string]bool)

	for _, p := range patterns {
		matches, err := GlobFiles(p)
		if err != nil {
			return nil, fmt.Errorf("expanding pattern %s: %w (from user pattern: %s)", p, err, originalPattern)
		}

		for _, path := range matches {
			// Skip hidden paths
			if isHiddenPath(path) {
				continue
			}

			// Skip duplicates
			if seenPaths[path] {
				continue
			}
			seenPaths[path] = true

			// Check kind
			hasTargetKind, err := hasKind(path, kind)
			if err != nil {
				// Skip files we can't read
				continue
			}

			if hasTargetKind {
				resultPaths = append(resultPaths, path)
			}
		}
	}

	return resultPaths, nil
}

// discoverYAMLsByKind discovers YAML files containing a specific kind from a pattern.
// Handles both explicit file paths (strict validation) and glob patterns (lenient filtering).
//
// For explicit paths:
//   - Validates file exists, is a file, has .yaml/.yml extension
//   - Checks if file contains the specified kind
//   - Returns error if any validation fails (fail loudly)
//
// For glob patterns:
//   - Expands pattern to find all YAML files
//   - Filters to only files containing the specified kind
//   - Silently skips files that don't match (allows mixed directories)
func discoverYAMLsByKind(pattern, kind, resourceName string) ([]string, error) {
	// Validate empty pattern
	if pattern == "" {
		return nil, fmt.Errorf("pattern cannot be empty")
	}

	// Preserve original for error messages
	originalPattern := pattern

	// Normalize path
	pattern = filepath.Clean(pattern)

	// Check if explicit path vs glob
	isExplicitPath := !ContainsGlob(pattern)

	if isExplicitPath {
		// Strict validation
		return validateExplicitYAMLFile(pattern, kind, resourceName)
	}

	// Glob pattern - build search patterns
	patterns, err := buildYAMLPatterns(pattern)
	if err != nil {
		return nil, err
	}

	// Lenient filtering
	return filterYAMLFilesByKind(patterns, originalPattern, kind)
}

// validateExplicitChartDir validates an explicit directory path for chart discovery.
// Returns the path in a slice for consistency with discovery functions.
func validateExplicitChartDir(path string) ([]string, error) {
	// Check exists and is directory
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("path does not exist")
		}
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory")
	}

	// Check has Chart.yaml or Chart.yml
	isChart, err := isChartDirectory(path)
	if err != nil {
		return nil, fmt.Errorf("checking if %s is chart directory: %w", path, err)
	}
	if !isChart {
		return nil, fmt.Errorf("directory %s is not a valid Helm chart (no Chart.yaml or Chart.yml found)", path)
	}

	return []string{path}, nil
}

// filterDirsByMarkerFile expands glob patterns to find marker files and returns their parent directories.
// Silently skips hidden paths and deduplicates results.
func filterDirsByMarkerFile(patterns []string, originalPattern string) ([]string, error) {
	var chartDirs []string
	seenDirs := make(map[string]bool)

	for _, p := range patterns {
		matches, err := GlobFiles(p)
		if err != nil {
			return nil, fmt.Errorf("expanding pattern %s: %w (from user pattern: %s)", p, err, originalPattern)
		}

		for _, markerPath := range matches {
			chartDir := filepath.Dir(markerPath)

			if isHiddenPath(chartDir) {
				continue
			}

			if seenDirs[chartDir] {
				continue
			}
			seenDirs[chartDir] = true

			chartDirs = append(chartDirs, chartDir)
		}
	}

	return chartDirs, nil
}

// discoverDirsByMarkerFile discovers directories containing specific marker files.
// Handles both explicit directory paths (strict validation) and glob patterns (lenient filtering).
//
// For explicit paths:
//   - Validates path exists and is a directory
//   - Checks if directory contains any of the marker files
//   - Returns error if validation fails
//
// For glob patterns:
//   - Expands pattern to find marker files
//   - Returns parent directories of found markers
//   - Silently skips paths that don't match
func discoverDirsByMarkerFile(pattern string, markerFiles []string, resourceName string) ([]string, error) {
	if pattern == "" {
		return nil, fmt.Errorf("pattern cannot be empty")
	}

	originalPattern := pattern
	pattern = filepath.Clean(pattern)

	// Check if explicit path vs glob
	isExplicitPath := !ContainsGlob(pattern)

	if isExplicitPath {
		// Strict validation
		return validateExplicitChartDir(pattern)
	}

	// Build patterns for marker files
	var patterns []string
	if strings.HasSuffix(pattern, markerFiles[0]) || (len(markerFiles) > 1 && strings.HasSuffix(pattern, markerFiles[1])) {
		patterns = []string{pattern}
	} else if strings.HasSuffix(pattern, "*") || strings.HasSuffix(pattern, "**") || strings.Contains(pattern, "{") {
		for _, marker := range markerFiles {
			patterns = append(patterns, filepath.Join(pattern, marker))
		}
	} else {
		// Literal directory - handled by explicit path check above
		return nil, fmt.Errorf("internal error: literal directory not caught")
	}

	// Filter to directories containing marker files
	return filterDirsByMarkerFile(patterns, originalPattern)
}

// discoverSupportBundlePaths discovers Support Bundle spec files from a glob pattern.
// This is a thin wrapper around discoverYAMLsByKind for backward compatibility.
//
// Supports patterns like:
//   - "./manifests/**"             (finds all Support Bundle specs recursively)
//   - "./manifests/**/*.yaml"      (explicit YAML extension)
//   - "./k8s/{dev,prod}/**/*.yaml" (environment-specific)
//   - "./support-bundle.yaml"      (explicit file path - validated strictly)
func DiscoverSupportBundlePaths(pattern string) ([]string, error) {
	return discoverYAMLsByKind(pattern, "SupportBundle", "support bundle spec")
}
