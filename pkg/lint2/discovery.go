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
		paths, err := discoverSupportBundlePaths(pattern)
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
// It finds all Chart.yaml or Chart.yml files matching the pattern, then returns
// their parent directories (the actual chart directories).
//
// Supports patterns like:
//   - "./charts/**"              (finds all charts recursively)
//   - "./charts/{app,api}/**"    (finds charts in specific subdirectories)
//   - "./pkg/**/Chart.yaml"      (explicit Chart.yaml pattern)
func discoverChartPaths(pattern string) ([]string, error) {
	// Validate: reject empty patterns
	if pattern == "" {
		return nil, fmt.Errorf("pattern cannot be empty")
	}

	// Preserve original for error messages
	originalPattern := pattern

	// Normalize pattern: clean path to handle //, /./, /../, and trailing slashes
	// This allows "./charts/**/" to work the same as "./charts/**"
	pattern = filepath.Clean(pattern)

	var chartDirs []string
	seenDirs := make(map[string]bool) // Deduplication

	// Build patterns for both Chart.yaml and Chart.yml
	var patterns []string

	// If pattern already specifies Chart.yaml/Chart.yml, use it directly
	if strings.HasSuffix(pattern, "Chart.yaml") || strings.HasSuffix(pattern, "Chart.yml") {
		patterns = []string{pattern}
	} else if strings.HasSuffix(pattern, "*") || strings.HasSuffix(pattern, "**") || strings.Contains(pattern, "{") {
		// Pattern ends with wildcard or contains brace expansion - append Chart.yaml and Chart.yml
		patterns = []string{
			filepath.Join(pattern, "Chart.yaml"),
			filepath.Join(pattern, "Chart.yml"),
		}
	} else {
		// Pattern is a literal directory path - check if it's a chart
		isChart, err := isChartDirectory(pattern)
		if err != nil {
			return nil, fmt.Errorf("checking if %s is chart directory: %w", pattern, err)
		}
		if isChart {
			return []string{pattern}, nil
		}
		return nil, fmt.Errorf("directory %s is not a valid Helm chart (no Chart.yaml or Chart.yml found)", pattern)
	}

	// Expand patterns to find Chart.yaml files
	for _, p := range patterns {
		matches, err := GlobFiles(p)
		if err != nil {
			return nil, fmt.Errorf("expanding pattern %s: %w (from user pattern: %s)", p, err, originalPattern)
		}

		// Extract parent directories (the chart directories)
		for _, chartFile := range matches {
			chartDir := filepath.Dir(chartFile)

			// Skip hidden directories (.git, .github, etc.)
			if isHiddenPath(chartDir) {
				continue
			}

			// Deduplicate (in case both Chart.yaml and Chart.yml exist)
			if seenDirs[chartDir] {
				continue
			}
			seenDirs[chartDir] = true

			chartDirs = append(chartDirs, chartDir)
		}
	}

	return chartDirs, nil
}

// isPreflightSpec checks if a YAML file contains a Preflight kind.
// Handles multi-document YAML files properly using yaml.NewDecoder, which correctly
// handles document separators (---) even when they appear inside strings or block scalars.
// For files with syntax errors, falls back to simple string matching to detect the kind.
func isPreflightSpec(path string) (bool, error) {
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
			// Fall back to regex matching to detect if this looks like a Preflight
			// This allows invalid YAML files to still be discovered and linted
			// Use regex to match "kind: Preflight" as a complete line (not in comments/strings)
			matched, _ := regexp.Match(`(?m)^kind:\s+Preflight\s*$`, data)
			if matched {
				return true, nil
			}
			return false, nil
		}

		// Check if this document is a Preflight
		if kindDoc.Kind == "Preflight" {
			return true, nil
		}
	}

	return false, nil
}

// discoverPreflightPaths discovers Preflight spec files from a glob pattern.
// It finds all YAML files matching the pattern, then filters to only those
// containing kind: Preflight.
//
// Supports patterns like:
//   - "./preflights/**"           (finds all Preflight specs recursively)
//   - "./preflights/**/*.yaml"    (explicit YAML extension)
//   - "./k8s/{dev,prod}/**/*.yaml" (environment-specific)
func discoverPreflightPaths(pattern string) ([]string, error) {
	// Validate: reject empty patterns
	if pattern == "" {
		return nil, fmt.Errorf("pattern cannot be empty")
	}

	// Preserve original for error messages
	originalPattern := pattern

	// Normalize pattern: clean path to handle //, /./, /../, and trailing slashes
	// This allows "./preflights/**/" to work the same as "./preflights/**"
	pattern = filepath.Clean(pattern)

	var preflightPaths []string
	seenPaths := make(map[string]bool) // Deduplication

	// Build patterns to find YAML files
	var patterns []string
	if strings.HasSuffix(pattern, ".yaml") || strings.HasSuffix(pattern, ".yml") {
		// Pattern already specifies extension
		patterns = []string{pattern}
	} else if strings.HasSuffix(pattern, "/*") {
		// Single-level wildcard: replace /* with /*.yaml and /*.yml
		basePattern := strings.TrimSuffix(pattern, "/*")
		patterns = []string{
			filepath.Join(basePattern, "*.yaml"),
			filepath.Join(basePattern, "*.yml"),
		}
	} else if strings.HasSuffix(pattern, "**") || strings.Contains(pattern, "{") {
		// Recursive wildcard or brace expansion - append file patterns
		patterns = []string{
			filepath.Join(pattern, "*.yaml"),
			filepath.Join(pattern, "*.yml"),
		}
	} else {
		// Pattern might be a single file
		ext := filepath.Ext(pattern)
		if ext == ".yaml" || ext == ".yml" {
			patterns = []string{pattern}
		} else {
			return nil, fmt.Errorf("pattern must end with .yaml, .yml, *, or **")
		}
	}

	// Expand patterns to find YAML files
	for _, p := range patterns {
		matches, err := GlobFiles(p)
		if err != nil {
			return nil, fmt.Errorf("expanding pattern %s: %w (from user pattern: %s)", p, err, originalPattern)
		}

		// Filter to only Preflight specs
		for _, path := range matches {
			// Skip hidden paths (.git, .github, etc.)
			if isHiddenPath(path) {
				continue
			}

			// Skip if already processed
			if seenPaths[path] {
				continue
			}
			seenPaths[path] = true

			// Check if it's a Preflight spec
			isPreflight, err := isPreflightSpec(path)
			if err != nil {
				// Skip files we can't read or parse
				continue
			}

			if isPreflight {
				preflightPaths = append(preflightPaths, path)
			}
		}
	}

	return preflightPaths, nil
}

