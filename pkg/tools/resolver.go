package tools

import (
	"context"
	"fmt"
	"os"
)

// Resolver resolves tool binaries, downloading and caching as needed
type Resolver struct {
	downloader *Downloader
}

// NewResolver creates a new tool resolver
func NewResolver() *Resolver {
	return &Resolver{
		downloader: NewDownloader(),
	}
}

// Resolve returns the path to a tool binary, downloading if not cached
func (r *Resolver) Resolve(ctx context.Context, name, version string) (string, error) {
	// Get cache path
	toolPath, err := GetToolPath(name, version)
	if err != nil {
		return "", fmt.Errorf("getting cache path: %w", err)
	}

	// Check if already cached
	cached, err := IsCached(name, version)
	if err != nil {
		return "", fmt.Errorf("checking cache: %w", err)
	}

	if cached {
		// Tool is cached, return immediately
		return toolPath, nil
	}

	// Not cached - download it
	fmt.Printf("Downloading %s %s...\n", name, version)
	actualVersion, err := r.downloader.Download(ctx, name, version)
	if err != nil {
		return "", fmt.Errorf("downloading %s %s: %w", name, version, err)
	}

	// If a different version was downloaded (due to fallback), get the correct path
	if actualVersion != version {
		toolPath, err = GetToolPath(name, actualVersion)
		if err != nil {
			return "", fmt.Errorf("getting cache path for actual version %s: %w", actualVersion, err)
		}
	}

	// Verify it now exists
	if _, err := os.Stat(toolPath); err != nil {
		return "", fmt.Errorf("tool not found after download: %w", err)
	}

	return toolPath, nil
}
