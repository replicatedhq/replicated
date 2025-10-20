package lint2

import (
	"fmt"
	"os"
	"path/filepath"
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

	var supportBundlePaths []string
	seenPaths := make(map[string]bool) // For deduplication

	for _, globPattern := range manifestGlobs {
		// Expand glob pattern to concrete file paths
		matches, err := filepath.Glob(globPattern)
		if err != nil {
			return nil, fmt.Errorf("expanding glob pattern %s: %w", globPattern, err)
		}

		for _, path := range matches {
			// Skip if already processed (handle overlapping globs)
			if seenPaths[path] {
				continue
			}
			seenPaths[path] = true

			// Check if file is a YAML file
			ext := filepath.Ext(path)
			if ext != ".yaml" && ext != ".yml" {
				continue
			}

			// Check if file is a regular file (not directory, symlink handled by Glob)
			info, err := os.Stat(path)
			if err != nil {
				// Skip files we can't stat (permission issues, etc.)
				continue
			}
			if info.IsDir() {
				continue
			}

			// Read and check if it's a support bundle spec
			isSupportBundle, err := isSupportBundleSpec(path)
			if err != nil {
				// Skip files we can't read or parse - linting will catch issues later
				continue
			}

			if isSupportBundle {
				supportBundlePaths = append(supportBundlePaths, path)
			}
		}
	}

	return supportBundlePaths, nil
}

// isSupportBundleSpec checks if a YAML file contains a SupportBundle kind.
// Handles multi-document YAML files (documents separated by ---).
func isSupportBundleSpec(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	// Handle multi-document YAML by splitting on document separator
	// Note: This is a simple split - proper YAML parsing would use yaml.Decoder
	// but this is sufficient for kind detection
	documents := strings.Split(string(data), "\n---")

	for _, doc := range documents {
		// Skip empty documents
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		// Parse just the kind field (lightweight)
		var kindDoc struct {
			Kind string `yaml:"kind"`
		}

		if err := yaml.Unmarshal([]byte(doc), &kindDoc); err != nil {
			// Skip documents that can't be parsed
			continue
		}

		// Check if this document is a SupportBundle
		if kindDoc.Kind == "SupportBundle" {
			return true, nil
		}
	}

	return false, nil
}
