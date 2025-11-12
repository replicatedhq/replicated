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

// patternTargetsHiddenPath checks if a user pattern explicitly targets a hidden path.
// Returns true if the pattern starts with or contains a hidden directory component.
// Examples:
//   - "./.git/**" -> true (explicitly targets .git)
//   - "/tmp/test/.git/**" -> true (explicitly targets .git)
//   - "./.github/workflows/**" -> true (explicitly targets .github)
//   - "./**" -> false (broad pattern, not targeting hidden)
//   - "./charts/**" -> false (normal path)
func patternTargetsHiddenPath(pattern string) bool {
	// Clean and normalize the pattern
	cleanPattern := filepath.Clean(pattern)
	cleanPattern = filepath.ToSlash(cleanPattern)

	// Remove leading "./" for easier parsing (handles relative paths)
	cleanPattern = strings.TrimPrefix(cleanPattern, "./")

	// Split into parts
	parts := strings.Split(cleanPattern, "/")

	// Check if any non-wildcard part is a hidden directory
	for _, part := range parts {
		// Skip wildcards
		if part == "*" || part == "**" {
			continue
		}

		// Skip empty parts (from double slashes or leading slash)
		if part == "" {
			continue
		}

		// Check if this part is a hidden directory
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

// hasGVK checks if a YAML file contains a resource matching the specified GroupVersionKind (GVK).
// Handles multi-document YAML files properly using yaml.NewDecoder.
// For files with syntax errors, falls back to simple regex matching.
//
// Parameters:
//   - path: Path to the YAML file
//   - group: API group (e.g., "embeddedcluster.replicated.com"). Empty string matches any group.
//   - version: API version (e.g., "v1beta1"). Empty string matches any version.
//   - kind: Resource kind (e.g., "Config", "Preflight"). Required, cannot be empty.
//
// The apiVersion field is parsed as "<group>/<version>" or just "<version>" (if no group).
// Empty group or version parameters act as wildcards (match any).
func hasGVK(path string, group string, version string, kind string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	// Use yaml.Decoder for proper multi-document YAML parsing
	decoder := yaml.NewDecoder(bytes.NewReader(data))

	// Iterate through all documents in the file
	for {
		// Parse apiVersion and kind fields
		var doc struct {
			APIVersion string `yaml:"apiVersion"`
			Kind       string `yaml:"kind"`
		}

		err := decoder.Decode(&doc)
		if err != nil {
			if err == io.EOF {
				// Reached end of file - no more documents
				break
			}
			// Parse error - file is malformed
			// Fall back to regex matching for kind only (can't easily regex match apiVersion)
			if group == "" && version == "" {
				// Only kind matching - use regex fallback
				pattern := fmt.Sprintf(`(?m)^kind:\s+%s\s*$`, regexp.QuoteMeta(kind))
				matched, _ := regexp.Match(pattern, data)
				return matched, nil
			}
			// Can't match group/version with malformed YAML - skip
			return false, nil
		}

		// Check if kind matches
		if doc.Kind != kind {
			continue
		}

		// Kind matches - now check group and version
		// Parse apiVersion into group and version components
		var docGroup, docVersion string
		if strings.Contains(doc.APIVersion, "/") {
			// Format: "group/version" (e.g., "embeddedcluster.replicated.com/v1beta1")
			parts := strings.SplitN(doc.APIVersion, "/", 2)
			docGroup = parts[0]
			docVersion = parts[1]
		} else {
			// Format: "version" only (e.g., "v1" for core Kubernetes resources)
			docGroup = ""
			docVersion = doc.APIVersion
		}

		// Check group match (empty = match any)
		if group != "" && docGroup != group {
			continue
		}

		// Check version match (empty = match any)
		if version != "" && docVersion != version {
			continue
		}

		// All criteria matched
		return true, nil
	}

	return false, nil
}

// DiscoverPreflightPaths discovers Preflight spec files from a glob pattern.
//
// Supports patterns like:
//   - "./preflights/**"           (finds all Preflight specs recursively)
//   - "./preflights/**/*.yaml"    (explicit YAML extension)
//   - "./k8s/{dev,prod}/**/*.yaml" (environment-specific)
//   - "./preflight.yaml"          (explicit file path - validated strictly)
func DiscoverPreflightPaths(pattern string) ([]string, error) {
	return discoverYAMLsByGVK(pattern, "", "", "Preflight", "preflight spec")
}

// DiscoverHelmChartPaths discovers HelmChart manifest files from a glob pattern.
//
// Supports patterns like:
//   - "./manifests/**"           (finds all HelmChart manifests recursively)
//   - "./manifests/**/*.yaml"    (explicit YAML extension)
//   - "./k8s/{dev,prod}/**/*.yaml" (environment-specific)
//   - "./helmchart.yaml"         (explicit file path - validated strictly)
func DiscoverHelmChartPaths(pattern string) ([]string, error) {
	return discoverYAMLsByGVK(pattern, "", "", "HelmChart", "HelmChart manifest")
}

// DiscoverEmbeddedClusterPaths discovers Embedded Cluster config files from a glob pattern.
// Filters by both kind and group to distinguish from KOTS Config (which also uses kind: Config).
//
// Matches:
//   - kind: Config
//   - group: embeddedcluster.replicated.com (any version: v1beta1, v1beta2, v1, etc.)
//
// Supports patterns like:
//   - "./manifests/**"           (finds all EC configs recursively)
//   - "./manifests/**/*.yaml"    (explicit YAML extension)
//   - "./k8s/{dev,prod}/**/*.yaml" (environment-specific)
//   - "./ec-config.yaml"         (explicit file path - validated strictly)
func DiscoverEmbeddedClusterPaths(pattern string) ([]string, error) {
	// Filter by group to distinguish from KOTS Config (kots.io/v1beta1)
	// Match any version within the embeddedcluster.replicated.com group
	return discoverYAMLsByGVK(pattern, "embeddedcluster.replicated.com", "", "Config", "embedded cluster config")
}

// DiscoverKotsPaths discovers KOTS Config files from a glob pattern.
// Filters by both kind and group to distinguish from Embedded Cluster Config (which also uses kind: Config).
//
// Matches:
//   - kind: Config
//   - group: kots.io (any version: v1beta1, v1beta2, v1, etc.)
//
// Supports patterns like:
//   - "./manifests/**"           (finds all KOTS configs recursively)
//   - "./manifests/**/*.yaml"    (explicit YAML extension)
//   - "./kots-config.yaml"       (explicit file path - validated strictly)
func DiscoverKotsPaths(pattern string) ([]string, error) {
	// Filter by group to distinguish from Embedded Cluster Config
	// Match any version within the kots.io group
	return discoverYAMLsByGVK(pattern, "kots.io", "", "Config", "kots config")
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

// validateExplicitYAMLFile validates a single YAML file path and checks its GVK.
// Returns the path in a slice for consistency with discovery functions.
// Returns error if file doesn't exist, isn't a file, has wrong extension, or doesn't contain matching GVK.
func validateExplicitYAMLFile(path, group, version, kind, resourceName string) ([]string, error) {
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

	// Check GVK
	hasTargetGVK, err := hasGVK(path, group, version, kind)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	if !hasTargetGVK {
		return nil, fmt.Errorf("file does not contain kind: %s (not a valid %s)", kind, resourceName)
	}

	return []string{path}, nil
}

// filterYAMLFilesByGVK expands glob patterns and filters to files with matching GVK.
// Silently skips files that can't be read or don't have the target GVK.
// Optionally filters by gitignore if checker is provided.
// Optionally skips hidden paths unless skipHidden is false (explicit bypass).
func filterYAMLFilesByGVK(patterns []string, originalPattern, group, version, kind string, gitignoreChecker *GitignoreChecker, skipHidden bool) ([]string, error) {
	var resultPaths []string
	seenPaths := make(map[string]bool)

	for _, p := range patterns {
		// Use gitignore filtering if checker provided
		var matches []string
		var err error
		if gitignoreChecker != nil {
			matches, err = GlobFiles(p, WithGitignoreChecker(gitignoreChecker))
		} else {
			matches, err = GlobFiles(p)
		}
		if err != nil {
			return nil, fmt.Errorf("expanding pattern %s: %w (from user pattern: %s)", p, err, originalPattern)
		}

		for _, path := range matches {
			// Skip hidden paths (.git, .github, .vscode, etc.) by default
			// But allow bypass if user explicitly specified a hidden path pattern
			if skipHidden && isHiddenPath(path) {
				continue
			}

			// Skip duplicates
			if seenPaths[path] {
				continue
			}
			seenPaths[path] = true

			// Check GVK
			hasTargetGVK, err := hasGVK(path, group, version, kind)
			if err != nil {
				// Skip files we can't read
				continue
			}

			if hasTargetGVK {
				resultPaths = append(resultPaths, path)
			}
		}
	}

	return resultPaths, nil
}

// discoverYAMLsByGVK discovers YAML files containing resources matching a GroupVersionKind (GVK) from a pattern.
// Handles both explicit file paths (strict validation) and glob patterns (lenient filtering).
// Respects gitignore unless the pattern explicitly references a gitignored path.
//
// Parameters:
//   - pattern: Glob pattern or explicit file path
//   - group: API group (empty = match any)
//   - version: API version (empty = match any)
//   - kind: Resource kind (required)
//   - resourceName: Descriptive name for error messages
//
// For explicit paths:
//   - Validates file exists, is a file, has .yaml/.yml extension
//   - Checks if file contains the specified GVK
//   - Returns error if any validation fails (fail loudly)
//
// For glob patterns:
//   - Expands pattern to find all YAML files
//   - Filters to only files containing the specified GVK
//   - Silently skips files that don't match (allows mixed directories)
//   - Respects gitignore unless pattern explicitly includes gitignored path
func discoverYAMLsByGVK(pattern, group, version, kind, resourceName string) ([]string, error) {
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
		return validateExplicitYAMLFile(pattern, group, version, kind, resourceName)
	}

	// Create gitignore checker (returns nil if not in git repo or no gitignore)
	gitignoreChecker, _ := NewGitignoreChecker(".")

	// Check if pattern explicitly bypasses gitignore
	var checkerToUse *GitignoreChecker
	if gitignoreChecker != nil && gitignoreChecker.PathMatchesIgnoredPattern(pattern) {
		// Pattern explicitly references gitignored path - bypass gitignore
		checkerToUse = nil
	} else {
		// Use gitignore filtering
		checkerToUse = gitignoreChecker
	}

	// Check if pattern explicitly targets hidden paths
	// If yes, allow bypass of hidden path filtering (like gitignore bypass)
	skipHidden := !patternTargetsHiddenPath(pattern)

	// Glob pattern - build search patterns
	patterns, err := buildYAMLPatterns(pattern)
	if err != nil {
		return nil, err
	}

	// Lenient filtering
	return filterYAMLFilesByGVK(patterns, originalPattern, group, version, kind, checkerToUse, skipHidden)
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
// Optionally filters by gitignore if checker is provided.
// Optionally skips hidden paths unless skipHidden is false (explicit bypass).
func filterDirsByMarkerFile(patterns []string, originalPattern string, gitignoreChecker *GitignoreChecker, skipHidden bool) ([]string, error) {
	var chartDirs []string
	seenDirs := make(map[string]bool)

	for _, p := range patterns {
		// Use gitignore filtering if checker provided
		var matches []string
		var err error
		if gitignoreChecker != nil {
			matches, err = GlobFiles(p, WithGitignoreChecker(gitignoreChecker))
		} else {
			matches, err = GlobFiles(p)
		}
		if err != nil {
			return nil, fmt.Errorf("expanding pattern %s: %w (from user pattern: %s)", p, err, originalPattern)
		}

		for _, markerPath := range matches {
			// Skip hidden paths (.git, .github, .vscode, etc.) by default
			// But allow bypass if user explicitly specified a hidden path pattern
			if skipHidden && isHiddenPath(markerPath) {
				continue
			}

			chartDir := filepath.Dir(markerPath)

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
// Respects gitignore unless the pattern explicitly references a gitignored path.
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
//   - Respects gitignore unless pattern explicitly includes gitignored path
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

	// Create gitignore checker (returns nil if not in git repo or no gitignore)
	gitignoreChecker, _ := NewGitignoreChecker(".")

	// Check if pattern explicitly bypasses gitignore
	var checkerToUse *GitignoreChecker
	if gitignoreChecker != nil && gitignoreChecker.PathMatchesIgnoredPattern(pattern) {
		// Pattern explicitly references gitignored path - bypass gitignore
		checkerToUse = nil
	} else {
		// Use gitignore filtering
		checkerToUse = gitignoreChecker
	}

	// Check if pattern explicitly targets hidden paths
	// If yes, allow bypass of hidden path filtering (like gitignore bypass)
	skipHidden := !patternTargetsHiddenPath(pattern)

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
	return filterDirsByMarkerFile(patterns, originalPattern, checkerToUse, skipHidden)
}

// DiscoverSupportBundlePaths discovers Support Bundle spec files from a glob pattern.
//
// Supports patterns like:
//   - "./manifests/**"             (finds all Support Bundle specs recursively)
//   - "./manifests/**/*.yaml"      (explicit YAML extension)
//   - "./k8s/{dev,prod}/**/*.yaml" (environment-specific)
//   - "./support-bundle.yaml"      (explicit file path - validated strictly)
func DiscoverSupportBundlePaths(pattern string) ([]string, error) {
	return discoverYAMLsByGVK(pattern, "", "", "SupportBundle", "support bundle spec")
}