// isSupportBundleSpec checks if a YAML file contains a SupportBundle kind.
// Handles multi-document YAML files properly using yaml.NewDecoder, which correctly
// handles document separators (---) even when they appear inside strings or block scalars.
func isSupportBundleSpec(path string) (bool, error) {
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
			// Fall back to regex matching to detect if this looks like a SupportBundle
			// This allows invalid YAML files to still be discovered and linted (consistent with preflights)
			// Use regex to match "kind: SupportBundle" as a complete line (not in comments/strings)
			matched, _ := regexp.Match(`(?m)^kind:\s+SupportBundle\s*$`, data)
			if matched {
				return true, nil
			}
			return false, nil
		}

		// Check if this document is a SupportBundle
		if kindDoc.Kind == "SupportBundle" {
			return true, nil
		}
	}

	return false, nil
}

// discoverSupportBundlePaths discovers Support Bundle spec files from a glob pattern.
// It finds all YAML files matching the pattern, then filters to only those
// containing kind: SupportBundle.
//
// Supports patterns like:
//   - "./manifests/**"             (finds all Support Bundle specs recursively)
//   - "./manifests/**/*.yaml"      (explicit YAML extension)
//   - "./k8s/{dev,prod}/**/*.yaml" (environment-specific)
func discoverSupportBundlePaths(pattern string) ([]string, error) {
	// Validate: reject empty patterns
	if pattern == "" {
		return nil, fmt.Errorf("pattern cannot be empty")
	}

	// Preserve original for error messages
	originalPattern := pattern

	// Normalize pattern: clean path to handle //, /./, /../, and trailing slashes
	// This allows "./manifests/**/" to work the same as "./manifests/**"
	pattern = filepath.Clean(pattern)

	var supportBundlePaths []string
	seenPaths := make(map[string]bool) // Deduplication

	// Build patterns to find YAML files
	var patterns []string
	if strings.HasSuffix(pattern, ".yaml") || strings.HasSuffix(pattern, ".yml") {
		// Pattern already specifies extension
		patterns = []string{pattern}
	} else if strings.HasSuffix(pattern, "/*") {
		// Single-level wildcard: replace /* with /*.yaml and /*.yml
		basePattern := strings.TrimSuffix(pattern, "/*")
		patterns = []string{
			filepath.Join(basePattern, "*.yaml"),
			filepath.Join(basePattern, "*.yml"),
		}
	} else if strings.HasSuffix(pattern, "**") || strings.Contains(pattern, "{") {
		// Recursive wildcard or brace expansion - append file patterns
		patterns = []string{
			filepath.Join(pattern, "*.yaml"),
			filepath.Join(pattern, "*.yml"),
		}
	} else {
		// Pattern might be a single file
		ext := filepath.Ext(pattern)
		if ext == ".yaml" || ext == ".yml" {
			patterns = []string{pattern}
		} else {
			return nil, fmt.Errorf("pattern must end with .yaml, .yml, *, or **")
		}
	}

	// Expand patterns to find YAML files
	for _, p := range patterns {
		matches, err := GlobFiles(p)
		if err != nil {
			return nil, fmt.Errorf("expanding pattern %s: %w (from user pattern: %s)", p, err, originalPattern)
		}

		// Filter to only Support Bundle specs
		for _, path := range matches {
			// Skip hidden paths (.git, .github, etc.)
			if isHiddenPath(path) {
				continue
			}

			// Skip if already processed
			if seenPaths[path] {
				continue
			}
			seenPaths[path] = true

			// Check if it's a Support Bundle spec
			isSupportBundle, err := isSupportBundleSpec(path)
			if err != nil {
				// Skip files we can't read or parse
				continue
			}

			if isSupportBundle {
				supportBundlePaths = append(supportBundlePaths, path)
			}
		}
	}

	return supportBundlePaths, nil
}
