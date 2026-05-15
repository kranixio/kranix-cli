package commands

import (
	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `To load completions:

Bash:
  $ source <(kranix completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ kranix completion bash > /etc/bash_completion.d/kranix
  # macOS:
  $ kranix completion bash > /usr/local/etc/bash_completion.d/kranix

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ kranix completion zsh > "${fpath[1]}/_kranix"

  # You will need to start a new shell for this setup to take effect.

fish:
  $ kranix completion fish | source

  # To load completions for each session, execute once:
  $ kranix completion fish > ~/.config/fish/completions/kranix.fish

PowerShell:
  PS> kranix completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> kranix completion powershell > kranix.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(cmd.OutOrStdout())
		case "zsh":
			cmd.Root().GenZshCompletion(cmd.OutOrStdout())
		case "fish":
			cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
		case "powershell":
			cmd.Root().GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
		}
	},
}
