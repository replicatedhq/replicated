package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

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
			name: "no args",
			args: []string{},
			expectedError: `Shell not specified.
See 'kubectl completion -h' for help and examples`,
		},
		{
			name: "too many args",
			args: []string{"bash", "zsh"},
			expectedError: `Too many arguments. Expected only the shell type.
See 'kubectl completion -h' for help and examples`,
		},
		{
			name: "unsupported shell",
			args: []string{"foo"},
			expectedError: `Unsupported shell type "foo".
See 'kubectl completion -h' for help and examples`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			parentCmd := &cobra.Command{
				Use: "replicated",
			}
			out := os.Stdout
			cmd := NewCmdCompletion(out, defaultBoilerPlate)
			parentCmd.AddCommand(cmd)
			err := RunCompletion(out, cmd, tc.args)
			if tc.expectedError == "" {
				if err != nil {
					tt.Fatalf("Unexpected error: %v", err)
				}
				if out.Len() == 0 {
					tt.Fatalf("Output was not written")
				}
				if !strings.Contains(out.String(), defaultBoilerPlate) {
					tt.Fatalf("Output does not contain boilerplate:\n%s", out.String())
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
