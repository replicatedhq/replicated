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
		name            string
		cli             string
		wantFormat      format
		wantLines       int
		wantAPIRequests []string
		ignoreCLIOutput bool
	}{
		{
			name:       "app-ls-empty",
			cli:        "app ls",
			wantFormat: FormatTable,
			wantLines:  1,
			wantAPIRequests: []string{
				"GET:/v3/apps?excludeChannels=false",
			},
		},
		{
			name:       "app-ls-empty",
			cli:        "app ls --output json",
			wantFormat: FormatJSON,
			wantAPIRequests: []string{
				"GET:/v3/apps?excludeChannels=false",
			},
		},
		{
			name:       "app-ls-single",
			cli:        "app ls",
			wantFormat: FormatTable,
			wantLines:  2,
			wantAPIRequests: []string{
				"GET:/v3/apps?excludeChannels=false",
			},
		},
		{
			name:       "app-ls-single",
			cli:        "app ls --output json",
			wantFormat: FormatJSON,
			wantAPIRequests: []string{
				"GET:/v3/apps?excludeChannels=false",
			},
		},
		{
			name: "app-rm",
			cli:  "app rm slug --force",
			wantAPIRequests: []string{
				"GET:/v3/apps?excludeChannels=true",
				"DELETE:/v3/app/id",
			},
			ignoreCLIOutput: true,
		},
		{
			name:       "app-create",
			cli:        "app create name",
			wantFormat: FormatTable,
			wantLines:  2,
			wantAPIRequests: []string{
				"POST:/v3/app",
			},
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

			if !tt.ignoreCLIOutput {
				AssertCLIOutput(t, string(out), tt.wantFormat, tt.wantLines)
			}

			AssertAPIRequests(t, tt.wantAPIRequests, apiCallLog.Name())
		})
	}
}
