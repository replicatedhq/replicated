package imageextract

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/distribution/reference"
)

// deduplicateImages removes duplicate image references and optionally excludes specified images.
// Ported from airgap-builder/pkg/builder/images.go lines 827-839
func deduplicateImages(allImages []string, excludedImages []string) []string {
	seenImages := make(map[string]bool)

	// Add all images to map
	for _, image := range allImages {
		if image != "" && !seenImages[image] {
			seenImages[image] = true
		}
	}

	// Remove excluded images
	for _, excludedImage := range excludedImages {
		if seenImages[excludedImage] {
			delete(seenImages, excludedImage)
		}
	}

	// Convert back to slice
	deduplicatedImages := []string{}
	for image := range seenImages {
		deduplicatedImages = append(deduplicatedImages, image)
	}

	// Sort for consistent output
	sort.Strings(deduplicatedImages)
	return deduplicatedImages
}

// parseImageRef parses an image reference into its components.
func parseImageRef(imageStr string) ImageRef {
	result := ImageRef{
		Raw: imageStr,
	}

	// Remove HTTP/HTTPS prefix if present
	imageStr = strings.TrimPrefix(strings.TrimPrefix(imageStr, "http://"), "https://")

	// Try to parse using Docker's reference library
	named, err := reference.ParseNormalizedNamed(imageStr)
	if err != nil {
		// Return what we can
		return result
	}

	result.Registry = reference.Domain(named)
	result.Repository = reference.Path(named)

	if tagged, ok := named.(reference.Tagged); ok {
		result.Tag = tagged.Tag()
	} else {
		result.Tag = "latest"
	}

	if digested, ok := named.(reference.Digested); ok {
		result.Digest = digested.Digest().String()
	}

	return result
}

// generateWarnings creates warnings for problematic image references.
func generateWarnings(img ImageRef) []Warning {
	var warnings []Warning
	src := &img.Sources[0]

	if img.Tag == "latest" {
		warnings = append(warnings, Warning{
			Image:   img.Raw,
			Type:    WarningLatestTag,
			Message: "Image uses 'latest' tag which is not recommended for production",
			Source:  src,
		})
	}

	if img.Tag == "" || (!strings.Contains(img.Raw, ":") && !strings.Contains(img.Raw, "@")) {
		warnings = append(warnings, Warning{
			Image:   img.Raw,
			Type:    WarningNoTag,
			Message: "Image has no tag specified",
			Source:  src,
		})
	}

	if strings.HasPrefix(img.Raw, "http://") {
		warnings = append(warnings, Warning{
			Image:   img.Raw,
			Type:    WarningInsecure,
			Message: "Image uses insecure HTTP registry",
			Source:  src,
		})
	}

	if img.Registry == "docker.io" && !strings.Contains(img.Raw, ".") && !strings.Contains(img.Raw, "/") {
		warnings = append(warnings, Warning{
			Image:   img.Raw,
			Type:    WarningUnqualified,
			Message: "Image reference is unqualified (no registry specified)",
			Source:  src,
		})
	}

	return warnings
}

// isYAMLFile checks if a file has a YAML extension.
func isYAMLFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yaml" || ext == ".yml"
}
