package cmd

import (
	"testing"

	"github.com/replicatedhq/replicated/pkg/types"
	"helm.sh/helm/v3/pkg/cli/values"
)

func Test_areReleaseChartsReady(t *testing.T) {
	tests := []struct {
		name    string
		charts  []types.Chart
		want    bool
		wantErr bool
	}{
		{"nil charts", nil, true, false},
		{"no charts", []types.Chart{}, true, false},
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
			got, err := areReleaseChartsPushed(tt.charts)
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

func Test_validateClusterPrepareFlags(t *testing.T) {
	type args struct {
		args runnerArgs
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"no flags", args{runnerArgs{}}, true},
		{"chart and yaml", args{runnerArgs{prepareClusterChart: "foo", prepareClusterYaml: "bar"}}, true},
		{"chart and yaml file", args{runnerArgs{prepareClusterChart: "foo", prepareClusterYamlFile: "bar"}}, true},
		{"chart and yaml dir", args{runnerArgs{prepareClusterChart: "foo", prepareClusterYamlDir: "bar"}}, true},
		{"yaml and yaml file", args{runnerArgs{prepareClusterYaml: "foo", prepareClusterYamlFile: "bar"}}, true},
		{"yaml and yaml dir", args{runnerArgs{prepareClusterYaml: "foo", prepareClusterYamlDir: "bar"}}, true},
		{"yaml file and yaml dir", args{runnerArgs{prepareClusterYamlFile: "foo", prepareClusterYamlDir: "bar"}}, true},
		{"chart and shared-password", args{runnerArgs{prepareClusterChart: "foo", prepareClusterKotsSharedPassword: "bar"}}, true},
		{"chart and config-values-file", args{runnerArgs{prepareClusterChart: "foo", prepareClusterKotsConfigValuesFile: "bar"}}, true},
		{"yaml and empty shared-password", args{runnerArgs{prepareClusterYaml: "foo", prepareClusterKotsSharedPassword: ""}}, true},
		{"yaml file and empty shared-password", args{runnerArgs{prepareClusterYamlFile: "foo", prepareClusterKotsSharedPassword: ""}}, true},
		{"yaml dir and empty shared-password", args{runnerArgs{prepareClusterYamlDir: "foo", prepareClusterKotsSharedPassword: ""}}, true},
		{"yaml and value opts set", args{runnerArgs{prepareClusterYaml: "foo", prepareClusterKotsSharedPassword: "test", prepareClusterValueOpts: values.Options{Values: []string{"test"}}}}, true},
		{"yaml and value opts set files", args{runnerArgs{prepareClusterYaml: "foo", prepareClusterKotsSharedPassword: "test", prepareClusterValueOpts: values.Options{ValueFiles: []string{"test"}}}}, true},
		{"yaml and value opts set string values", args{runnerArgs{prepareClusterYaml: "foo", prepareClusterKotsSharedPassword: "test", prepareClusterValueOpts: values.Options{StringValues: []string{"test"}}}}, true},
		{"yaml and value opts set file values", args{runnerArgs{prepareClusterYaml: "foo", prepareClusterKotsSharedPassword: "test", prepareClusterValueOpts: values.Options{FileValues: []string{"test"}}}}, true},
		{"yaml and value opts set json values", args{runnerArgs{prepareClusterYaml: "foo", prepareClusterKotsSharedPassword: "test", prepareClusterValueOpts: values.Options{JSONValues: []string{"test"}}}}, true},
		{"yaml and value opts set literal values", args{runnerArgs{prepareClusterYaml: "foo", prepareClusterKotsSharedPassword: "test", prepareClusterValueOpts: values.Options{LiteralValues: []string{"test"}}}}, true},
		{"yaml and shared password", args{runnerArgs{prepareClusterYaml: "foo", prepareClusterKotsSharedPassword: "test"}}, false},
		{"yaml file and shared password", args{runnerArgs{prepareClusterYamlFile: "foo", prepareClusterKotsSharedPassword: "test"}}, false},
		{"charts only", args{runnerArgs{prepareClusterChart: "foo"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateClusterPrepareFlags(tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("validateClusterPrepareFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
