package lint2

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// Test helpers

// createTestChart creates a minimal Chart.yaml file in the specified directory
func createTestChart(t *testing.T, dir, name string) string {
	t.Helper()
	chartDir := filepath.Join(dir, name)
	if err := os.MkdirAll(chartDir, 0755); err != nil {
		t.Fatalf("failed to create chart directory %s: %v", chartDir, err)
	}
	chartYaml := filepath.Join(chartDir, "Chart.yaml")
	content := "apiVersion: v2\nname: " + name + "\nversion: 1.0.0\n"
	if err := os.WriteFile(chartYaml, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write Chart.yaml: %v", err)
	}
	return chartDir
}

// createTestChartWithExtension creates a Chart file with specified extension (.yaml or .yml)
func createTestChartWithExtension(t *testing.T, dir, name, ext string) string {
	t.Helper()
	chartDir := filepath.Join(dir, name)
	if err := os.MkdirAll(chartDir, 0755); err != nil {
		t.Fatalf("failed to create chart directory %s: %v", chartDir, err)
	}
	chartFile := filepath.Join(chartDir, "Chart."+ext)
	content := "apiVersion: v2\nname: " + name + "\nversion: 1.0.0\n"
	if err := os.WriteFile(chartFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write Chart file: %v", err)
	}
	return chartDir
}

// createTestPreflight creates a Preflight spec YAML file
func createTestPreflight(t *testing.T, path string) string {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create directory %s: %v", dir, err)
	}
	content := `apiVersion: troubleshoot.sh/v1beta2
kind: Preflight
metadata:
  name: test
spec:
  collectors:
    - logs: {}
  analyzers:
    - textAnalyze: {}
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write preflight spec: %v", err)
	}
	return path
}

// createTestSupportBundle creates a SupportBundle spec YAML file
func createTestSupportBundle(t *testing.T, path string) string {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create directory %s: %v", dir, err)
	}
	content := `apiVersion: troubleshoot.sh/v1beta2
kind: SupportBundle
metadata:
  name: test
spec:
  collectors:
    - logs: {}
  analyzers:
    - textAnalyze: {}
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write support bundle spec: %v", err)
	}
	return path
}

// createTestK8sResource creates a K8s resource YAML file with specified kind
func createTestK8sResource(t *testing.T, path, kind string) string {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create directory %s: %v", dir, err)
	}
	content := "apiVersion: v1\nkind: " + kind + "\nmetadata:\n  name: test\nspec: {}\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write k8s resource: %v", err)
	}
	return path
}

// createMultiDocYAML creates a multi-document YAML file with specified kinds
func createMultiDocYAML(t *testing.T, path string, kinds []string) string {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create directory %s: %v", dir, err)
	}

	var content string
	for i, kind := range kinds {
		if i > 0 {
			content += "---\n"
		}
		content += "apiVersion: v1\nkind: " + kind + "\nmetadata:\n  name: test" + string(rune(i+'0')) + "\nspec: {}\n"
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write multi-doc YAML: %v", err)
	}
	return path
}

// assertPathsEqual asserts that two path slices contain the same elements (order-independent)
func assertPathsEqual(t *testing.T, got, want []string) {
	t.Helper()

	// Sort both slices for comparison
	gotSorted := make([]string, len(got))
	copy(gotSorted, got)
	sort.Strings(gotSorted)

	wantSorted := make([]string, len(want))
	copy(wantSorted, want)
	sort.Strings(wantSorted)

	if len(gotSorted) != len(wantSorted) {
		t.Errorf("path count mismatch: got %d paths, want %d paths\ngot:  %v\nwant: %v",
			len(gotSorted), len(wantSorted), gotSorted, wantSorted)
		return
	}

	for i := range gotSorted {
		if gotSorted[i] != wantSorted[i] {
			t.Errorf("path mismatch at index %d:\ngot:  %s\nwant: %s\nall got:  %v\nall want: %v",
				i, gotSorted[i], wantSorted[i], gotSorted, wantSorted)
			return
		}
	}
}

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

// Phase 1 Tests: Helper Functions

func TestIsHiddenPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		// Git directories
		{
			name: "git directory",
			path: ".git/config",
			want: true,
		},
		{
			name: "git hooks",
			path: ".git/hooks/pre-commit",
			want: true,
		},
		{
			name: "nested git",
			path: "charts/.git/config",
			want: true,
		},

		// GitHub directories
		{
			name: "github workflows",
			path: ".github/workflows/test.yaml",
			want: true,
		},
		{
			name: "github actions",
			path: ".github/actions/setup/action.yaml",
			want: true,
		},

		// General hidden
		{
			name: "hidden directory",
			path: ".hidden/file",
			want: true,
		},
		{
			name: "hidden in middle",
			path: "foo/.bar/baz",
			want: true,
		},
		{
			name: "DS_Store",
			path: ".DS_Store",
			want: true,
		},
		{
			name: "hidden yaml",
			path: ".hidden-config.yaml",
			want: true,
		},

		// Not hidden
		{
			name: "normal path",
			path: "charts/app/Chart.yaml",
			want: false,
		},
		{
			name: "current directory",
			path: ".",
			want: false,
		},
		{
			name: "parent directory",
			path: "..",
			want: false,
		},
		{
			name: "relative path with ./",
			path: "./charts/app",
			want: false,
		},
		{
			name: "path starting with dot-something",
			path: "dotfiles/config",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isHiddenPath(tt.path)
			if got != tt.want {
				t.Errorf("isHiddenPath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestIsChartDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create valid chart with Chart.yaml
	validYamlDir := filepath.Join(tmpDir, "valid-yaml")
	if err := os.MkdirAll(validYamlDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(validYamlDir, "Chart.yaml"), []byte("name: test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create valid chart with Chart.yml
	validYmlDir := filepath.Join(tmpDir, "valid-yml")
	if err := os.MkdirAll(validYmlDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(validYmlDir, "Chart.yml"), []byte("name: test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create invalid directory (no Chart file)
	invalidDir := filepath.Join(tmpDir, "invalid")
	if err := os.MkdirAll(invalidDir, 0755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		dir     string
		want    bool
		wantErr bool
	}{
		{
			name: "valid chart with Chart.yaml",
			dir:  validYamlDir,
			want: true,
		},
		{
			name: "valid chart with Chart.yml",
			dir:  validYmlDir,
			want: true,
		},
		{
			name: "directory without Chart file",
			dir:  invalidDir,
			want: false,
		},
		{
			name: "non-existent directory",
			dir:  filepath.Join(tmpDir, "nonexistent"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isChartDirectory(tt.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("isChartDirectory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("isChartDirectory(%q) = %v, want %v", tt.dir, got, tt.want)
			}
		})
	}
}

// Phase 2 Tests: Chart Discovery - Pattern Variations

func TestDiscoverChartPaths_TrailingDoublestar(t *testing.T) {
	// Pattern: ./charts/** should find all charts at any depth
	tmpDir := t.TempDir()
	chartsDir := filepath.Join(tmpDir, "charts")

	// Create charts at different depths
	appDir := createTestChart(t, chartsDir, "app")
	apiDir := createTestChart(t, chartsDir, "api")
	baseCommonDir := createTestChart(t, filepath.Join(chartsDir, "base"), "common")

	pattern := filepath.Join(chartsDir, "**")
	paths, err := discoverChartPaths(pattern)
	if err != nil {
		t.Fatalf("discoverChartPaths() error = %v", err)
	}

	want := []string{appDir, apiDir, baseCommonDir}
	assertPathsEqual(t, paths, want)
}

func TestDiscoverChartPaths_ExplicitChartYaml(t *testing.T) {
	// Pattern: ./charts/**/Chart.yaml should find all Chart.yaml files
	tmpDir := t.TempDir()
	chartsDir := filepath.Join(tmpDir, "charts")

	appDir := createTestChart(t, chartsDir, "app")
	apiDir := createTestChart(t, chartsDir, "api")
	baseCommonDir := createTestChart(t, filepath.Join(chartsDir, "base"), "common")

	pattern := filepath.Join(chartsDir, "**", "Chart.yaml")
	paths, err := discoverChartPaths(pattern)
	if err != nil {
		t.Fatalf("discoverChartPaths() error = %v", err)
	}

	want := []string{appDir, apiDir, baseCommonDir}
	assertPathsEqual(t, paths, want)
}

func TestDiscoverChartPaths_ExplicitChartYml(t *testing.T) {
	// Pattern: ./charts/**/Chart.yml should find charts with .yml extension
	tmpDir := t.TempDir()
	chartsDir := filepath.Join(tmpDir, "charts")

	appDir := createTestChartWithExtension(t, chartsDir, "app", "yml")
	apiDir := createTestChartWithExtension(t, chartsDir, "api", "yml")

	pattern := filepath.Join(chartsDir, "**", "Chart.yml")
	paths, err := discoverChartPaths(pattern)
	if err != nil {
		t.Fatalf("discoverChartPaths() error = %v", err)
	}

	want := []string{appDir, apiDir}
	assertPathsEqual(t, paths, want)
}

func TestDiscoverChartPaths_SingleLevelWildcard(t *testing.T) {
	// Pattern: ./charts/* should only find charts at immediate depth
	tmpDir := t.TempDir()
	chartsDir := filepath.Join(tmpDir, "charts")

	// Create charts at immediate level
	appDir := createTestChart(t, chartsDir, "app")
	apiDir := createTestChart(t, chartsDir, "api")

	// Create chart at deeper level (should not be found)
	createTestChart(t, filepath.Join(chartsDir, "base"), "common")

	// Create non-chart directory (should be ignored)
	baseDir := filepath.Join(chartsDir, "base")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		t.Fatal(err)
	}

	pattern := filepath.Join(chartsDir, "*")
	paths, err := discoverChartPaths(pattern)
	if err != nil {
		t.Fatalf("discoverChartPaths() error = %v", err)
	}

	want := []string{appDir, apiDir}
	assertPathsEqual(t, paths, want)
}

func TestDiscoverChartPaths_BraceExpansionWithDoublestar(t *testing.T) {
	// Pattern: ./charts/{dev,prod}/** should find charts only in dev and prod
	tmpDir := t.TempDir()
	chartsDir := filepath.Join(tmpDir, "charts")

	// Create charts in dev environment
	devAppDir := createTestChart(t, filepath.Join(chartsDir, "dev"), "app")
	devApiDir := createTestChart(t, filepath.Join(chartsDir, "dev"), "api")

	// Create charts in prod environment
	prodWebDir := createTestChart(t, filepath.Join(chartsDir, "prod"), "web")

	// Create chart in staging (should not be found)
	createTestChart(t, filepath.Join(chartsDir, "staging"), "db")

	pattern := filepath.Join(chartsDir, "{dev,prod}", "**")
	paths, err := discoverChartPaths(pattern)
	if err != nil {
		t.Fatalf("discoverChartPaths() error = %v", err)
	}

	want := []string{devAppDir, devApiDir, prodWebDir}
	assertPathsEqual(t, paths, want)
}

func TestDiscoverChartPaths_NoahExample(t *testing.T) {
	// Pattern: ./pkg/** should find deeply nested chart without erroring on intermediate dirs
	// This is the bug case from Noah's example
	tmpDir := t.TempDir()
	pkgDir := filepath.Join(tmpDir, "pkg")

	// Create intermediate directories without Chart.yaml
	imageextractDir := filepath.Join(pkgDir, "imageextract")
	testdataDir := filepath.Join(imageextractDir, "testdata")
	if err := os.MkdirAll(testdataDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create chart at deep nesting
	helmChartDir := createTestChart(t, testdataDir, "helm-chart")

	pattern := filepath.Join(pkgDir, "**")
	paths, err := discoverChartPaths(pattern)
	if err != nil {
		t.Fatalf("discoverChartPaths() error = %v (should not error on intermediate dirs)", err)
	}

	want := []string{helmChartDir}
	assertPathsEqual(t, paths, want)
}

func TestDiscoverChartPaths_RootLevelDoublestar(t *testing.T) {
	// Pattern: ./** should find all charts regardless of depth in root
	tmpDir := t.TempDir()

	// Create charts at various depths
	shallowDir := createTestChart(t, tmpDir, "shallow")
	mediumDir := createTestChart(t, filepath.Join(tmpDir, "level1"), "medium")
	deepDir := createTestChart(t, filepath.Join(tmpDir, "level1", "level2", "level3"), "deep")

	pattern := filepath.Join(tmpDir, "**")
	paths, err := discoverChartPaths(pattern)
	if err != nil {
		t.Fatalf("discoverChartPaths() error = %v", err)
	}

	want := []string{shallowDir, mediumDir, deepDir}
	assertPathsEqual(t, paths, want)
}

// Phase 2 Tests: Chart Discovery - Content Scenarios

func TestDiscoverChartPaths_MixedValidInvalid(t *testing.T) {
	// Pattern should filter out invalid directories and only return valid charts
	tmpDir := t.TempDir()
	chartsDir := filepath.Join(tmpDir, "charts")

	// Create valid charts
	validDir := createTestChart(t, chartsDir, "valid-chart")
	anotherValidDir := createTestChart(t, chartsDir, "another-valid")

	// Create invalid directory (no Chart.yaml)
	invalidDir := filepath.Join(chartsDir, "invalid-dir")
	if err := os.MkdirAll(invalidDir, 0755); err != nil {
		t.Fatal(err)
	}

	pattern := filepath.Join(chartsDir, "*")
	paths, err := discoverChartPaths(pattern)
	if err != nil {
		t.Fatalf("discoverChartPaths() error = %v", err)
	}

	want := []string{validDir, anotherValidDir}
	assertPathsEqual(t, paths, want)
}

func TestDiscoverChartPaths_BothYamlAndYml(t *testing.T) {
	// Chart directory with both Chart.yaml and Chart.yml should return directory once (deduplicated)
	tmpDir := t.TempDir()
	chartsDir := filepath.Join(tmpDir, "charts")

	appDir := filepath.Join(chartsDir, "app")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create both Chart.yaml and Chart.yml in same directory
	if err := os.WriteFile(filepath.Join(appDir, "Chart.yaml"), []byte("name: app"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(appDir, "Chart.yml"), []byte("name: app"), 0644); err != nil {
		t.Fatal(err)
	}

	pattern := filepath.Join(chartsDir, "**")
	paths, err := discoverChartPaths(pattern)
	if err != nil {
		t.Fatalf("discoverChartPaths() error = %v", err)
	}

	// Should return app directory only once
	want := []string{appDir}
	assertPathsEqual(t, paths, want)
}

func TestDiscoverChartPaths_HiddenPathFiltering(t *testing.T) {
	// Pattern: ./** should filter out hidden directories like .git and .github
	tmpDir := t.TempDir()

	// Create charts in hidden directories (should be filtered)
	gitDir := filepath.Join(tmpDir, ".git", "hooks")
	createTestChart(t, gitDir, "git-chart")

	githubDir := filepath.Join(tmpDir, ".github", "workflows", "chart")
	createTestChart(t, githubDir, "github-chart")

	// Create chart in normal directory (should be found)
	chartsDir := filepath.Join(tmpDir, "charts")
	appDir := createTestChart(t, chartsDir, "app")

	pattern := filepath.Join(tmpDir, "**")
	paths, err := discoverChartPaths(pattern)
	if err != nil {
		t.Fatalf("discoverChartPaths() error = %v", err)
	}

	// Should only find the normal chart, hidden ones filtered
	want := []string{appDir}
	assertPathsEqual(t, paths, want)
}

func TestDiscoverChartPaths_EmptyResult(t *testing.T) {
	// Pattern matches directory but no Chart.yaml files exist
	tmpDir := t.TempDir()
	chartsDir := filepath.Join(tmpDir, "charts")
	if err := os.MkdirAll(chartsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create some non-chart files
	if err := os.WriteFile(filepath.Join(chartsDir, "README.md"), []byte("readme"), 0644); err != nil {
		t.Fatal(err)
	}

	pattern := filepath.Join(chartsDir, "**")
	paths, err := discoverChartPaths(pattern)
	if err != nil {
		t.Fatalf("discoverChartPaths() error = %v", err)
	}

	// Should return empty slice
	if len(paths) != 0 {
		t.Errorf("discoverChartPaths() returned %d paths, want 0 (empty result)", len(paths))
	}
}

func TestDiscoverChartPaths_IntermediateDirectoriesOnly(t *testing.T) {
	// Multiple levels of intermediate directories without Chart.yaml should not cause errors
	tmpDir := t.TempDir()
	chartsDir := filepath.Join(tmpDir, "charts")

	// Create intermediate directories without Chart.yaml
	level1 := filepath.Join(chartsDir, "level1")
	level2 := filepath.Join(level1, "level2")
	if err := os.MkdirAll(level2, 0755); err != nil {
		t.Fatal(err)
	}

	// Create chart only at deepest level
	appDir := createTestChart(t, level2, "app")

	pattern := filepath.Join(chartsDir, "**")
	paths, err := discoverChartPaths(pattern)
	if err != nil {
		t.Fatalf("discoverChartPaths() error = %v (should not error on intermediate dirs)", err)
	}

	want := []string{appDir}
	assertPathsEqual(t, paths, want)
}

// Phase 3 Tests: Preflight Discovery - isPreflightSpec Unit Tests

func TestIsPreflightSpec_ValidPreflight(t *testing.T) {
	// Valid Preflight spec should return true
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "preflight.yaml")
	createTestPreflight(t, path)

	got, err := isPreflightSpec(path)
	if err != nil {
		t.Fatalf("isPreflightSpec() error = %v", err)
	}
	if !got {
		t.Errorf("isPreflightSpec() = false, want true for valid Preflight")
	}
}

func TestIsPreflightSpec_K8sDeployment(t *testing.T) {
	// K8s Deployment should return false
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "deployment.yaml")
	createTestK8sResource(t, path, "Deployment")

	got, err := isPreflightSpec(path)
	if err != nil {
		t.Fatalf("isPreflightSpec() error = %v", err)
	}
	if got {
		t.Errorf("isPreflightSpec() = true, want false for Deployment")
	}
}

func TestIsPreflightSpec_SupportBundle(t *testing.T) {
	// SupportBundle should return false
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "bundle.yaml")
	createTestSupportBundle(t, path)

	got, err := isPreflightSpec(path)
	if err != nil {
		t.Fatalf("isPreflightSpec() error = %v", err)
	}
	if got {
		t.Errorf("isPreflightSpec() = true, want false for SupportBundle")
	}
}

func TestIsPreflightSpec_MultipleK8sResources(t *testing.T) {
	// Test multiple K8s resource types (all should return false)
	tmpDir := t.TempDir()

	kinds := []string{"ConfigMap", "Service", "Pod", "Secret"}
	for _, kind := range kinds {
		t.Run(kind, func(t *testing.T) {
			path := filepath.Join(tmpDir, kind+".yaml")
			createTestK8sResource(t, path, kind)

			got, err := isPreflightSpec(path)
			if err != nil {
				t.Fatalf("isPreflightSpec() error = %v", err)
			}
			if got {
				t.Errorf("isPreflightSpec() = true, want false for %s", kind)
			}
		})
	}
}

func TestIsPreflightSpec_MultiDocumentWithPreflight(t *testing.T) {
	// Multi-document YAML with Preflight somewhere should return true
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "multi.yaml")
	createMultiDocYAML(t, path, []string{"Deployment", "Preflight", "Service"})

	got, err := isPreflightSpec(path)
	if err != nil {
		t.Fatalf("isPreflightSpec() error = %v", err)
	}
	if !got {
		t.Errorf("isPreflightSpec() = false, want true for multi-doc with Preflight")
	}
}

func TestIsPreflightSpec_MultiDocumentWithoutPreflight(t *testing.T) {
	// Multi-document YAML without Preflight should return false
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "multi.yaml")
	createMultiDocYAML(t, path, []string{"Deployment", "Service", "ConfigMap"})

	got, err := isPreflightSpec(path)
	if err != nil {
		t.Fatalf("isPreflightSpec() error = %v", err)
	}
	if got {
		t.Errorf("isPreflightSpec() = true, want false for multi-doc without Preflight")
	}
}

func TestIsPreflightSpec_MultiDocumentMultiplePreflights(t *testing.T) {
	// Multi-document YAML with multiple Preflights should return true
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "multi.yaml")
	createMultiDocYAML(t, path, []string{"Preflight", "Preflight"})

	got, err := isPreflightSpec(path)
	if err != nil {
		t.Fatalf("isPreflightSpec() error = %v", err)
	}
	if !got {
		t.Errorf("isPreflightSpec() = false, want true for multi-doc with multiple Preflights")
	}
}

func TestIsPreflightSpec_MissingKind(t *testing.T) {
	// YAML without kind field should return false
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "nokind.yaml")
	content := "apiVersion: v1\nmetadata:\n  name: test\nspec: {}\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := isPreflightSpec(path)
	if err != nil {
		t.Fatalf("isPreflightSpec() error = %v", err)
	}
	if got {
		t.Errorf("isPreflightSpec() = true, want false for YAML without kind")
	}
}

// Phase 3 Tests: Preflight Discovery - Pattern Variations

func TestDiscoverPreflightPaths_TrailingDoublestar(t *testing.T) {
	// Pattern: ./preflights/** should find all Preflight specs at any depth
	tmpDir := t.TempDir()
	preflightsDir := filepath.Join(tmpDir, "preflights")

	// Create Preflight specs
	check1Path := filepath.Join(preflightsDir, "check1.yaml")
	createTestPreflight(t, check1Path)

	check2Path := filepath.Join(preflightsDir, "checks", "check2.yaml")
	createTestPreflight(t, check2Path)

	// Create non-Preflight files (should be filtered)
	readmePath := filepath.Join(preflightsDir, "README.md")
	if err := os.WriteFile(readmePath, []byte("readme"), 0644); err != nil {
		t.Fatal(err)
	}

	pattern := filepath.Join(preflightsDir, "**")
	paths, err := discoverPreflightPaths(pattern)
	if err != nil {
		t.Fatalf("discoverPreflightPaths() error = %v", err)
	}

	want := []string{check1Path, check2Path}
	assertPathsEqual(t, paths, want)
}

func TestDiscoverPreflightPaths_ExplicitYaml(t *testing.T) {
	// Pattern: ./preflights/**/*.yaml should find all .yaml Preflight specs
	tmpDir := t.TempDir()
	preflightsDir := filepath.Join(tmpDir, "preflights")

	check1Path := filepath.Join(preflightsDir, "check1.yaml")
	createTestPreflight(t, check1Path)

	check2Path := filepath.Join(preflightsDir, "checks", "check2.yaml")
	createTestPreflight(t, check2Path)

	pattern := filepath.Join(preflightsDir, "**", "*.yaml")
	paths, err := discoverPreflightPaths(pattern)
	if err != nil {
		t.Fatalf("discoverPreflightPaths() error = %v", err)
	}

	want := []string{check1Path, check2Path}
	assertPathsEqual(t, paths, want)
}

func TestDiscoverPreflightPaths_ExplicitYml(t *testing.T) {
	// Pattern: ./preflights/**/*.yml should find .yml Preflight specs
	tmpDir := t.TempDir()
	preflightsDir := filepath.Join(tmpDir, "preflights")

	// Create Preflight with .yml extension
	checkPath := filepath.Join(preflightsDir, "check.yml")
	dir := filepath.Dir(checkPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	content := `apiVersion: troubleshoot.sh/v1beta2
kind: Preflight
metadata:
  name: test
spec:
  collectors: []
`
	if err := os.WriteFile(checkPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	pattern := filepath.Join(preflightsDir, "**", "*.yml")
	paths, err := discoverPreflightPaths(pattern)
	if err != nil {
		t.Fatalf("discoverPreflightPaths() error = %v", err)
	}

	want := []string{checkPath}
	assertPathsEqual(t, paths, want)
}

func TestDiscoverPreflightPaths_SingleLevel(t *testing.T) {
	// Pattern: ./preflights/* should only find Preflights at immediate depth
	tmpDir := t.TempDir()
	preflightsDir := filepath.Join(tmpDir, "preflights")

	// Create Preflight at immediate level
	checkPath := filepath.Join(preflightsDir, "check.yaml")
	createTestPreflight(t, checkPath)

	// Create Preflight in subdirectory (should not be found)
	createTestPreflight(t, filepath.Join(preflightsDir, "subdir", "check2.yaml"))

	pattern := filepath.Join(preflightsDir, "*")
	paths, err := discoverPreflightPaths(pattern)
	if err != nil {
		t.Fatalf("discoverPreflightPaths() error = %v", err)
	}

	want := []string{checkPath}
	assertPathsEqual(t, paths, want)
}

func TestDiscoverPreflightPaths_BraceExpansion(t *testing.T) {
	// Pattern: ./preflights/{dev,prod}/** should only find Preflights in dev and prod
	tmpDir := t.TempDir()
	preflightsDir := filepath.Join(tmpDir, "preflights")

	// Create Preflights in dev and prod
	devPath := filepath.Join(preflightsDir, "dev", "check.yaml")
	createTestPreflight(t, devPath)

	prodPath := filepath.Join(preflightsDir, "prod", "check.yaml")
	createTestPreflight(t, prodPath)

	// Create Preflight in staging (should not be found)
	createTestPreflight(t, filepath.Join(preflightsDir, "staging", "check.yaml"))

	pattern := filepath.Join(preflightsDir, "{dev,prod}", "**")
	paths, err := discoverPreflightPaths(pattern)
	if err != nil {
		t.Fatalf("discoverPreflightPaths() error = %v", err)
	}

	want := []string{devPath, prodPath}
	assertPathsEqual(t, paths, want)
}

func TestDiscoverPreflightPaths_MixedDirectory(t *testing.T) {
	// Pattern: ./k8s/** should only find Preflights, filtering out other K8s resources
	tmpDir := t.TempDir()
	k8sDir := filepath.Join(tmpDir, "k8s")

	// Create Preflight
	preflightPath := filepath.Join(k8sDir, "preflight.yaml")
	createTestPreflight(t, preflightPath)

	// Create other K8s resources (should be filtered)
	createTestK8sResource(t, filepath.Join(k8sDir, "deployment.yaml"), "Deployment")
	createTestK8sResource(t, filepath.Join(k8sDir, "service.yaml"), "Service")

	pattern := filepath.Join(k8sDir, "**")
	paths, err := discoverPreflightPaths(pattern)
	if err != nil {
		t.Fatalf("discoverPreflightPaths() error = %v", err)
	}

	want := []string{preflightPath}
	assertPathsEqual(t, paths, want)
}

// Phase 3 Tests: Preflight Discovery - Content Filtering

func TestDiscoverPreflightPaths_NonYamlFilesFiltered(t *testing.T) {
	// Pattern should filter out non-YAML files
	tmpDir := t.TempDir()
	preflightsDir := filepath.Join(tmpDir, "preflights")

	// Create Preflight
	checkPath := filepath.Join(preflightsDir, "check.yaml")
	createTestPreflight(t, checkPath)

	// Create non-YAML files (should be ignored)
	if err := os.WriteFile(filepath.Join(preflightsDir, "README.md"), []byte("readme"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(preflightsDir, "config.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(preflightsDir, "notes.txt"), []byte("notes"), 0644); err != nil {
		t.Fatal(err)
	}

	pattern := filepath.Join(preflightsDir, "**")
	paths, err := discoverPreflightPaths(pattern)
	if err != nil {
		t.Fatalf("discoverPreflightPaths() error = %v", err)
	}

	want := []string{checkPath}
	assertPathsEqual(t, paths, want)
}

func TestDiscoverPreflightPaths_EmptyYaml(t *testing.T) {
	// Empty YAML file should be filtered out
	tmpDir := t.TempDir()
	preflightsDir := filepath.Join(tmpDir, "preflights")
	if err := os.MkdirAll(preflightsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create empty YAML
	emptyPath := filepath.Join(preflightsDir, "empty.yaml")
	if err := os.WriteFile(emptyPath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	pattern := filepath.Join(preflightsDir, "**")
	paths, err := discoverPreflightPaths(pattern)
	if err != nil {
		t.Fatalf("discoverPreflightPaths() error = %v", err)
	}

	// Should be empty (empty file filtered)
	if len(paths) != 0 {
		t.Errorf("discoverPreflightPaths() returned %d paths, want 0 (empty file should be filtered)", len(paths))
	}
}

func TestDiscoverPreflightPaths_InvalidYaml(t *testing.T) {
	// Invalid YAML should be filtered gracefully, not error
	tmpDir := t.TempDir()
	preflightsDir := filepath.Join(tmpDir, "preflights")

	// Create valid Preflight
	validPath := filepath.Join(preflightsDir, "valid.yaml")
	createTestPreflight(t, validPath)

	// Create invalid YAML (malformed syntax)
	brokenPath := filepath.Join(preflightsDir, "broken.yaml")
	if err := os.WriteFile(brokenPath, []byte("this is: not: valid: yaml:"), 0644); err != nil {
		t.Fatal(err)
	}

	pattern := filepath.Join(preflightsDir, "**")
	paths, err := discoverPreflightPaths(pattern)
	if err != nil {
		t.Fatalf("discoverPreflightPaths() error = %v", err)
	}

	// Should only find valid Preflight, broken one filtered
	want := []string{validPath}
	assertPathsEqual(t, paths, want)
}

func TestDiscoverPreflightPaths_BothExtensions(t *testing.T) {
	// Both .yaml and .yml files should be found if they're valid Preflights
	tmpDir := t.TempDir()
	preflightsDir := filepath.Join(tmpDir, "preflights")

	// Create Preflight with .yaml
	yamlPath := filepath.Join(preflightsDir, "check.yaml")
	createTestPreflight(t, yamlPath)

	// Create Preflight with .yml
	ymlPath := filepath.Join(preflightsDir, "check.yml")
	dir := filepath.Dir(ymlPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	content := `apiVersion: troubleshoot.sh/v1beta2
kind: Preflight
metadata:
  name: test-yml
spec:
  collectors: []
`
	if err := os.WriteFile(ymlPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	pattern := filepath.Join(preflightsDir, "**")
	paths, err := discoverPreflightPaths(pattern)
	if err != nil {
		t.Fatalf("discoverPreflightPaths() error = %v", err)
	}

	want := []string{yamlPath, ymlPath}
	assertPathsEqual(t, paths, want)
}

// Phase 4 Tests: Support Bundle Discovery - Additional Pattern Tests

func TestDiscoverSupportBundlesFromManifests_TrailingDoublestarPattern(t *testing.T) {
	// Pattern: ./manifests/** should find all SupportBundle specs at any depth
	tmpDir := t.TempDir()
	manifestsDir := filepath.Join(tmpDir, "manifests")

	// Create SupportBundle specs at different depths
	bundle1Path := filepath.Join(manifestsDir, "bundle1.yaml")
	createTestSupportBundle(t, bundle1Path)

	bundle2Path := filepath.Join(manifestsDir, "sub", "bundle2.yaml")
	createTestSupportBundle(t, bundle2Path)

	// Create non-SupportBundle resources (should be filtered)
	createTestK8sResource(t, filepath.Join(manifestsDir, "deployment.yaml"), "Deployment")

	pattern := filepath.Join(manifestsDir, "**", "*.yaml")
	paths, err := DiscoverSupportBundlesFromManifests([]string{pattern})
	if err != nil {
		t.Fatalf("DiscoverSupportBundlesFromManifests() error = %v", err)
	}

	want := []string{bundle1Path, bundle2Path}
	assertPathsEqual(t, paths, want)
}

func TestDiscoverSupportBundlesFromManifests_MixedK8sResourcesFiltering(t *testing.T) {
	// Pattern should find only SupportBundles, filtering out other K8s resources
	tmpDir := t.TempDir()
	k8sDir := filepath.Join(tmpDir, "k8s")

	// Create SupportBundle
	bundlePath := filepath.Join(k8sDir, "bundle.yaml")
	createTestSupportBundle(t, bundlePath)

	// Create various K8s resources (all should be filtered)
	createTestK8sResource(t, filepath.Join(k8sDir, "deployment.yaml"), "Deployment")
	createTestK8sResource(t, filepath.Join(k8sDir, "service.yaml"), "Service")
	createTestK8sResource(t, filepath.Join(k8sDir, "configmap.yaml"), "ConfigMap")
	createTestK8sResource(t, filepath.Join(k8sDir, "secret.yaml"), "Secret")

	pattern := filepath.Join(k8sDir, "**", "*.yaml")
	paths, err := DiscoverSupportBundlesFromManifests([]string{pattern})
	if err != nil {
		t.Fatalf("DiscoverSupportBundlesFromManifests() error = %v", err)
	}

	want := []string{bundlePath}
	assertPathsEqual(t, paths, want)
}

func TestDiscoverSupportBundlesFromManifests_BraceExpansionPattern(t *testing.T) {
	// Pattern with brace expansion should find bundles only in specified directories
	tmpDir := t.TempDir()
	k8sDir := filepath.Join(tmpDir, "k8s")

	// Create SupportBundles in dev and prod
	devPath := filepath.Join(k8sDir, "dev", "bundle.yaml")
	createTestSupportBundle(t, devPath)

	prodPath := filepath.Join(k8sDir, "prod", "bundle.yaml")
	createTestSupportBundle(t, prodPath)

	// Create SupportBundle in staging (should not be found)
	createTestSupportBundle(t, filepath.Join(k8sDir, "staging", "bundle.yaml"))

	pattern := filepath.Join(k8sDir, "{dev,prod}", "**", "*.yaml")
	paths, err := DiscoverSupportBundlesFromManifests([]string{pattern})
	if err != nil {
		t.Fatalf("DiscoverSupportBundlesFromManifests() error = %v", err)
	}

	want := []string{devPath, prodPath}
	assertPathsEqual(t, paths, want)
}

func TestDiscoverSupportBundlesFromManifests_MultiplePatternsNonOverlapping(t *testing.T) {
	// Multiple patterns should find bundles from all patterns
	tmpDir := t.TempDir()

	// Create SupportBundles in different directories
	devPath := filepath.Join(tmpDir, "dev", "bundle.yaml")
	createTestSupportBundle(t, devPath)

	prodPath := filepath.Join(tmpDir, "prod", "bundle.yaml")
	createTestSupportBundle(t, prodPath)

	patterns := []string{
		filepath.Join(tmpDir, "dev", "**", "*.yaml"),
		filepath.Join(tmpDir, "prod", "**", "*.yaml"),
	}
	paths, err := DiscoverSupportBundlesFromManifests(patterns)
	if err != nil {
		t.Fatalf("DiscoverSupportBundlesFromManifests() error = %v", err)
	}

	want := []string{devPath, prodPath}
	assertPathsEqual(t, paths, want)
}

func TestDiscoverSupportBundlesFromManifests_OverlappingPatternsDeduplication(t *testing.T) {
	// Overlapping patterns should return each bundle only once (deduplication)
	tmpDir := t.TempDir()
	manifestsDir := filepath.Join(tmpDir, "manifests")

	// Create SupportBundle in prod subdirectory
	bundlePath := filepath.Join(manifestsDir, "prod", "bundle.yaml")
	createTestSupportBundle(t, bundlePath)

	// Use overlapping patterns that both match the same file
	patterns := []string{
		filepath.Join(manifestsDir, "**", "*.yaml"),     // Matches all
		filepath.Join(manifestsDir, "prod", "*.yaml"),   // Matches prod only
	}
	paths, err := DiscoverSupportBundlesFromManifests(patterns)
	if err != nil {
		t.Fatalf("DiscoverSupportBundlesFromManifests() error = %v", err)
	}

	// Should return bundle only once despite overlapping patterns
	want := []string{bundlePath}
	assertPathsEqual(t, paths, want)
}

// Phase 4 Tests: Support Bundle Discovery - Additional Content Filtering Tests

func TestIsSupportBundleSpec_VariousK8sResources(t *testing.T) {
	// Test that various K8s resource types return false
	tmpDir := t.TempDir()

	kinds := []string{
		"Deployment",
		"Service",
		"Pod",
		"ConfigMap",
		"Secret",
		"StatefulSet",
		"DaemonSet",
		"Job",
	}

	for _, kind := range kinds {
		t.Run(kind, func(t *testing.T) {
			path := filepath.Join(tmpDir, kind+".yaml")
			createTestK8sResource(t, path, kind)

			got, err := isSupportBundleSpec(path)
			if err != nil {
				t.Fatalf("isSupportBundleSpec() error = %v", err)
			}
			if got {
				t.Errorf("isSupportBundleSpec() = true, want false for %s", kind)
			}
		})
	}
}

func TestIsSupportBundleSpec_PreflightResource(t *testing.T) {
	// Preflight spec should return false
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "preflight.yaml")
	createTestPreflight(t, path)

	got, err := isSupportBundleSpec(path)
	if err != nil {
		t.Fatalf("isSupportBundleSpec() error = %v", err)
	}
	if got {
		t.Errorf("isSupportBundleSpec() = true, want false for Preflight")
	}
}

func TestIsSupportBundleSpec_MultiDocumentWithBundle(t *testing.T) {
	// Multi-document YAML with SupportBundle should return true
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "multi.yaml")
	createMultiDocYAML(t, path, []string{"Deployment", "SupportBundle", "Service"})

	got, err := isSupportBundleSpec(path)
	if err != nil {
		t.Fatalf("isSupportBundleSpec() error = %v", err)
	}
	if !got {
		t.Errorf("isSupportBundleSpec() = false, want true for multi-doc with SupportBundle")
	}
}

func TestIsSupportBundleSpec_MultiDocumentMultipleBundles(t *testing.T) {
	// Multi-document YAML with multiple SupportBundles should return true
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "multi.yaml")
	createMultiDocYAML(t, path, []string{"SupportBundle", "SupportBundle"})

	got, err := isSupportBundleSpec(path)
	if err != nil {
		t.Fatalf("isSupportBundleSpec() error = %v", err)
	}
	if !got {
		t.Errorf("isSupportBundleSpec() = false, want true for multi-doc with multiple SupportBundles")
	}
}

func TestDiscoverSupportBundlesFromManifests_NonYamlFilesIgnored(t *testing.T) {
	// Non-YAML files should be ignored during discovery
	tmpDir := t.TempDir()
	manifestsDir := filepath.Join(tmpDir, "manifests")

	// Create SupportBundle
	bundlePath := filepath.Join(manifestsDir, "bundle.yaml")
	createTestSupportBundle(t, bundlePath)

	// Create non-YAML files (should be ignored)
	if err := os.WriteFile(filepath.Join(manifestsDir, "README.md"), []byte("readme"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(manifestsDir, "config.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	pattern := filepath.Join(manifestsDir, "**", "*.yaml")
	paths, err := DiscoverSupportBundlesFromManifests([]string{pattern})
	if err != nil {
		t.Fatalf("DiscoverSupportBundlesFromManifests() error = %v", err)
	}

	want := []string{bundlePath}
	assertPathsEqual(t, paths, want)
}
