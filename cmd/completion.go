package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish]",
	Short: "Generate shell completion script",
	Long: `Generate shell completion script for envswitch.

To load completions:

Bash:
  $ source <(envswitch completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ envswitch completion bash > /etc/bash_completion.d/envswitch
  # macOS:
  $ envswitch completion bash > /usr/local/etc/bash_completion.d/envswitch

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ envswitch completion zsh > "${fpath[1]}/_envswitch"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ envswitch completion fish | source

  # To load completions for each session, execute once:
  $ envswitch completion fish > ~/.config/fish/completions/envswitch.fish
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE:                  runCompletion,
}

func init() {
	rootCmd.AddCommand(completionCmd)
}

func runCompletion(cmd *cobra.Command, args []string) error {
	switch args[0] {
	case "bash":
		return cmd.Root().GenBashCompletion(os.Stdout)
	case "zsh":
		return cmd.Root().GenZshCompletion(os.Stdout)
	case "fish":
		return cmd.Root().GenFishCompletion(os.Stdout, true)
	}
	return nil
}
