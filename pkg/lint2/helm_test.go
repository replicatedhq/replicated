package lint2

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestGetChartMetadata(t *testing.T) {
	t.Run("valid chart with name and version", func(t *testing.T) {
		tmpDir := t.TempDir()
		chartYaml := `apiVersion: v2
name: my-chart
version: 1.2.3
description: A test chart
`
		chartPath := filepath.Join(tmpDir, "Chart.yaml")
		if err := os.WriteFile(chartPath, []byte(chartYaml), 0644); err != nil {
			t.Fatal(err)
		}

		metadata, err := GetChartMetadata(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if metadata.Name != "my-chart" {
			t.Errorf("expected name 'my-chart', got %q", metadata.Name)
		}
		if metadata.Version != "1.2.3" {
			t.Errorf("expected version '1.2.3', got %q", metadata.Version)
		}
	})

	t.Run("missing Chart.yaml returns error", func(t *testing.T) {
		tmpDir := t.TempDir()

		_, err := GetChartMetadata(tmpDir)
		if err == nil {
			t.Fatal("expected error for missing Chart.yaml, got nil")
		}
		if !errors.Is(err, os.ErrNotExist) {
			t.Errorf("expected IsNotExist error, got: %v", err)
		}
	})

	t.Run("invalid YAML returns error", func(t *testing.T) {
		tmpDir := t.TempDir()
		chartYaml := `this is not valid: yaml: : :`
		chartPath := filepath.Join(tmpDir, "Chart.yaml")
		if err := os.WriteFile(chartPath, []byte(chartYaml), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := GetChartMetadata(tmpDir)
		if err == nil {
			t.Fatal("expected error for invalid YAML, got nil")
		}
	})

	t.Run("missing name returns error", func(t *testing.T) {
		tmpDir := t.TempDir()
		chartYaml := `apiVersion: v2
version: 1.2.3
`
		chartPath := filepath.Join(tmpDir, "Chart.yaml")
		if err := os.WriteFile(chartPath, []byte(chartYaml), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := GetChartMetadata(tmpDir)
		if err == nil {
			t.Fatal("expected error for missing name, got nil")
		}
	})

	t.Run("missing version returns error", func(t *testing.T) {
		tmpDir := t.TempDir()
		chartYaml := `apiVersion: v2
name: my-chart
`
		chartPath := filepath.Join(tmpDir, "Chart.yaml")
		if err := os.WriteFile(chartPath, []byte(chartYaml), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := GetChartMetadata(tmpDir)
		if err == nil {
			t.Fatal("expected error for missing version, got nil")
		}
	})
}
