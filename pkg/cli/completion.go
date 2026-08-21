package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion <shell>",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for arena-v2.
Supported shells: bash, zsh, fish, powershell.

Usage:
  arena-v2 completion bash > /etc/bash_completion.d/arena-v2
  arena-v2 completion zsh > "${fpath[1]}/_arena-v2"
  arena-v2 completion fish > ~/.config/fish/completions/arena-v2.fish
  arena-v2 completion powershell > arena-v2.ps1`,
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := cmd.Root()
			out := cmd.OutOrStdout()
			switch args[0] {
			case "bash":
				return root.GenBashCompletionV2(out, true)
			case "zsh":
				return root.GenZshCompletion(out)
			case "fish":
				return root.GenFishCompletion(out, true)
			case "powershell":
				return root.GenPowerShellCompletionWithDesc(out)
			default:
				return fmt.Errorf("unsupported shell: %q (must be bash, zsh, fish, or powershell)", args[0])
			}
		},
	}
	return cmd
}
