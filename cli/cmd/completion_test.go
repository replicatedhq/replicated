package cmd

import (
	"bytes"
	"os"
	"sync"
	"testing"

	"github.com/spf13/cobra"
)

// TestShellCompletionsNoHome verifies that completion commands work even when
// HOME is unset or invalid (e.g., during Nix builds). This is a regression
// test for https://github.com/replicatedhq/replicated/issues/535.
func TestShellCompletionsNoHome(t *testing.T) {
	// Save and clear HOME to simulate an environment where cache initialization
	// would fail.
	origHome := os.Getenv("HOME")
	os.Unsetenv("HOME")
	os.Unsetenv("XDG_CACHE_HOME")
	defer func() {
		os.Setenv("HOME", origHome)
	}()

	// Reset the lazy cache state so this test doesn't interfere with others.
	cacheOnce = sync.Once{}
	cacheInstance = nil
	cacheErr = nil
	defer func() {
		cacheOnce = sync.Once{}
		cacheInstance = nil
		cacheErr = nil
	}()

	parentCmd := &cobra.Command{
		Use: "replicated",
	}
	out := bytes.NewBufferString("")
	cmd := NewCmdCompletion(out, parentCmd.Name())
	parentCmd.AddCommand(cmd)

	err := RunCompletion(out, cmd, []string{"bash"})
	if err != nil {
		t.Fatalf("Unexpected error with no HOME: %v", err)
	}
	if out.Len() == 0 {
		t.Fatalf("Output was not written")
	}
}

// This unit-test was adapted from kubectl repo
// Link: https://github.com/kubernetes/kubectl/blob/826006cdb947f80a679ff1eb3cb53f183a6a9bf2/pkg/cmd/completion/completion_test.go
func TestShellCompletions(t *testing.T) {

	testCases := []struct {
		name          string
		args          []string
		expectedError string
	}{
		{
			name: "bash",
			args: []string{"bash"},
		},
		{
			name: "zsh",
			args: []string{"zsh"},
		},
		{
			name: "fish",
			args: []string{"fish"},
		},
		{
			name: "powershell",
			args: []string{"powershell"},
		},
		{
			name:          "no args",
			args:          []string{},
			expectedError: ErrCompletionShellNotSpecified.Error(),
		},
		{
			name:          "too many args",
			args:          []string{"bash", "zsh"},
			expectedError: ErrCompletionTooMayArguments.Error(),
		},
		{
			name:          "unsupported shell",
			args:          []string{"foo"},
			expectedError: `Unsupported shell type "foo".`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			parentCmd := &cobra.Command{
				Use: "replicated",
			}
			out := bytes.NewBufferString("")
			cmd := NewCmdCompletion(out, parentCmd.Name())
			parentCmd.AddCommand(cmd)
			err := RunCompletion(out, cmd, tc.args)
			if tc.expectedError == "" {
				if err != nil {
					tt.Fatalf("Unexpected error: %v", err)
				}
				if out.Len() == 0 {
					tt.Fatalf("Output was not written")
				}
			} else {
				if err == nil {
					tt.Fatalf("An error was expected but no error was returned")
				}
				if err.Error() != tc.expectedError {
					tt.Fatalf("Unexpected error: %v\nexpected: %v\n", err, tc.expectedError)
				}
			}
		})
	}
}
