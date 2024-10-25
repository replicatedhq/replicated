package integration

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestRelease(t *testing.T) {
	tests := []struct {
		name            string
		cli             string
		wantFormat      format
		wantLines       int
		wantAPIRequests []string
		ignoreCLIOutput bool
	}{
		{
			name:       "release-ls",
			cli:        "release ls",
			wantFormat: FormatTable,
			wantLines:  1,
			wantAPIRequests: []string{
				"GET:/v3/app/id/releases",
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

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, CLIPath(), args...)

			if ctx.Err() == context.DeadlineExceeded {
				t.Fatalf("Command execution timed out")
			}

			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("error running cli: %v", err)
			}

			fmt.Printf("out: %s\n", string(out))
			if !tt.ignoreCLIOutput {
				AssertCLIOutput(t, string(out), tt.wantFormat, tt.wantLines)
			}

			AssertAPIRequests(t, tt.wantAPIRequests, apiCallLog.Name())
		})
	}
}
