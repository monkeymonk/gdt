package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/monkeymonk/gdt/internal/plugins"
	"github.com/spf13/cobra"
)

func newCompletionCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion script",
		Long: `Generate shell completion script for gdt.

  bash:       gdt completion bash > /etc/bash_completion.d/gdt
  zsh:        gdt completion zsh > "${fpath[1]}/_gdt"
  fish:       gdt completion fish > ~/.config/fish/completions/gdt.fish
  powershell: gdt completion powershell | Out-String | Invoke-Expression`,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			shell := args[0]
			var err error
			switch shell {
			case "bash":
				err = cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				err = cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				err = cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				err = cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
			if err != nil {
				return err
			}

			// Append plugin completions
			svc := plugins.NewService(app.PluginsDir())
			for _, p := range svc.DiscoverCompletionPlugins() {
				binPath := filepath.Join(p.Dir, p.Manifest.Name)
				out, runErr := plugins.RunPluginSubcommand(
					binPath, p.Dir, nil, plugins.DefaultHookTimeout, "completions", shell)
				if runErr != nil {
					fmt.Fprintf(os.Stderr, "warning: plugin %s completion failed: %s\n", p.Manifest.Name, runErr)
					continue
				}
				fmt.Fprint(os.Stdout, out)
			}
			return nil
		},
	}
	return cmd
}
