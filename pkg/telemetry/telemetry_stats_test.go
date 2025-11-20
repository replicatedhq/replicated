package telemetry

import (
	"testing"
)

func TestRecordStats(t *testing.T) {
	tel := &Telemetry{
		disabled: false,
		command:  "test command",
	}

	stats := ResourceStats{
		HelmChartsCount:     3,
		ManifestsCount:      45,
		PreflightsCount:     2,
		SupportBundlesCount: 1,
		ToolVersions: map[string]string{
			"helm":    "3.12.0",
			"kubectl": "1.28.0",
		},
	}

	tel.RecordStats(stats)

	if tel.stats == nil {
		t.Error("Stats not recorded")
	}
	if tel.stats.HelmChartsCount != 3 {
		t.Errorf("Expected 3 helm charts, got %d", tel.stats.HelmChartsCount)
	}
	if tel.stats.ManifestsCount != 45 {
		t.Errorf("Expected 45 manifests, got %d", tel.stats.ManifestsCount)
	}
	if tel.stats.PreflightsCount != 2 {
		t.Errorf("Expected 2 preflights, got %d", tel.stats.PreflightsCount)
	}
	if tel.stats.SupportBundlesCount != 1 {
		t.Errorf("Expected 1 support bundle, got %d", tel.stats.SupportBundlesCount)
	}
	if len(tel.stats.ToolVersions) != 2 {
		t.Errorf("Expected 2 tool versions, got %d", len(tel.stats.ToolVersions))
	}
}

func TestRecordStats_WhenDisabled(t *testing.T) {
	tel := &Telemetry{
		disabled: true,
	}

	stats := ResourceStats{
		HelmChartsCount: 3,
	}

	tel.RecordStats(stats)

	if tel.stats != nil {
		t.Error("Stats should not be recorded when disabled")
	}
}

func TestRecordStats_Overwrite(t *testing.T) {
	tel := &Telemetry{
		disabled: false,
		command:  "test",
	}

	// Record first stats
	stats1 := ResourceStats{
		HelmChartsCount: 3,
	}
	tel.RecordStats(stats1)

	// Record second stats (should overwrite)
	stats2 := ResourceStats{
		HelmChartsCount: 5,
	}
	tel.RecordStats(stats2)

	if tel.stats.HelmChartsCount != 5 {
		t.Errorf("Expected stats to be overwritten, got %d", tel.stats.HelmChartsCount)
	}
}

