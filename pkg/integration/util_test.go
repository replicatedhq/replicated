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

func AssertAPIRequests(t *testing.T, wantMethod string, wantPath string, apiCallLogFilename string) {
	apiCallLog, err := os.ReadFile(apiCallLogFilename)
	if err != nil {
		t.Errorf("error reading api call log: %v", err)
	}

	if string(apiCallLog) != fmt.Sprintf("%s:%s\n", wantMethod, wantPath) {
		t.Errorf("got %s, want %s", apiCallLog, fmt.Sprintf("%s:%s\n", wantMethod, wantPath))
	}
}
