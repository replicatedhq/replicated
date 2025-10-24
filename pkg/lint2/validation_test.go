package lint2

import (
	"strings"
	"testing"
)

func TestValidateChartToHelmChartMapping_AllChartsHaveManifests(t *testing.T) {
	// Setup: 3 charts, 3 matching HelmChart manifests
	charts := []ChartWithMetadata{
		{Path: "/path/to/chart1", Name: "app1", Version: "1.0.0"},
		{Path: "/path/to/chart2", Name: "app2", Version: "2.0.0"},
		{Path: "/path/to/chart3", Name: "app3", Version: "1.5.0"},
	}

	helmCharts := map[string]*HelmChartManifest{
		"app1:1.0.0": {Name: "app1", ChartVersion: "1.0.0", FilePath: "/manifests/app1.yaml"},
		"app2:2.0.0": {Name: "app2", ChartVersion: "2.0.0", FilePath: "/manifests/app2.yaml"},
		"app3:1.5.0": {Name: "app3", ChartVersion: "1.5.0", FilePath: "/manifests/app3.yaml"},
	}

	result, err := ValidateChartToHelmChartMapping(charts, helmCharts)

	// Should succeed with no errors or warnings
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if len(result.Warnings) != 0 {
		t.Errorf("expected no warnings, got %d: %v", len(result.Warnings), result.Warnings)
	}
}

func TestValidateChartToHelmChartMapping_SingleChartMissing(t *testing.T) {
	// Setup: 1 chart, 0 HelmChart manifests
	charts := []ChartWithMetadata{
		{Path: "/path/to/chart", Name: "my-app", Version: "1.0.0"},
	}

	helmCharts := map[string]*HelmChartManifest{} // Empty - no manifests

	result, err := ValidateChartToHelmChartMapping(charts, helmCharts)

	// Should return error
	if err == nil {
		t.Fatal("expected error for missing HelmChart manifest, got nil")
	}

	// Should be a MultipleChartsMissingHelmChartsError
	multiErr, ok := err.(*MultipleChartsMissingHelmChartsError)
	if !ok {
		t.Fatalf("expected *MultipleChartsMissingHelmChartsError, got %T", err)
	}

	// Should report exactly 1 missing chart
	if len(multiErr.MissingCharts) != 1 {
		t.Errorf("expected 1 missing chart, got %d", len(multiErr.MissingCharts))
	}

	// Verify error message contains chart details
	errMsg := err.Error()
	if !strings.Contains(errMsg, "my-app") {
		t.Errorf("error message should contain chart name 'my-app': %s", errMsg)
	}
	if !strings.Contains(errMsg, "1.0.0") {
		t.Errorf("error message should contain version '1.0.0': %s", errMsg)
	}
	if !strings.Contains(errMsg, "/path/to/chart") {
		t.Errorf("error message should contain chart path: %s", errMsg)
	}

	// Result should be nil on error
	if result != nil {
		t.Error("expected nil result on error")
	}
}

func TestValidateChartToHelmChartMapping_MultipleChartsMissing(t *testing.T) {
	// Setup: 3 charts, 1 matching HelmChart (2 missing)
	charts := []ChartWithMetadata{
		{Path: "/charts/frontend", Name: "frontend", Version: "1.0.0"},
		{Path: "/charts/backend", Name: "backend", Version: "2.1.0"},
		{Path: "/charts/database", Name: "database", Version: "1.5.0"},
	}

	helmCharts := map[string]*HelmChartManifest{
		"backend:2.1.0": {Name: "backend", ChartVersion: "2.1.0", FilePath: "/manifests/backend.yaml"},
		// frontend and database are missing
	}

	result, err := ValidateChartToHelmChartMapping(charts, helmCharts)

	// Should return error
	if err == nil {
		t.Fatal("expected error for missing HelmChart manifests, got nil")
	}

	// Should be a MultipleChartsMissingHelmChartsError
	multiErr, ok := err.(*MultipleChartsMissingHelmChartsError)
	if !ok {
		t.Fatalf("expected *MultipleChartsMissingHelmChartsError, got %T", err)
	}

	// Should report exactly 2 missing charts
	if len(multiErr.MissingCharts) != 2 {
		t.Errorf("expected 2 missing charts, got %d", len(multiErr.MissingCharts))
	}

	// Verify batch error message format
	errMsg := err.Error()
	if !strings.Contains(errMsg, "2 charts missing") {
		t.Errorf("error message should mention '2 charts missing': %s", errMsg)
	}

	// Should list both missing charts
	if !strings.Contains(errMsg, "frontend") {
		t.Errorf("error message should contain 'frontend': %s", errMsg)
	}
	if !strings.Contains(errMsg, "database") {
		t.Errorf("error message should contain 'database': %s", errMsg)
	}

	// Should NOT mention the chart that has a manifest
	if strings.Contains(errMsg, "backend") {
		t.Errorf("error message should not contain 'backend' (it has a manifest): %s", errMsg)
	}

	// Verify error message has actionable guidance
	if !strings.Contains(errMsg, "HelmChart manifest") {
		t.Errorf("error message should mention HelmChart manifest: %s", errMsg)
	}

	// Result should be nil on error
	if result != nil {
		t.Error("expected nil result on error")
	}
}

