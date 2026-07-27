package cmd

import (
	"errors"
	"os"
	"testing"

	"github.com/manifoldco/promptui"
)

func TestEnsureInteractiveInCI(t *testing.T) {
	t.Setenv("CI", "true")

	if err := ensureInteractive(); err == nil {
		t.Fatal("expected ensureInteractive to fail when CI is set")
	} else if !errors.Is(err, errNonInteractive) {
		t.Fatalf("expected errNonInteractive, got %v", err)
	}

	if _, err := runPrompt(promptui.Prompt{Label: "test"}); err == nil {
		t.Fatal("expected runPrompt to fail when CI is set")
	}

	if _, _, err := runSelect(promptui.Select{Label: "test", Items: []string{"a"}}); err == nil {
		t.Fatal("expected runSelect to fail when CI is set")
	}
}

func TestEnsureInteractiveUnsetCI(t *testing.T) {
	// Clear CI for this test process; if stdin is not a TTY (common in test
	// runners), ensureInteractive should still report non-interactive.
	t.Setenv("CI", "")
	_ = os.Unsetenv("CI")

	err := ensureInteractive()
	// Either interactive (local terminal) or non-interactive (piped CI runner).
	// Just ensure it does not panic and returns a typed error when failing.
	if err != nil && !errors.Is(err, errNonInteractive) {
		t.Fatalf("unexpected error: %v", err)
	}
}
