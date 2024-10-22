package integration

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestApp(t *testing.T) {
	tests := []struct {
		name       string
		cli        string
		wantFormat format
		wantLines  int
		wantMethod string
		wantPath   string
	}{
		{
			name:       "app-ls-empty",
			cli:        "app ls",
			wantFormat: FormatTable,
			wantLines:  1,
			wantMethod: "GET",
			wantPath:   "/v3/apps?excludeChannels=false",
		},
		{
			name:       "app-ls-empty",
			cli:        "app ls --output json",
			wantFormat: FormatJSON,
			wantMethod: "GET",
			wantPath:   "/v3/apps?excludeChannels=false",
		},
		{
			name:       "app-ls-single",
			cli:        "app ls",
			wantFormat: FormatTable,
			wantLines:  2,
			wantMethod: "GET",
			wantPath:   "/v3/apps?excludeChannels=false",
		},
		{
			name:       "app-ls-single",
			cli:        "app ls --output json",
			wantFormat: FormatJSON,
			wantMethod: "GET",
			wantPath:   "/v3/apps?excludeChannels=false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := strings.Split(tt.cli, " ")
			args = append(args, "--integration-test", tt.name)

			apiCallLog, err := os.CreateTemp("", "")
			if err != nil {
				log.Fatal(err)
			}

			defer os.RemoveAll(apiCallLog.Name())

			args = append(args, "--log-api-calls", apiCallLog.Name())

			cmd := exec.Command(CLIPath(), args...)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("error running cli: %v", err)
			}

			AssertCLIOutput(t, string(out), tt.wantFormat, tt.wantLines)
			AssertAPIRequests(t, tt.wantMethod, tt.wantPath, apiCallLog.Name())
		})
	}
}
