package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestResolveLatestVersion tests that the resolver can fetch the latest version from GitHub
func TestResolveLatestVersion(t *testing.T) {
	tests := []struct {
		name       string
		tool       string
		wantErr    bool
		errMessage string
	}{
		{
			name:    "resolve latest helm version",
			tool:    ToolHelm,
			wantErr: false,
		},
		{
			name:    "resolve latest preflight version",
			tool:    ToolPreflight,
			wantErr: false,
		},
		{
			name:    "resolve latest support-bundle version",
			tool:    ToolSupportBundle,
			wantErr: false,
		},
		{
			name:       "unknown tool should error",
			tool:       "unknown-tool",
			wantErr:    true,
			errMessage: "failed to get latest version for unknown-tool: unknown tool: unknown-tool",
		},
	}

	ctx := context.Background()
	resolver := NewResolver()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := resolver.ResolveLatestVersion(ctx, tt.tool)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ResolveLatestVersion() expected error but got none")
				} else if tt.errMessage != "" && err.Error() != tt.errMessage {
					t.Errorf("ResolveLatestVersion() error = %v, want %v", err, tt.errMessage)
				}
				return
			}

			if err != nil {
				t.Errorf("ResolveLatestVersion() unexpected error = %v", err)
				return
			}

			if version == "" {
				t.Error("ResolveLatestVersion() returned empty version")
			}

			// Version should not be "latest" - it should be resolved
			if version == "latest" {
				t.Error("ResolveLatestVersion() returned 'latest' instead of actual version")
			}

			t.Logf("Resolved %s to version %s", tt.tool, version)
		})
	}
}

// TestIsCached tests the cache detection logic
func TestIsCached(t *testing.T) {
	// Create a temporary cache directory
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	if runtime.GOOS == "windows" {
		originalHome = os.Getenv("USERPROFILE")
		os.Setenv("USERPROFILE", tmpDir)
	} else {
		os.Setenv("HOME", tmpDir)
	}
	defer func() {
		if runtime.GOOS == "windows" {
			os.Setenv("USERPROFILE", originalHome)
		} else {
			os.Setenv("HOME", originalHome)
		}
	}()

	// Create a fake cached tool
	osArch := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	binaryName := "helm"
	if runtime.GOOS == "windows" {
		binaryName = "helm.exe"
	}

	cachedPath := filepath.Join(tmpDir, ".replicated", "tools", "helm", "3.14.4", osArch, binaryName)
	if err := os.MkdirAll(filepath.Dir(cachedPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cachedPath, []byte("fake binary"), 0755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		tool    string
		version string
		want    bool
	}{
		{
			name:    "cached tool should be found",
			tool:    ToolHelm,
			version: "3.14.4",
			want:    true,
		},
		{
			name:    "uncached version should not be found",
			tool:    ToolHelm,
			version: "3.13.0",
			want:    false,
		},
		{
			name:    "uncached tool should not be found",
			tool:    ToolPreflight,
			version: "0.123.9",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cached, err := IsCached(tt.tool, tt.version)
			if err != nil {
				t.Fatalf("IsCached() unexpected error = %v", err)
			}

			if cached != tt.want {
				t.Errorf("IsCached() = %v, want %v", cached, tt.want)
			}
		})
	}
}

// TestResolveWithCache tests that Resolve uses cached tools when available
func TestResolveWithCache(t *testing.T) {
	// Create a temporary cache directory
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	if runtime.GOOS == "windows" {
		originalHome = os.Getenv("USERPROFILE")
		os.Setenv("USERPROFILE", tmpDir)
	} else {
		os.Setenv("HOME", tmpDir)
	}
	defer func() {
		if runtime.GOOS == "windows" {
			os.Setenv("USERPROFILE", originalHome)
		} else {
			os.Setenv("HOME", originalHome)
		}
	}()

	// Create a fake cached tool
	osArch := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	binaryName := "helm"
	if runtime.GOOS == "windows" {
		binaryName = "helm.exe"
	}

	cachedPath := filepath.Join(tmpDir, ".replicated", "tools", "helm", "3.14.4", osArch, binaryName)
	if err := os.MkdirAll(filepath.Dir(cachedPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cachedPath, []byte("fake cached binary"), 0755); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	resolver := NewResolver()

	// Test resolving a cached tool
	toolPath, err := resolver.Resolve(ctx, ToolHelm, "3.14.4")
	if err != nil {
		t.Fatalf("Resolve() unexpected error = %v", err)
	}

	if toolPath != cachedPath {
		t.Errorf("Resolve() returned path %s, want %s", toolPath, cachedPath)
	}

	// Verify the tool exists at the returned path
	if _, err := os.Stat(toolPath); err != nil {
		t.Errorf("Tool not found at returned path %s: %v", toolPath, err)
	}

	// Read the file to verify it's our cached version
	content, err := os.ReadFile(toolPath)
	if err != nil {
		t.Fatalf("Failed to read tool at %s: %v", toolPath, err)
	}

	if string(content) != "fake cached binary" {
		t.Error("Resolve() should have returned the cached binary without downloading")
	}
}

// TestGetToolPath tests the path construction logic
func TestGetToolPath(t *testing.T) {
	// Set a known HOME for testing
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	if runtime.GOOS == "windows" {
		originalHome = os.Getenv("USERPROFILE")
		os.Setenv("USERPROFILE", tmpDir)
	} else {
		os.Setenv("HOME", tmpDir)
	}
	defer func() {
		if runtime.GOOS == "windows" {
			os.Setenv("USERPROFILE", originalHome)
		} else {
			os.Setenv("HOME", originalHome)
		}
	}()

	tests := []struct {
		name    string
		tool    string
		version string
		want    string
	}{
		{
			name:    "helm path",
			tool:    ToolHelm,
			version: "3.14.4",
			want: filepath.Join(tmpDir, ".replicated", "tools", "helm", "3.14.4",
				fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH),
				"helm"+func() string {
					if runtime.GOOS == "windows" {
						return ".exe"
					}
					return ""
				}()),
		},
		{
			name:    "preflight path",
			tool:    ToolPreflight,
			version: "0.123.9",
			want: filepath.Join(tmpDir, ".replicated", "tools", "preflight", "0.123.9",
				fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH),
				"preflight"+func() string {
					if runtime.GOOS == "windows" {
						return ".exe"
					}
					return ""
				}()),
		},
		{
			name:    "support-bundle path",
			tool:    ToolSupportBundle,
			version: "0.123.9",
			want: filepath.Join(tmpDir, ".replicated", "tools", "support-bundle", "0.123.9",
				fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH),
				"support-bundle"+func() string {
					if runtime.GOOS == "windows" {
						return ".exe"
					}
					return ""
				}()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetToolPath(tt.tool, tt.version)
			if err != nil {
				t.Fatalf("GetToolPath() unexpected error = %v", err)
			}

			if got != tt.want {
				t.Errorf("GetToolPath() = %s, want %s", got, tt.want)
			}
		})
	}
}

// TestUnknownTool tests that unknown tools are properly rejected
func TestUnknownTool(t *testing.T) {
	ctx := context.Background()
	resolver := NewResolver()

	// Test with unknown tool
	_, err := resolver.Resolve(ctx, "unknown-tool", "1.0.0")
	if err == nil {
		t.Error("Resolve() should have returned error for unknown tool")
	}

	// The error should contain "unknown tool: unknown-tool"
	expectedErrorSubstring := "unknown tool: unknown-tool"
	if !strings.Contains(err.Error(), expectedErrorSubstring) {
		t.Errorf("Resolve() error = %v, should contain %v", err, expectedErrorSubstring)
	}
}
