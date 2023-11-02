package cmd

import (
	"os"
	"testing"
	"text/tabwriter"
)

func TestRegistryAddGCR(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		wantErr       bool
		wantErrString string
	}{
		{
			name:          "endpoint required",
			args:          []string{"registry", "add", "gcr"},
			wantErr:       true,
			wantErrString: "endpoint must be specified, serviceaccountkey or serviceaccountkey-stdin must be specified",
		},
		{
			name:          "service account key required",
			args:          []string{"registry", "add", "gcr", "--endpoint", "gcr.io"},
			wantErr:       true,
			wantErrString: "serviceaccountkey or serviceaccountkey-stdin must be specified",
		},
		{
			name:          "invalid service account key",
			args:          []string{"registry", "add", "gcr", "--endpoint", "gcr.io", "--serviceaccountkey", "./testdata/invalid-gcr-service-account-key.json"},
			wantErr:       true,
			wantErrString: "Not valid json key file",
		},
		{
			name:          "service account key file not found",
			args:          []string{"registry", "add", "gcr", "--endpoint", "gcr.io", "--serviceaccountkey", "./testdata/does-not-exist.json"},
			wantErr:       true,
			wantErrString: "read service account key: open ./testdata/does-not-exist.json: no such file or directory",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := GetRootCmd()
			w := tabwriter.NewWriter(os.Stdout, minWidth, tabWidth, padding, padChar, tabwriter.TabIndent)
			runCmds := &runners{
				rootCmd: cmd,
				stdin:   os.Stdin,
				w:       w,
			}

			registryCmd := runCmds.InitRegistryCommand(runCmds.rootCmd)
			runCmds.InitRegistryCommand(registryCmd)
			registryAddCmd := runCmds.InitRegistryAdd(registryCmd)
			runCmds.InitRegistryAddGCR(registryAddCmd)
			runCmds.rootCmd.SetArgs(test.args)
			err := runCmds.rootCmd.Execute()
			if test.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if err.Error() != test.wantErrString {
					t.Errorf("expected error string %q, got %q", test.wantErrString, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}

		})
	}
}
