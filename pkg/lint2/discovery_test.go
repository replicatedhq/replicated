package lint2

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverSupportBundlesFromManifests(t *testing.T) {
	// Create temporary directory with test files
	tmpDir := t.TempDir()

	// Create support bundle spec
	sbSpec := filepath.Join(tmpDir, "support-bundle.yaml")
	sbContent := `apiVersion: troubleshoot.sh/v1beta2
kind: SupportBundle
metadata:
  name: test-support-bundle
spec:
  collectors:
    - clusterInfo: {}`

	if err := os.WriteFile(sbSpec, []byte(sbContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create preflight spec (should be ignored)
	preflightSpec := filepath.Join(tmpDir, "preflight.yaml")
	preflightContent := `apiVersion: troubleshoot.sh/v1beta2
kind: Preflight
metadata:
  name: test-preflight
spec:
  collectors:
    - clusterInfo: {}`

	if err := os.WriteFile(preflightSpec, []byte(preflightContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create regular K8s manifest (should be ignored)
	deploymentSpec := filepath.Join(tmpDir, "deployment.yaml")
	deploymentContent := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  replicas: 1`

	if err := os.WriteFile(deploymentSpec, []byte(deploymentContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create non-YAML file (should be skipped)
	txtFile := filepath.Join(tmpDir, "readme.txt")
	if err := os.WriteFile(txtFile, []byte("not yaml"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		manifestGlobs []string
		wantPaths     []string
		wantErr       bool
	}{
		{
			name:          "empty manifests array",
			manifestGlobs: []string{},
			wantPaths:     []string{},
			wantErr:       false,
		},
		{
			name:          "single support bundle",
			manifestGlobs: []string{sbSpec},
			wantPaths:     []string{sbSpec},
			wantErr:       false,
		},
		{
			name:          "glob pattern matching all yaml files",
			manifestGlobs: []string{filepath.Join(tmpDir, "*.yaml")},
			wantPaths:     []string{sbSpec}, // Only support bundle, not preflight or deployment
			wantErr:       false,
		},
		{
			name:          "glob pattern with no matches",
			manifestGlobs: []string{filepath.Join(tmpDir, "nonexistent", "*.yaml")},
			wantPaths:     []string{},
			wantErr:       false,
		},
		{
			name: "multiple glob patterns with overlap",
			manifestGlobs: []string{
				filepath.Join(tmpDir, "*.yaml"),
				sbSpec, // Duplicate - should be deduplicated
			},
			wantPaths: []string{sbSpec},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paths, err := DiscoverSupportBundlesFromManifests(tt.manifestGlobs)

			if (err != nil) != tt.wantErr {
				t.Errorf("DiscoverSupportBundlesFromManifests() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(paths) != len(tt.wantPaths) {
				t.Errorf("DiscoverSupportBundlesFromManifests() returned %d paths, want %d", len(paths), len(tt.wantPaths))
				t.Logf("Got: %v", paths)
				t.Logf("Want: %v", tt.wantPaths)
				return
			}

			// Check that all expected paths are present (order-independent)
			pathMap := make(map[string]bool)
			for _, p := range paths {
				pathMap[p] = true
			}

			for _, expectedPath := range tt.wantPaths {
				if !pathMap[expectedPath] {
					t.Errorf("Expected path %s not found in results", expectedPath)
				}
			}
		})
	}
}

func TestDiscoverSupportBundlesFromManifests_MultiDocument(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create multi-document YAML with support bundle and other resources
	multiDocFile := filepath.Join(tmpDir, "multi-doc.yaml")
	multiDocContent := `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
data:
  key: value
---
apiVersion: troubleshoot.sh/v1beta2
kind: SupportBundle
metadata:
  name: embedded-support-bundle
spec:
  collectors:
    - clusterInfo: {}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  replicas: 1`

	if err := os.WriteFile(multiDocFile, []byte(multiDocContent), 0644); err != nil {
		t.Fatal(err)
	}

	paths, err := DiscoverSupportBundlesFromManifests([]string{multiDocFile})
	if err != nil {
		t.Fatalf("DiscoverSupportBundlesFromManifests() unexpected error: %v", err)
	}

	if len(paths) != 1 {
		t.Errorf("DiscoverSupportBundlesFromManifests() returned %d paths, want 1", len(paths))
		return
	}

	if paths[0] != multiDocFile {
		t.Errorf("DiscoverSupportBundlesFromManifests() path = %s, want %s", paths[0], multiDocFile)
	}
}

func TestDiscoverSupportBundlesFromManifests_InvalidYAML(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create invalid YAML file (should be skipped, not error)
	invalidFile := filepath.Join(tmpDir, "invalid.yaml")
	invalidContent := `this is not
  valid: yaml: syntax
    - broken`

	if err := os.WriteFile(invalidFile, []byte(invalidContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create valid support bundle
	validFile := filepath.Join(tmpDir, "valid.yaml")
	validContent := `apiVersion: troubleshoot.sh/v1beta2
kind: SupportBundle
metadata:
  name: valid`

	if err := os.WriteFile(validFile, []byte(validContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Should skip invalid file and return valid one
	paths, err := DiscoverSupportBundlesFromManifests([]string{filepath.Join(tmpDir, "*.yaml")})
	if err != nil {
		t.Fatalf("DiscoverSupportBundlesFromManifests() unexpected error: %v", err)
	}

	if len(paths) != 1 {
		t.Errorf("DiscoverSupportBundlesFromManifests() returned %d paths, want 1 (invalid should be skipped)", len(paths))
		return
	}

	if paths[0] != validFile {
		t.Errorf("DiscoverSupportBundlesFromManifests() path = %s, want %s", paths[0], validFile)
	}
}

func TestDiscoverSupportBundlesFromManifests_SubdirectoryGlob(t *testing.T) {
	// Create nested directory structure
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "manifests")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create support bundle in subdirectory
	sbSpec := filepath.Join(subDir, "support-bundle.yaml")
	sbContent := `apiVersion: troubleshoot.sh/v1beta2
kind: SupportBundle
metadata:
  name: subdirectory-sb`

	if err := os.WriteFile(sbSpec, []byte(sbContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Test recursive glob pattern
	paths, err := DiscoverSupportBundlesFromManifests([]string{filepath.Join(tmpDir, "**", "*.yaml")})
	if err != nil {
		t.Fatalf("DiscoverSupportBundlesFromManifests() unexpected error: %v", err)
	}

	if len(paths) != 1 {
		t.Errorf("DiscoverSupportBundlesFromManifests() returned %d paths, want 1", len(paths))
		return
	}

	if paths[0] != sbSpec {
		t.Errorf("DiscoverSupportBundlesFromManifests() path = %s, want %s", paths[0], sbSpec)
	}
}

func TestDiscoverSupportBundlesFromManifests_YmlExtension(t *testing.T) {
	// Test that .yml extension is also supported (not just .yaml)
	tmpDir := t.TempDir()

	// Create support bundle with .yml extension
	sbSpec := filepath.Join(tmpDir, "support-bundle.yml")
	sbContent := `apiVersion: troubleshoot.sh/v1beta2
kind: SupportBundle
metadata:
  name: test-yml-extension`

	if err := os.WriteFile(sbSpec, []byte(sbContent), 0644); err != nil {
		t.Fatal(err)
	}

	paths, err := DiscoverSupportBundlesFromManifests([]string{filepath.Join(tmpDir, "*.yml")})
	if err != nil {
		t.Fatalf("DiscoverSupportBundlesFromManifests() unexpected error: %v", err)
	}

	if len(paths) != 1 {
		t.Errorf("DiscoverSupportBundlesFromManifests() returned %d paths, want 1", len(paths))
		return
	}

	if paths[0] != sbSpec {
		t.Errorf("DiscoverSupportBundlesFromManifests() path = %s, want %s", paths[0], sbSpec)
	}
}

func TestDiscoverSupportBundlesFromManifests_DirectoryWithYamlExtension(t *testing.T) {
	// Test that directories with .yaml extension are skipped
	tmpDir := t.TempDir()

	// Create a directory with .yaml extension
	yamlDir := filepath.Join(tmpDir, "not-a-file.yaml")
	if err := os.MkdirAll(yamlDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a valid support bundle file
	sbSpec := filepath.Join(tmpDir, "valid-bundle.yaml")
	sbContent := `apiVersion: troubleshoot.sh/v1beta2
kind: SupportBundle
metadata:
  name: valid`

	if err := os.WriteFile(sbSpec, []byte(sbContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Should skip directory and only return the file
	paths, err := DiscoverSupportBundlesFromManifests([]string{filepath.Join(tmpDir, "*.yaml")})
	if err != nil {
		t.Fatalf("DiscoverSupportBundlesFromManifests() unexpected error: %v", err)
	}

	if len(paths) != 1 {
		t.Errorf("DiscoverSupportBundlesFromManifests() returned %d paths, want 1 (directory should be skipped)", len(paths))
		return
	}

	if paths[0] != sbSpec {
		t.Errorf("DiscoverSupportBundlesFromManifests() path = %s, want %s", paths[0], sbSpec)
	}
}

func TestIsSupportBundleSpec(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name: "valid support bundle",
			content: `apiVersion: troubleshoot.sh/v1beta2
kind: SupportBundle
metadata:
  name: test`,
			want: true,
		},
		{
			name: "preflight spec",
			content: `apiVersion: troubleshoot.sh/v1beta2
kind: Preflight
metadata:
  name: test`,
			want: false,
		},
		{
			name: "deployment",
			content: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test`,
			want: false,
		},
		{
			name: "multi-document with support bundle",
			content: `apiVersion: v1
kind: ConfigMap
---
apiVersion: troubleshoot.sh/v1beta2
kind: SupportBundle
metadata:
  name: test`,
			want: true,
		},
		{
			name: "multi-document without support bundle",
			content: `apiVersion: v1
kind: ConfigMap
---
apiVersion: apps/v1
kind: Deployment`,
			want: false,
		},
		{
			name:    "empty file",
			content: "",
			want:    false,
		},
		{
			name:    "invalid yaml",
			content: "this is: not: valid: yaml:",
			want:    false,
		},
		{
			name: "triple dash in string content",
			content: `apiVersion: troubleshoot.sh/v1beta2
kind: SupportBundle
metadata:
  name: test
  description: "This string contains --- which should not be treated as document separator"
spec:
  collectors: []`,
			want: true,
		},
		{
			name: "triple dash in multiline string",
			content: `apiVersion: v1
kind: ConfigMap
data:
  script: |
    #!/bin/bash
    # This is a comment
    ---
    # The above should not be treated as separator
---
apiVersion: troubleshoot.sh/v1beta2
kind: SupportBundle
metadata:
  name: test`,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file with content
			tmpFile := filepath.Join(tmpDir, "test.yaml")
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpFile)

			got, err := isSupportBundleSpec(tmpFile)
			if err != nil && tt.want {
				t.Errorf("isSupportBundleSpec() unexpected error: %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("isSupportBundleSpec() = %v, want %v", got, tt.want)
			}
		})
	}
}
