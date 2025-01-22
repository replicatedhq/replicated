package cmd

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func (r *runners) InitCompletionCommand(parent *cobra.Command) *cobra.Command {
	cmd := NewCmdCompletion(r.w, parent.Name())
	parent.AddCommand(cmd)
	return cmd
}

var (
	ErrCompletionShellNotSpecified = errors.New("Shell not specified.")
	ErrCompletionTooMayArguments   = errors.New("Too many arguments. Expected only the shell type.")
)

func NewCmdCompletion(out io.Writer, parentName string) *cobra.Command {

	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Example: fmt.Sprintf(`To load completions:

Bash:

	This script depends on the 'bash-completion' package.
	If it is not installed already, you can install it via your OS's package manager.

	$ source <(%[1]s completion bash)

	# To load completions for each session, execute once:
	# Linux:
	$ %[1]s completion bash > /etc/bash_completion.d/%[1]s
	# macOS:
	$ %[1]s completion bash > $(brew --prefix)/etc/bash_completion.d/%[1]s

Zsh:

	# If shell completion is not already enabled in your environment,
	# you will need to enable it.  You can execute the following once:

	$ echo "autoload -U compinit; compinit" >> ~/.zshrc

	# To load completions for each session, execute once:
	$ %[1]s completion zsh > "${fpath[1]}/_%[1]s"

	# You will need to start a new shell for this setup to take effect.

fish:

	$ %[1]s completion fish | source

	# To load completions for each session, execute once:
	$ %[1]s completion fish > ~/.config/fish/completions/%[1]s.fish

PowerShell:

	PS> %[1]s completion powershell | Out-String | Invoke-Expression

	# To load completions for every new session, run:
	PS> %[1]s completion powershell > %[1]s.ps1
	# and source this file from your PowerShell profile.
`, parentName),

		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			RunCompletion(out, cmd, args)
		},
	}

}

func RunCompletion(out io.Writer, cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return ErrCompletionShellNotSpecified
	}

	if len(args) > 1 {
		return ErrCompletionTooMayArguments
	}

	switch args[0] {
	case "bash":
		return cmd.Root().GenBashCompletion(out)
	case "zsh":
		zshHead := fmt.Sprintf("#compdef %[1]s\ncompdef _%[1]s %[1]s\n", cmd.Root().Name())
		out.Write([]byte(zshHead))
		return cmd.Root().GenZshCompletion(out)
	case "fish":
		return cmd.Root().GenFishCompletion(out, true)
	case "powershell":
		return cmd.Root().GenPowerShellCompletionWithDesc(out)
	default:
		return fmt.Errorf("Unsupported shell type %q.", args[0])
	}

}
