package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func (r *runners) InitCompletionCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: fmt.Sprintf(`To load completions:

	Bash:

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
	`, parent.Name()),
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				zshHead := fmt.Sprintf("#compdef %[1]s\ncompdef _%[1]s %[1]s\n", cmd.Root().Name())
				os.Stdout.Write([]byte(zshHead))
				cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
		},
	}
	parent.AddCommand(cmd)

	return cmd
}
