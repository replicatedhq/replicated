package lint2

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverHelmChartManifests(t *testing.T) {
	t.Run("empty manifests list returns error", func(t *testing.T) {
		_, err := DiscoverHelmChartManifests([]string{})
		if err == nil {
			t.Fatal("expected error for empty manifests list, got nil")
		}
		if err.Error() != "no manifests configured - cannot discover HelmChart resources (required for templated preflights)" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("single valid HelmChart with builder values", func(t *testing.T) {
		tmpDir := t.TempDir()
		helmChartFile := filepath.Join(tmpDir, "helmchart.yaml")
		content := `apiVersion: kots.io/v1beta1
kind: HelmChart
metadata:
  name: my-app-chart
spec:
  chart:
    name: my-app
    chartVersion: 1.2.3
  builder:
    postgresql:
      enabled: true
    redis:
      enabled: true
`
		if err := os.WriteFile(helmChartFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		pattern := filepath.Join(tmpDir, "*.yaml")
		manifests, err := DiscoverHelmChartManifests([]string{pattern})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(manifests) != 1 {
			t.Fatalf("expected 1 manifest, got %d", len(manifests))
		}

		key := "my-app:1.2.3"
		manifest, found := manifests[key]
		if !found {
			t.Fatalf("expected manifest with key %q not found", key)
		}

		if manifest.Name != "my-app" {
			t.Errorf("expected name 'my-app', got %q", manifest.Name)
		}
		if manifest.ChartVersion != "1.2.3" {
			t.Errorf("expected chartVersion '1.2.3', got %q", manifest.ChartVersion)
		}
		if manifest.FilePath != helmChartFile {
			t.Errorf("expected filePath %q, got %q", helmChartFile, manifest.FilePath)
		}

		if manifest.BuilderValues == nil {
			t.Fatal("expected builder values, got nil")
		}
		postgresql, ok := manifest.BuilderValues["postgresql"].(map[string]interface{})
		if !ok {
			t.Fatal("expected postgresql in builder values")
		}
		if postgresql["enabled"] != true {
			t.Errorf("expected postgresql.enabled=true, got %v", postgresql["enabled"])
		}
	})

	t.Run("multiple unique HelmCharts", func(t *testing.T) {
		tmpDir := t.TempDir()

		// First chart
		helmChart1 := filepath.Join(tmpDir, "chart1.yaml")
		content1 := `apiVersion: kots.io/v1beta1
kind: HelmChart
metadata:
  name: chart1
spec:
  chart:
    name: app-one
    chartVersion: 1.0.0
  builder:
    enabled: true
`
		if err := os.WriteFile(helmChart1, []byte(content1), 0644); err != nil {
			t.Fatal(err)
		}

		// Second chart
		helmChart2 := filepath.Join(tmpDir, "chart2.yaml")
		content2 := `apiVersion: kots.io/v1beta2
kind: HelmChart
metadata:
  name: chart2
spec:
  chart:
    name: app-two
    chartVersion: 2.0.0
  builder:
    features:
      analytics: true
`
		if err := os.WriteFile(helmChart2, []byte(content2), 0644); err != nil {
			t.Fatal(err)
		}

		pattern := filepath.Join(tmpDir, "*.yaml")
		manifests, err := DiscoverHelmChartManifests([]string{pattern})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(manifests) != 2 {
			t.Fatalf("expected 2 manifests, got %d", len(manifests))
		}

		// Check first chart
		manifest1, found := manifests["app-one:1.0.0"]
		if !found {
			t.Fatal("expected manifest 'app-one:1.0.0' not found")
		}
		if manifest1.Name != "app-one" {
			t.Errorf("expected name 'app-one', got %q", manifest1.Name)
		}

		// Check second chart
		manifest2, found := manifests["app-two:2.0.0"]
		if !found {
			t.Fatal("expected manifest 'app-two:2.0.0' not found")
		}
		if manifest2.Name != "app-two" {
			t.Errorf("expected name 'app-two', got %q", manifest2.Name)
		}
	})

	t.Run("duplicate HelmChart returns error with both paths", func(t *testing.T) {
		tmpDir := t.TempDir()

		// First chart
		helmChart1 := filepath.Join(tmpDir, "helmchart-dev.yaml")
		content := `apiVersion: kots.io/v1beta1
kind: HelmChart
metadata:
  name: dev-chart
spec:
  chart:
    name: my-app
    chartVersion: 1.2.3
  builder:
    env: dev
`
		if err := os.WriteFile(helmChart1, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		// Duplicate chart (same name:version)
		helmChart2 := filepath.Join(tmpDir, "helmchart-prod.yaml")
		content2 := `apiVersion: kots.io/v1beta1
kind: HelmChart
metadata:
  name: prod-chart
spec:
  chart:
    name: my-app
    chartVersion: 1.2.3
  builder:
    env: prod
`
		if err := os.WriteFile(helmChart2, []byte(content2), 0644); err != nil {
			t.Fatal(err)
		}

		pattern := filepath.Join(tmpDir, "*.yaml")
		_, err := DiscoverHelmChartManifests([]string{pattern})

		if err == nil {
			t.Fatal("expected error for duplicate HelmChart, got nil")
		}

		dupErr, ok := err.(*DuplicateHelmChartError)
		if !ok {
			t.Fatalf("expected DuplicateHelmChartError, got %T: %v", err, err)
		}

		if dupErr.ChartKey != "my-app:1.2.3" {
			t.Errorf("expected ChartKey 'my-app:1.2.3', got %q", dupErr.ChartKey)
		}

		// Check that both file paths are in the error
		errMsg := dupErr.Error()
		if errMsg == "" {
			t.Error("error message is empty")
		}
		// Error should mention both files (order may vary depending on filesystem)
		hasDevFile := filepath.Base(dupErr.FirstFile) == "helmchart-dev.yaml" ||
			filepath.Base(dupErr.SecondFile) == "helmchart-dev.yaml"
		hasProdFile := filepath.Base(dupErr.FirstFile) == "helmchart-prod.yaml" ||
			filepath.Base(dupErr.SecondFile) == "helmchart-prod.yaml"

		if !hasDevFile || !hasProdFile {
			t.Errorf("error should reference both files, got: %v", errMsg)
		}
	})

	t.Run("empty builder section is valid", func(t *testing.T) {
		tmpDir := t.TempDir()
		helmChartFile := filepath.Join(tmpDir, "helmchart.yaml")
		content := `apiVersion: kots.io/v1beta1
kind: HelmChart
metadata:
  name: my-chart
spec:
  chart:
    name: my-app
    chartVersion: 1.0.0
  builder: {}
`
		if err := os.WriteFile(helmChartFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		pattern := filepath.Join(tmpDir, "*.yaml")
		manifests, err := DiscoverHelmChartManifests([]string{pattern})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(manifests) != 1 {
			t.Fatalf("expected 1 manifest, got %d", len(manifests))
		}

		manifest := manifests["my-app:1.0.0"]
		if manifest.BuilderValues == nil || len(manifest.BuilderValues) != 0 {
			t.Errorf("expected empty builder values map, got %v", manifest.BuilderValues)
		}
	})

	t.Run("missing builder section is valid", func(t *testing.T) {
		tmpDir := t.TempDir()
		helmChartFile := filepath.Join(tmpDir, "helmchart.yaml")
		content := `apiVersion: kots.io/v1beta1
kind: HelmChart
metadata:
  name: my-chart
spec:
  chart:
    name: my-app
    chartVersion: 1.0.0
`
		if err := os.WriteFile(helmChartFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		pattern := filepath.Join(tmpDir, "*.yaml")
		manifests, err := DiscoverHelmChartManifests([]string{pattern})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(manifests) != 1 {
			t.Fatalf("expected 1 manifest, got %d", len(manifests))
		}

		manifest := manifests["my-app:1.0.0"]
		// Builder values can be nil or empty map when not specified - both are valid
		if manifest.BuilderValues != nil && len(manifest.BuilderValues) != 0 {
			t.Errorf("expected empty/nil builder values, got %v", manifest.BuilderValues)
		}
	})

	t.Run("missing required fields skipped", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Missing name
		helmChart1 := filepath.Join(tmpDir, "missing-name.yaml")
		content1 := `apiVersion: kots.io/v1beta1
kind: HelmChart
metadata:
  name: test
spec:
  chart:
    chartVersion: 1.0.0
  builder: {}
`
		if err := os.WriteFile(helmChart1, []byte(content1), 0644); err != nil {
			t.Fatal(err)
		}

		// Missing chartVersion
		helmChart2 := filepath.Join(tmpDir, "missing-version.yaml")
		content2 := `apiVersion: kots.io/v1beta1
kind: HelmChart
metadata:
  name: test
spec:
  chart:
    name: my-app
  builder: {}
`
		if err := os.WriteFile(helmChart2, []byte(content2), 0644); err != nil {
			t.Fatal(err)
		}

		pattern := filepath.Join(tmpDir, "*.yaml")
		manifests, err := DiscoverHelmChartManifests([]string{pattern})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(manifests) != 0 {
			t.Fatalf("expected 0 manifests (invalid files skipped), got %d", len(manifests))
		}
	})

	t.Run("invalid YAML skipped gracefully", func(t *testing.T) {
		tmpDir := t.TempDir()
		invalidFile := filepath.Join(tmpDir, "invalid.yaml")
		content := `apiVersion: kots.io/v1beta1
kind: HelmChart
metadata:
  name: [invalid yaml here
spec:
  chart:
`
		if err := os.WriteFile(invalidFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		pattern := filepath.Join(tmpDir, "*.yaml")
		manifests, err := DiscoverHelmChartManifests([]string{pattern})
		if err != nil {
			t.Fatalf("unexpected error (should skip invalid YAML): %v", err)
		}

		if len(manifests) != 0 {
			t.Fatalf("expected 0 manifests (invalid YAML skipped), got %d", len(manifests))
		}
	})

	t.Run("multi-document YAML with HelmChart", func(t *testing.T) {
		tmpDir := t.TempDir()
		multiDocFile := filepath.Join(tmpDir, "multi.yaml")
		content := `apiVersion: v1
kind: ConfigMap
metadata:
  name: config
data:
  foo: bar
---
apiVersion: kots.io/v1beta1
kind: HelmChart
metadata:
  name: my-chart
spec:
  chart:
    name: my-app
    chartVersion: 1.0.0
  builder:
    enabled: true
---
apiVersion: v1
kind: Service
metadata:
  name: svc
`
		if err := os.WriteFile(multiDocFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		pattern := filepath.Join(tmpDir, "*.yaml")
		manifests, err := DiscoverHelmChartManifests([]string{pattern})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(manifests) != 1 {
			t.Fatalf("expected 1 manifest from multi-doc YAML, got %d", len(manifests))
		}

		manifest := manifests["my-app:1.0.0"]
		if manifest == nil {
			t.Fatal("expected manifest 'my-app:1.0.0' not found")
		}
	})

	t.Run("non-HelmChart files skipped", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a mix of files
		configMap := filepath.Join(tmpDir, "configmap.yaml")
		cm := `apiVersion: v1
kind: ConfigMap
metadata:
  name: config
`
		if err := os.WriteFile(configMap, []byte(cm), 0644); err != nil {
			t.Fatal(err)
		}

		deployment := filepath.Join(tmpDir, "deployment.yaml")
		dep := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
`
		if err := os.WriteFile(deployment, []byte(dep), 0644); err != nil {
			t.Fatal(err)
		}

		helmChart := filepath.Join(tmpDir, "helmchart.yaml")
		hc := `apiVersion: kots.io/v1beta1
kind: HelmChart
metadata:
  name: test
spec:
  chart:
    name: my-app
    chartVersion: 1.0.0
`
		if err := os.WriteFile(helmChart, []byte(hc), 0644); err != nil {
			t.Fatal(err)
		}

		pattern := filepath.Join(tmpDir, "*.yaml")
		manifests, err := DiscoverHelmChartManifests([]string{pattern})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(manifests) != 1 {
			t.Fatalf("expected 1 HelmChart (others skipped), got %d", len(manifests))
		}

		if _, found := manifests["my-app:1.0.0"]; !found {
			t.Fatal("expected manifest 'my-app:1.0.0' not found")
		}
	})

	t.Run("glob pattern expansion", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create nested structure
		devDir := filepath.Join(tmpDir, "dev")
		prodDir := filepath.Join(tmpDir, "prod")
		if err := os.MkdirAll(devDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(prodDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Dev chart
		devChart := filepath.Join(devDir, "helmchart.yaml")
		devContent := `apiVersion: kots.io/v1beta1
kind: HelmChart
metadata:
  name: dev-chart
spec:
  chart:
    name: app
    chartVersion: 1.0.0-dev
  builder:
    env: dev
`
		if err := os.WriteFile(devChart, []byte(devContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Prod chart
		prodChart := filepath.Join(prodDir, "helmchart.yaml")
		prodContent := `apiVersion: kots.io/v1beta1
kind: HelmChart
metadata:
  name: prod-chart
spec:
  chart:
    name: app
    chartVersion: 1.0.0-prod
  builder:
    env: prod
`
		if err := os.WriteFile(prodChart, []byte(prodContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Use recursive glob pattern
		pattern := filepath.Join(tmpDir, "**", "*.yaml")
		manifests, err := DiscoverHelmChartManifests([]string{pattern})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(manifests) != 2 {
			t.Fatalf("expected 2 manifests from recursive glob, got %d", len(manifests))
		}

		if _, found := manifests["app:1.0.0-dev"]; !found {
			t.Error("expected dev manifest not found")
		}
		if _, found := manifests["app:1.0.0-prod"]; !found {
			t.Error("expected prod manifest not found")
		}
	})

	t.Run("hidden directories skipped", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create .git directory with HelmChart
		gitDir := filepath.Join(tmpDir, ".git")
		if err := os.MkdirAll(gitDir, 0755); err != nil {
			t.Fatal(err)
		}

		gitChart := filepath.Join(gitDir, "helmchart.yaml")
		gitContent := `apiVersion: kots.io/v1beta1
kind: HelmChart
metadata:
  name: git-chart
spec:
  chart:
    name: should-be-ignored
    chartVersion: 1.0.0
`
		if err := os.WriteFile(gitChart, []byte(gitContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Create normal chart
		normalChart := filepath.Join(tmpDir, "helmchart.yaml")
		normalContent := `apiVersion: kots.io/v1beta1
kind: HelmChart
metadata:
  name: normal-chart
spec:
  chart:
    name: app
    chartVersion: 1.0.0
`
		if err := os.WriteFile(normalChart, []byte(normalContent), 0644); err != nil {
			t.Fatal(err)
		}

		pattern := filepath.Join(tmpDir, "**", "*.yaml")
		manifests, err := DiscoverHelmChartManifests([]string{pattern})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(manifests) != 1 {
			t.Fatalf("expected 1 manifest (hidden dir skipped), got %d", len(manifests))
		}

		if _, found := manifests["should-be-ignored:1.0.0"]; found {
			t.Error("chart from .git directory should be ignored")
		}
		if _, found := manifests["app:1.0.0"]; !found {
			t.Error("normal chart should be found")
		}
	})

	t.Run("both v1beta1 and v1beta2 supported", func(t *testing.T) {
		tmpDir := t.TempDir()

		// v1beta1
		v1Chart := filepath.Join(tmpDir, "v1.yaml")
		v1Content := `apiVersion: kots.io/v1beta1
kind: HelmChart
metadata:
  name: v1-chart
spec:
  chart:
    name: app-v1
    chartVersion: 1.0.0
    releaseName: old-style
  builder:
    version: v1
`
		if err := os.WriteFile(v1Chart, []byte(v1Content), 0644); err != nil {
			t.Fatal(err)
		}

		// v1beta2
		v2Chart := filepath.Join(tmpDir, "v2.yaml")
		v2Content := `apiVersion: kots.io/v1beta2
kind: HelmChart
metadata:
  name: v2-chart
spec:
  chart:
    name: app-v2
    chartVersion: 2.0.0
  releaseName: new-style
  builder:
    version: v2
`
		if err := os.WriteFile(v2Chart, []byte(v2Content), 0644); err != nil {
			t.Fatal(err)
		}

		pattern := filepath.Join(tmpDir, "*.yaml")
		manifests, err := DiscoverHelmChartManifests([]string{pattern})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(manifests) != 2 {
			t.Fatalf("expected 2 manifests (v1 and v2), got %d", len(manifests))
		}

		v1Manifest, found := manifests["app-v1:1.0.0"]
		if !found {
			t.Fatal("v1beta1 chart not found")
		}
		if v1Manifest.BuilderValues["version"] != "v1" {
			t.Errorf("expected v1 builder values, got %v", v1Manifest.BuilderValues)
		}

		v2Manifest, found := manifests["app-v2:2.0.0"]
		if !found {
			t.Fatal("v1beta2 chart not found")
		}
		if v2Manifest.BuilderValues["version"] != "v2" {
			t.Errorf("expected v2 builder values, got %v", v2Manifest.BuilderValues)
		}
	})

	t.Run("future apiVersion accepted", func(t *testing.T) {
		tmpDir := t.TempDir()
		helmChartFile := filepath.Join(tmpDir, "v3.yaml")
		content := `apiVersion: kots.io/v1beta3
kind: HelmChart
metadata:
  name: future-chart
spec:
  chart:
    name: my-app
    chartVersion: 2.0.0
  builder:
    future: true
`
		if err := os.WriteFile(helmChartFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		pattern := filepath.Join(tmpDir, "*.yaml")
		manifests, err := DiscoverHelmChartManifests([]string{pattern})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Discovery should accept any apiVersion - validation happens in linter
		if len(manifests) != 1 {
			t.Fatalf("expected 1 manifest (future apiVersion accepted), got %d", len(manifests))
		}

		manifest := manifests["my-app:2.0.0"]
		if manifest == nil {
			t.Fatal("expected future apiVersion to be discovered")
		}
		if manifest.BuilderValues["future"] != true {
			t.Errorf("expected future=true in builder values, got %v", manifest.BuilderValues["future"])
		}
	})

	t.Run("complex nested builder values", func(t *testing.T) {
		tmpDir := t.TempDir()
		helmChartFile := filepath.Join(tmpDir, "complex.yaml")
		content := `apiVersion: kots.io/v1beta1
kind: HelmChart
metadata:
  name: complex-chart
spec:
  chart:
    name: my-app
    chartVersion: 1.0.0
  builder:
    postgresql:
      enabled: true
      resources:
        requests:
          memory: "256Mi"
          cpu: "100m"
        limits:
          memory: "512Mi"
          cpu: "500m"
    redis:
      enabled: true
      cluster:
        nodes: 3
    features:
      - analytics
      - logging
      - monitoring
`
		if err := os.WriteFile(helmChartFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		pattern := filepath.Join(tmpDir, "*.yaml")
		manifests, err := DiscoverHelmChartManifests([]string{pattern})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		manifest := manifests["my-app:1.0.0"]
		if manifest == nil {
			t.Fatal("manifest not found")
		}

		// Verify nested structure is preserved
		postgresql, ok := manifest.BuilderValues["postgresql"].(map[string]interface{})
		if !ok {
			t.Fatal("postgresql not found in builder values")
		}

		resources, ok := postgresql["resources"].(map[string]interface{})
		if !ok {
			t.Fatal("resources not found in postgresql")
		}

		requests, ok := resources["requests"].(map[string]interface{})
		if !ok {
			t.Fatal("requests not found in resources")
		}

		if requests["memory"] != "256Mi" {
			t.Errorf("expected memory=256Mi, got %v", requests["memory"])
		}

		// Verify arrays are preserved
		features, ok := manifest.BuilderValues["features"].([]interface{})
		if !ok {
			t.Fatal("features not found or not an array")
		}
		if len(features) != 3 {
			t.Errorf("expected 3 features, got %d", len(features))
		}
	})
}
