package tools

import (
	"strings"
	"testing"
)

func TestReadAllLimited(t *testing.T) {
	t.Parallel()

	got, err := readAllLimited(strings.NewReader("abc"), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != "abc" {
		t.Fatalf("got %q, want abc", got)
	}

	_, err = readAllLimited(strings.NewReader(strings.Repeat("x", 5)), 4)
	if err == nil {
		t.Fatal("expected error when body exceeds limit")
	}
}
