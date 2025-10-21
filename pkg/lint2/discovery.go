package lint2

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

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
		// Expand glob pattern to concrete file paths (files only)
		matches, err := GlobFiles(globPattern)
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
			// Parse error - file is malformed, skip it gracefully
			// Cannot continue decoding as decoder is in error state
			return false, nil
		}

		// Check if this document is a SupportBundle
		if kindDoc.Kind == "SupportBundle" {
			return true, nil
		}
	}

	return false, nil
}
