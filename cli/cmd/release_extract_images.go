package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/imageextract"
	"github.com/spf13/cobra"
)

func (r *runners) InitReleaseExtractImages(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "extract-images --yaml-dir DIRECTORY | --chart CHART_PATH",
		Short: "Extract container image references from Kubernetes manifests or Helm charts",
		Long: `Extract all container image references from Kubernetes manifests or Helm charts locally.

This command extracts image reference strings (like "nginx:1.19", "postgres:14") 
from YAML files without making any network calls or downloading images.`,
		Example: `  # Extract from manifest directory
  replicated release extract-images --yaml-dir ./manifests

  # Extract from Helm chart with custom values
  replicated release extract-images --chart ./mychart.tgz --values prod-values.yaml

  # JSON output for scripting
  replicated release extract-images --yaml-dir ./manifests -o json

  # Simple list for piping
  replicated release extract-images --yaml-dir ./manifests -o list`,
	}

	cmd.Flags().StringVar(&r.args.extractImagesYamlDir, "yaml-dir", "", "Directory containing Kubernetes manifests")
	cmd.Flags().StringVar(&r.args.extractImagesChart, "chart", "", "Helm chart file (.tgz) or directory")
	cmd.Flags().StringSliceVar(&r.args.extractImagesValues, "values", nil, "Values files for Helm rendering (can specify multiple)")
	cmd.Flags().StringSliceVar(&r.args.extractImagesSet, "set", nil, "Set values on command line (can specify multiple, format: key=value)")
	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "Output format: table, json, or list")
	cmd.Flags().BoolVar(&r.args.extractImagesShowDuplicates, "show-duplicates", false, "Show all occurrences instead of unique images only")
	cmd.Flags().BoolVar(&r.args.extractImagesNoWarnings, "no-warnings", false, "Suppress warnings about image references")
	cmd.Flags().StringVar(&r.args.extractImagesNamespace, "namespace", "default", "Default namespace for Helm rendering")

	cmd.RunE = r.releaseExtractImages
	parent.AddCommand(cmd)
}

func (r *runners) releaseExtractImages(cmd *cobra.Command, args []string) error {
	// Display deprecation warning
	deprecationMsg := `⚠️  WARNING: This command is deprecated and will be removed in a future version.
   Please use: replicated lint --verbose
   
   Note: The new 'replicated lint' command extracts images from Helm charts 
   defined in your .replicated config file.

`
	fmt.Fprint(r.w, deprecationMsg)
	r.w.Flush()

	// Validate inputs
	if r.args.extractImagesYamlDir == "" && r.args.extractImagesChart == "" {
		return errors.New("either --yaml-dir or --chart must be specified")
	}

	if r.args.extractImagesYamlDir != "" && r.args.extractImagesChart != "" {
		return errors.New("cannot specify both --yaml-dir and --chart")
	}

	// Validate output format
	validFormats := map[string]bool{"table": true, "json": true, "list": true}
	if !validFormats[r.outputFormat] {
		return fmt.Errorf("invalid output format %q, must be one of: table, json, list", r.outputFormat)
	}

	// Prepare options
	opts := imageextract.Options{
		HelmValuesFiles:   r.args.extractImagesValues,
		HelmValues:        parseSetValues(r.args.extractImagesSet),
		Namespace:         r.args.extractImagesNamespace,
		IncludeDuplicates: r.args.extractImagesShowDuplicates,
		NoWarnings:        r.args.extractImagesNoWarnings,
	}

	// Create extractor
	extractor := imageextract.NewExtractor()
	ctx := context.Background()

	// Extract images
	var result *imageextract.Result
	var err error

	if r.args.extractImagesYamlDir != "" {
		result, err = extractor.ExtractFromDirectory(ctx, r.args.extractImagesYamlDir, opts)
	} else {
		result, err = extractor.ExtractFromChart(ctx, r.args.extractImagesChart, opts)
	}

	if err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	// Print results
	return print.Images(r.outputFormat, r.w, result)
}

// parseSetValues parses --set flags into a map
// Format: key=value or key.nested=value
func parseSetValues(setValues []string) map[string]interface{} {
	result := make(map[string]interface{})

	for _, kv := range setValues {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		// Handle nested keys (e.g., image.repository=nginx)
		keys := strings.Split(key, ".")
		current := result

		for i, k := range keys {
			if i == len(keys)-1 {
				// Last key - set the value
				current[k] = value
			} else {
				// Intermediate key - create nested map
				if _, ok := current[k]; !ok {
					current[k] = make(map[string]interface{})
				}
				if nested, ok := current[k].(map[string]interface{}); ok {
					current = nested
				}
			}
		}
	}

	return result
}