func TestValidateChartToHelmChartMapping_OrphanedHelmChartManifest(t *testing.T) {
	// Setup: 1 chart, 2 HelmChart manifests (1 orphaned)
	charts := []ChartWithMetadata{
		{Path: "/charts/current-app", Name: "current-app", Version: "1.0.0"},
	}

	helmCharts := map[string]*HelmChartManifest{
		"current-app:1.0.0": {Name: "current-app", ChartVersion: "1.0.0", FilePath: "/manifests/current.yaml"},
		"old-app:1.0.0":     {Name: "old-app", ChartVersion: "1.0.0", FilePath: "/manifests/old.yaml"},
	}

	result, err := ValidateChartToHelmChartMapping(charts, helmCharts)

	// Should succeed (orphans are warnings, not errors)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Should have exactly 1 warning
	if len(result.Warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d: %v", len(result.Warnings), result.Warnings)
	}

	// Verify warning message
	warning := result.Warnings[0]
	if !strings.Contains(warning, "old-app") {
		t.Errorf("warning should contain orphaned chart name 'old-app': %s", warning)
	}
	if !strings.Contains(warning, "old-app:1.0.0") {
		t.Errorf("warning should contain orphaned chart key 'old-app:1.0.0': %s", warning)
	}
	if !strings.Contains(warning, "/manifests/old.yaml") {
		t.Errorf("warning should contain orphaned manifest path: %s", warning)
	}
	if !strings.Contains(warning, "no corresponding chart") {
		t.Errorf("warning should explain the issue: %s", warning)
	}
}

