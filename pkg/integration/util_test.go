package integration

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func getCommand(cliArgs []string, server *httptest.Server) *exec.Cmd {
	cmd := exec.Command(CLIPath(), cliArgs...)
	cmd.Env = append(cmd.Env, "REPLICATED_API_ORIGIN="+server.URL)
	cmd.Env = append(cmd.Env, "REPLICATED_API_TOKEN=test-token")
	cmd.Env = append(cmd.Env, "CI="+os.Getenv("CI")) // disable update checks
	cmd.Env = append(cmd.Env, "HOME="+os.TempDir())
	return cmd
}

func getCommandWithoutToken(cliArgs []string, server *httptest.Server) *exec.Cmd {
	cmd := exec.Command(CLIPath(), cliArgs...)
	cmd.Env = append(cmd.Env, "REPLICATED_API_ORIGIN="+server.URL)
	cmd.Env = append(cmd.Env, "CI="+os.Getenv("CI")) // disable update checks
	cmd.Env = append(cmd.Env, "HOME="+os.TempDir())
	return cmd
}

func AssertCLIOutput(t *testing.T, got string, wantFormat format, wantLines int) {
	gotFormat := FormatTable
	var i interface{}
	err := json.Unmarshal([]byte(got), &i)
	if err == nil {
		gotFormat = FormatJSON
	}

	gotLines := strings.Split(strings.TrimSpace(string(got)), "\n")

	if gotFormat != wantFormat {
		t.Errorf("got format %s, want %s: %s", gotFormat, wantFormat, got)
	}

	if wantFormat == FormatTable {
		if len(gotLines) != wantLines {
			t.Errorf("got %d lines, want %d: %s", len(gotLines), wantLines, got)
		}
	}
}

func AssertAPIRequests(t *testing.T, wantAPIRequests []string, apiCallLogFilename string) {
	apiCallLog, err := os.ReadFile(apiCallLogFilename)
	if err != nil {
		t.Errorf("error reading api call log: %v", err)
		return
	}

	gotAPIRequests := strings.Split(string(apiCallLog), "\n")

	cleaned := []string{}
	for _, line := range gotAPIRequests {
		if strings.TrimSpace(line) != "" {
			cleaned = append(cleaned, line)
		}
	}
	gotAPIRequests = cleaned

	if len(gotAPIRequests) != len(wantAPIRequests) {
		fmt.Printf("got: %v\n", gotAPIRequests)
		fmt.Printf("want: %v\n", wantAPIRequests)
		t.Errorf("got %d requests, want %d", len(gotAPIRequests), len(wantAPIRequests))
		return
	}

	for i, wantAPIRequest := range wantAPIRequests {
		gotAPIRequest := gotAPIRequests[i]
		if gotAPIRequest != wantAPIRequest {
			t.Errorf("got %s, want %s", gotAPIRequest, wantAPIRequest)
		}
	}
}
