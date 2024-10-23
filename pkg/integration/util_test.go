package integration

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

func AssertCLIOutput(t *testing.T, got string, wantFormat format, wantLines int) {
	gotFormat := FormatTable
	var i interface{}
	err := json.Unmarshal([]byte(got), &i)
	if err == nil {
		gotFormat = FormatJSON
	}

	gotLines := strings.Split(strings.TrimSpace(string(got)), "\n")

	if gotFormat != wantFormat {
		t.Errorf("got format %s, want %s", gotFormat, wantFormat)
	}

	if wantFormat == FormatTable {
		if len(gotLines) != wantLines {
			t.Errorf("got %d lines, want %d", len(gotLines), wantLines)
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
