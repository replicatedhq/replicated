package cmd

import (
	"testing"

	"github.com/replicatedhq/replicated/pkg/types"
)

func Test_areReleaseChartsReady(t *testing.T) {
	tests := []struct {
		name    string
		charts  []types.Chart
		want    bool
		wantErr bool
	}{
		{"nil charts", nil, false, false},
		{"no charts", []types.Chart{}, false, false},
		{"one chart, no status", []types.Chart{{}}, false, true},
		{"one chart, status unkown", []types.Chart{{Status: types.ChartStatusUnknown}}, false, false},
		{"one chart, status pushing", []types.Chart{{Status: types.ChartStatusPushing}}, false, false},
		{"one chart, status pushed", []types.Chart{{Status: types.ChartStatusPushed}}, true, false},
		{"one chart, status error", []types.Chart{{Status: types.ChartStatusError}}, false, true},
		{"two charts, status pushed", []types.Chart{{Status: types.ChartStatusPushed}, {Status: types.ChartStatusPushed}}, true, false},
		{"two charts, status pushed and pushing", []types.Chart{{Status: types.ChartStatusPushed}, {Status: types.ChartStatusPushing}}, false, false},
		{"two charts, status pushed and error", []types.Chart{{Status: types.ChartStatusPushed}, {Status: types.ChartStatusError}}, false, true},
		{"two charts, status pushed and unknown", []types.Chart{{Status: types.ChartStatusPushed}, {Status: types.ChartStatusUnknown}}, false, false},
		{"two charts, status pushing and error", []types.Chart{{Status: types.ChartStatusPushing}, {Status: types.ChartStatusError}}, false, true},
		{"two charts, status pushing and unknown", []types.Chart{{Status: types.ChartStatusPushing}, {Status: types.ChartStatusUnknown}}, false, false},
		{"two charts, status error and unknown", []types.Chart{{Status: types.ChartStatusError}, {Status: types.ChartStatusUnknown}}, false, true},
		{"two charts, status error and error", []types.Chart{{Status: types.ChartStatusError}, {Status: types.ChartStatusError}}, false, true},
		{"three charts, status pushed, pushing, and error", []types.Chart{{Status: types.ChartStatusPushed}, {Status: types.ChartStatusPushing}, {Status: types.ChartStatusError}}, false, true},
		{"three charts, status pushed, pushing, and unknown", []types.Chart{{Status: types.ChartStatusPushed}, {Status: types.ChartStatusPushing}, {Status: types.ChartStatusUnknown}}, false, false},
		{"three charts, status pushed, error, and unknown", []types.Chart{{Status: types.ChartStatusPushed}, {Status: types.ChartStatusError}, {Status: types.ChartStatusUnknown}}, false, true},
		{"four charts, status pushed, pushing, error, and unknown", []types.Chart{{Status: types.ChartStatusPushed}, {Status: types.ChartStatusPushing}, {Status: types.ChartStatusError}, {Status: types.ChartStatusUnknown}}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := areReleaseChartsReady(tt.charts)
			if (err != nil) != tt.wantErr {
				t.Errorf("areReleaseChartsReady() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("areReleaseChartsReady() = %v, want %v", got, tt.want)
			}
		})
	}
}
