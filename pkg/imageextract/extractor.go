package imageextract

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"path/filepath"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
)

type extractor struct{}

// NewExtractor creates a new Extractor instance.
func NewExtractor() Extractor {
	return &extractor{}
}

// ExtractFromDirectory recursively processes all YAML files in a directory.
func (e *extractor) ExtractFromDirectory(ctx context.Context, dir string, opts Options) (*Result, error) {
	result := &Result{}
	allExcludedImages := []string{}

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !isYAMLFile(path) {
			return err
		}

		data, err := os.ReadFile(path)
		if err != nil {
			result.Errors = append(result.Errors, err)
			return nil
		}

		// Extract images from this file using airgap extraction logic
		images, excluded := extractImagesFromFile(data)
		allExcludedImages = append(allExcludedImages, excluded...)

		// Convert to ImageRef with source information
		for _, imgStr := range images {
			img := parseImageRef(imgStr)
			img.Sources = []Source{{
				File: path,
			}}
			result.Images = append(result.Images, img)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Deduplicate if requested
	if !opts.IncludeDuplicates {
		result.deduplicateAndExclude(allExcludedImages)
	}

	// Generate warnings
	if !opts.NoWarnings {
		for _, img := range result.Images {
			result.Warnings = append(result.Warnings, generateWarnings(img)...)
		}
	}

	return result, nil
}

// ExtractFromChart loads and renders a Helm chart, then extracts images.
func (e *extractor) ExtractFromChart(ctx context.Context, chartPath string, opts Options) (*Result, error) {
	// Load chart
	chart, err := loader.Load(chartPath)
	if err != nil {
		return nil, err
	}

	// Prepare values
	vals, err := prepareValues(opts)
	if err != nil {
		return nil, err
	}

	// Set namespace
	ns := opts.Namespace
	if ns == "" {
		ns = "default"
	}

	// Render chart
	cfg := new(action.Configuration)
	client := action.NewInstall(cfg)
	client.DryRun = true
	client.ReleaseName = "release"
	client.Namespace = ns
	client.ClientOnly = true

	validatedVals, err := chartutil.CoalesceValues(chart, vals)
	if err != nil {
		return nil, err
	}

	rel, err := client.Run(chart, validatedVals)
	if err != nil {
		return nil, err
	}

	// Collect rendered manifests
	var buf bytes.Buffer
	buf.WriteString(rel.Manifest)
	for _, hook := range rel.Hooks {
		buf.WriteString("\n---\n")
		buf.WriteString(hook.Manifest)
	}

	// Extract from rendered manifests
	return e.ExtractFromManifests(ctx, buf.Bytes(), opts)
}

// ExtractFromManifests parses raw YAML and extracts image references.
func (e *extractor) ExtractFromManifests(ctx context.Context, manifests []byte, opts Options) (*Result, error) {
	result := &Result{}

	// Extract images using airgap extraction logic
	images, excludedImages := extractImagesFromFile(manifests)

	// Convert to ImageRef
	for _, imgStr := range images {
		img := parseImageRef(imgStr)
		img.Sources = []Source{{}}
		result.Images = append(result.Images, img)
	}

	// Deduplicate if requested
	if !opts.IncludeDuplicates {
		result.deduplicateAndExclude(excludedImages)
	}

	// Generate warnings
	if !opts.NoWarnings {
		for _, img := range result.Images {
			result.Warnings = append(result.Warnings, generateWarnings(img)...)
		}
	}

	return result, nil
}

// deduplicateAndExclude removes duplicates and excluded images from the result.
func (r *Result) deduplicateAndExclude(excludedImages []string) {
	// Extract image strings
	imageStrings := make([]string, len(r.Images))
	for i, img := range r.Images {
		imageStrings[i] = img.Raw
	}

	// Deduplicate using airgap logic
	deduped := deduplicateImages(imageStrings, excludedImages)

	// Convert back to ImageRef
	newImages := make([]ImageRef, 0, len(deduped))
	for _, imgStr := range deduped {
		img := parseImageRef(imgStr)

		// Merge sources from original images
		for _, origImg := range r.Images {
			if origImg.Raw == imgStr {
				img.Sources = append(img.Sources, origImg.Sources...)
			}
		}

		newImages = append(newImages, img)
	}

	r.Images = newImages
}

// prepareValues merges values from multiple sources for Helm rendering.
func prepareValues(opts Options) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	if len(opts.HelmValuesFiles) > 0 {
		valueOpts := &values.Options{ValueFiles: opts.HelmValuesFiles}
		vals, err := valueOpts.MergeValues(getter.All(&cli.EnvSettings{}))
		if err != nil {
			return nil, err
		}
		result = vals
	}

	if opts.HelmValues != nil {
		result = mergeMaps(result, opts.HelmValues)
	}

	return result, nil
}

// mergeMaps deeply merges two maps.
func mergeMaps(a, b map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range a {
		result[k] = v
	}
	for k, v := range b {
		if existing, ok := result[k]; ok {
			if em, ok := existing.(map[string]interface{}); ok {
				if vm, ok := v.(map[string]interface{}); ok {
					result[k] = mergeMaps(em, vm)
					continue
				}
			}
		}
		result[k] = v
	}
	return result
}
