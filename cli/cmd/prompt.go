package cmd

import (
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/tools"
)

// errNonInteractive is returned when a prompt is required but the process is
// non-interactive (CI set, or stdin is not a terminal).
var errNonInteractive = errors.New("interactive prompt required but stdin is not a terminal")

// ensureInteractive fails fast when prompts would hang (CI / piped stdin).
func ensureInteractive() error {
	if tools.IsNonInteractive() {
		return errNonInteractive
	}
	return nil
}

// runPrompt runs a promptui.Prompt after verifying stdin is interactive.
func runPrompt(prompt promptui.Prompt) (string, error) {
	if err := ensureInteractive(); err != nil {
		return "", err
	}
	return prompt.Run()
}

// runSelect runs a promptui.Select after verifying stdin is interactive.
func runSelect(prompt promptui.Select) (int, string, error) {
	if err := ensureInteractive(); err != nil {
		return 0, "", err
	}
	return prompt.Run()
}
