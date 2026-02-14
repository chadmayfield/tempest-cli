package cmd

import (
	"os"

	"github.com/chadmayfield/tempest-cli/internal/config"
	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate shell completion scripts for tempest.

To load completions:

Bash:
  $ source <(tempest completion bash)
  # Or add to ~/.bashrc:
  $ tempest completion bash > /etc/bash_completion.d/tempest

Zsh:
  $ source <(tempest completion zsh)
  # Or add to fpath:
  $ tempest completion zsh > "${fpath[1]}/_tempest"

Fish:
  $ tempest completion fish | source
  # Or persist:
  $ tempest completion fish > ~/.config/fish/completions/tempest.fish

PowerShell:
  PS> tempest completion powershell | Out-String | Invoke-Expression
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)

	// Add station name completion to the --station flag
	_ = rootCmd.RegisterFlagCompletionFunc("station", completeStationNames)
}

func completeStationNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	cfg, err := config.Load()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return cfg.StationNames(), cobra.ShellCompDirectiveNoFileComp
}
