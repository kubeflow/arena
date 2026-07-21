package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion <shell>",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for arena.
Supported shells: bash, zsh, fish, powershell.

Usage:
  arena completion bash > /etc/bash_completion.d/arena
  arena completion zsh > "${fpath[1]}/_arena"
  arena completion fish > ~/.config/fish/completions/arena.fish
  arena completion powershell > arena.ps1`,
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	Args:      cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		default:
			return fmt.Errorf("unsupported shell: %q (must be bash, zsh, fish, or powershell)", args[0])
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
