package tools

import (
	"context"
	"os"
	"testing"
)

// TestResolverWithInvalidVersionFallback tests that when an invalid version is
// requested, the resolver successfully falls back to the latest version and
// returns a working tool path.
//
// Currently FAILS due to bug: Download() discards the actual version used,
// so Resolver looks for the tool at the wrong path.
func TestResolverWithInvalidVersionFallback(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test that downloads from network")
	}

	ctx := context.Background()
	resolver := NewResolver()

	// Request an invalid version that will trigger fallback to latest
	invalidVersion := "99.99.99"

	// This SHOULD succeed (fallback downloads latest), but currently FAILS
	toolPath, err := resolver.Resolve(ctx, ToolHelm, invalidVersion)
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	// Verify the tool binary actually exists at the returned path
	if _, err := os.Stat(toolPath); err != nil {
		t.Fatalf("Tool not found at returned path %s: %v", toolPath, err)
	}

	t.Logf("Success! Tool found at: %s", toolPath)
}