func TestValidateChartToHelmChartMapping_EmptyCharts(t *testing.T) {
	// Setup: 0 charts, 2 HelmChart manifests
	charts := []ChartWithMetadata{} // Empty

	helmCharts := map[string]*HelmChartManifest{
		"app1:1.0.0": {Name: "app1", ChartVersion: "1.0.0", FilePath: "/manifests/app1.yaml"},
		"app2:2.0.0": {Name: "app2", ChartVersion: "2.0.0", FilePath: "/manifests/app2.yaml"},
	}

	result, err := ValidateChartToHelmChartMapping(charts, helmCharts)

	// Should succeed (no charts to validate)
	if err != nil {
		t.Fatalf("expected no error with empty charts, got: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Should warn about all orphaned manifests (both)
	if len(result.Warnings) != 2 {
		t.Errorf("expected 2 warnings for orphaned manifests, got %d: %v", len(result.Warnings), result.Warnings)
	}
}

func TestValidateChartToHelmChartMapping_EmptyManifests(t *testing.T) {
	// Setup: 2 charts, 0 HelmChart manifests
	charts := []ChartWithMetadata{
		{Path: "/charts/app1", Name: "app1", Version: "1.0.0"},
		{Path: "/charts/app2", Name: "app2", Version: "2.0.0"},
	}

	helmCharts := map[string]*HelmChartManifest{} // Empty

	result, err := ValidateChartToHelmChartMapping(charts, helmCharts)

	// Should error (all charts missing manifests)
	if err == nil {
		t.Fatal("expected error when all charts missing manifests, got nil")
	}

	// Should be a batch error
	multiErr, ok := err.(*MultipleChartsMissingHelmChartsError)
	if !ok {
		t.Fatalf("expected *MultipleChartsMissingHelmChartsError, got %T", err)
	}

	// Should report both charts
	if len(multiErr.MissingCharts) != 2 {
		t.Errorf("expected 2 missing charts, got %d", len(multiErr.MissingCharts))
	}

	// Result should be nil on error
	if result != nil {
		t.Error("expected nil result on error")
	}
}

func TestValidateChartToHelmChartMapping_VersionMismatch(t *testing.T) {
	// Setup: Chart "app:1.0.0", HelmChart "app:2.0.0" (version mismatch)
	charts := []ChartWithMetadata{
		{Path: "/charts/app", Name: "app", Version: "1.0.0"},
	}

	helmCharts := map[string]*HelmChartManifest{
		"app:2.0.0": {Name: "app", ChartVersion: "2.0.0", FilePath: "/manifests/app.yaml"},
		// app:1.0.0 is missing
	}

	result, err := ValidateChartToHelmChartMapping(charts, helmCharts)

	// Should error (version doesn't match, so chart is considered missing)
	if err == nil {
		t.Fatal("expected error for version mismatch, got nil")
	}

	// Should report the chart with correct version as missing
	errMsg := err.Error()
	if !strings.Contains(errMsg, "app") {
		t.Errorf("error message should contain chart name: %s", errMsg)
	}
	if !strings.Contains(errMsg, "1.0.0") {
		t.Errorf("error message should contain chart version 1.0.0: %s", errMsg)
	}

	// Result should be nil on error
	if result != nil {
		t.Error("expected nil result on error")
	}

	// There should be a warning about the orphaned 2.0.0 manifest
	// (Actually, this won't happen because result is nil on error)
	// The validation reports the chart as missing, user needs to fix the version
}

func TestValidateChartToHelmChartMapping_NameMismatch(t *testing.T) {
	// Setup: Chart "my-app:1.0.0", HelmChart "my-app-v2:1.0.0" (name mismatch)
	charts := []ChartWithMetadata{
		{Path: "/charts/my-app", Name: "my-app", Version: "1.0.0"},
	}

	helmCharts := map[string]*HelmChartManifest{
		"my-app-v2:1.0.0": {Name: "my-app-v2", ChartVersion: "1.0.0", FilePath: "/manifests/my-app-v2.yaml"},
		// my-app:1.0.0 is missing
	}

	result, err := ValidateChartToHelmChartMapping(charts, helmCharts)

	// Should error (name doesn't match)
	if err == nil {
		t.Fatal("expected error for name mismatch, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "my-app") {
		t.Errorf("error message should contain chart name 'my-app': %s", errMsg)
	}

	// Result should be nil on error
	if result != nil {
		t.Error("expected nil result on error")
	}
}

func TestValidateChartToHelmChartMapping_MultipleOrphans(t *testing.T) {
	// Setup: 1 chart, 3 HelmChart manifests (2 orphaned)
	charts := []ChartWithMetadata{
		{Path: "/charts/current", Name: "current", Version: "3.0.0"},
	}

	helmCharts := map[string]*HelmChartManifest{
		"current:3.0.0": {Name: "current", ChartVersion: "3.0.0", FilePath: "/manifests/current.yaml"},
		"old-v1:1.0.0":  {Name: "old-v1", ChartVersion: "1.0.0", FilePath: "/manifests/old-v1.yaml"},
		"old-v2:2.0.0":  {Name: "old-v2", ChartVersion: "2.0.0", FilePath: "/manifests/old-v2.yaml"},
	}

	result, err := ValidateChartToHelmChartMapping(charts, helmCharts)

	// Should succeed
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Should have 2 warnings
	if len(result.Warnings) != 2 {
		t.Fatalf("expected 2 warnings, got %d: %v", len(result.Warnings), result.Warnings)
	}

	// Both warnings should mention orphaned manifests
	warningsStr := strings.Join(result.Warnings, " ")
	if !strings.Contains(warningsStr, "old-v1") {
		t.Errorf("warnings should contain 'old-v1': %v", result.Warnings)
	}
	if !strings.Contains(warningsStr, "old-v2") {
		t.Errorf("warnings should contain 'old-v2': %v", result.Warnings)
	}
}

func TestMultipleChartsMissingHelmChartsError_SingleChartMessage(t *testing.T) {
	// Test the error message format for a single chart
	err := &MultipleChartsMissingHelmChartsError{
		MissingCharts: []ChartMissingHelmChartInfo{
			{
				ChartPath:    "/charts/my-app",
				ChartName:    "my-app",
				ChartVersion: "1.0.0",
			},
		},
	}

	msg := err.Error()

	// Should use singular message format
	if strings.Contains(msg, "charts missing") {
		t.Errorf("single chart error should not use plural format: %s", msg)
	}

	// Should contain chart details
	if !strings.Contains(msg, "my-app") {
		t.Errorf("error should contain chart name: %s", msg)
	}
	if !strings.Contains(msg, "1.0.0") {
		t.Errorf("error should contain version: %s", msg)
	}
	if !strings.Contains(msg, "/charts/my-app") {
		t.Errorf("error should contain path: %s", msg)
	}

	// Should contain guidance
	if !strings.Contains(msg, "HelmChart manifest") {
		t.Errorf("error should mention HelmChart manifest: %s", msg)
	}
}

func TestMultipleChartsMissingHelmChartsError_MultipleChartsMessage(t *testing.T) {
	// Test the error message format for multiple charts
	err := &MultipleChartsMissingHelmChartsError{
		MissingCharts: []ChartMissingHelmChartInfo{
			{ChartPath: "/charts/app1", ChartName: "app1", ChartVersion: "1.0.0"},
			{ChartPath: "/charts/app2", ChartName: "app2", ChartVersion: "2.0.0"},
			{ChartPath: "/charts/app3", ChartName: "app3", ChartVersion: "3.0.0"},
		},
	}

	msg := err.Error()

	// Should use plural message format
	if !strings.Contains(msg, "3 charts missing") {
		t.Errorf("multiple charts error should show count: %s", msg)
	}

	// Should list all charts
	if !strings.Contains(msg, "app1") || !strings.Contains(msg, "1.0.0") {
		t.Errorf("error should contain app1: %s", msg)
	}
	if !strings.Contains(msg, "app2") || !strings.Contains(msg, "2.0.0") {
		t.Errorf("error should contain app2: %s", msg)
	}
	if !strings.Contains(msg, "app3") || !strings.Contains(msg, "3.0.0") {
		t.Errorf("error should contain app3: %s", msg)
	}

	// Should contain guidance
	if !strings.Contains(msg, "HelmChart manifest") {
		t.Errorf("error should mention HelmChart manifest: %s", msg)
	}
	if !strings.Contains(msg, "manifests") && !strings.Contains(msg, ".replicated") {
		t.Errorf("error should mention configuration: %s", msg)
	}
}
