package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/monkeymonk/gdt/internal/plugins"
	"github.com/spf13/cobra"
)

func newPluginCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage plugins",
	}

	cmd.AddCommand(
		newPluginInstallCmd(app),
		newPluginListCmd(app),
		newPluginUpdateCmd(app),
		newPluginRemoveCmd(app),
		newPluginNewCmd(app),
	)

	return cmd
}

func newPluginInstallCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "install [repository]",
		Short: "Install a plugin from a Git repository",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo := ""
			if len(args) > 0 {
				repo = args[0]
			}
			if repo == "" && isTTY() {
				v, err := promptInput("Plugin repository", "user/repo or https://...")
				if err != nil {
					return err
				}
				repo = v
			}
			if repo == "" {
				return fmt.Errorf("repository required\n\n  gdt plugin install <repository>")
			}
			svc := plugins.NewService(app.PluginsDir())
			fmt.Fprintf(os.Stderr, "Installing plugin...\n")
			m, err := svc.Install(repo)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Plugin %s v%s installed\n", m.Name, m.Version)
			return nil
		},
	}
}

func newPluginListCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed plugins",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc := plugins.NewService(app.PluginsDir())
			pluginList, err := svc.Discover()
			if err != nil {
				return err
			}
			if len(pluginList) == 0 {
				fmt.Fprintln(os.Stderr, "No plugins installed")
				return nil
			}
			fmt.Println("Installed plugins")
			for _, p := range pluginList {
				line := fmt.Sprintf("  %s v%s", p.Manifest.Name, p.Manifest.Version)
				if p.Manifest.HasContributions() {
					var contribs []string
					if len(p.Manifest.Contributions.Templates) > 0 {
						contribs = append(contribs, fmt.Sprintf("%d templates", len(p.Manifest.Contributions.Templates)))
					}
					if len(p.Manifest.Contributions.Presets) > 0 {
						contribs = append(contribs, fmt.Sprintf("%d presets", len(p.Manifest.Contributions.Presets)))
					}
					if p.Manifest.Contributions.Doctor {
						contribs = append(contribs, "doctor")
					}
					if len(contribs) > 0 {
						line += " (" + strings.Join(contribs, ", ") + ")"
					}
				}
				fmt.Println(line)
			}
			return nil
		},
	}
}

func newPluginUpdateCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "update [name]",
		Short: "Update plugins (all or by name)",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc := plugins.NewService(app.PluginsDir())
			pluginList, err := svc.Discover()
			if err != nil {
				return err
			}
			if len(pluginList) == 0 {
				fmt.Fprintln(os.Stderr, "No plugins installed")
				return nil
			}

			name := ""
			if len(args) > 0 {
				name = args[0]
			}

			results, err := svc.Update(name)
			if err != nil {
				return err
			}
			for _, r := range results {
				if r.Err != nil {
					fmt.Fprintf(os.Stderr, "  failed to update %s: %s\n", r.Name, r.Err)
				} else {
					fmt.Fprintf(os.Stderr, "  %s updated\n", r.Name)
				}
			}
			return nil
		},
	}
}

func newPluginNewCmd(app *App) *cobra.Command {
	var lang string

	cmd := &cobra.Command{
		Use:   "new [name]",
		Short: "Scaffold a new plugin",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			if name == "" && isTTY() {
				v, err := promptInput("Plugin name", "gdt-my-plugin")
				if err != nil {
					return err
				}
				name = v
			}
			if name == "" {
				return fmt.Errorf("name required\n\n  gdt plugin new <name>")
			}
			svc := plugins.NewService(app.PluginsDir())
			dir, err := svc.ScaffoldV2(plugins.ScaffoldOptions{Name: name, Lang: lang})
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Plugin scaffolded at %s\n", dir)
			fmt.Fprintf(os.Stderr, "\n  Edit %s to configure your plugin\n", dir+"/"+plugins.ManifestFile)
			return nil
		},
	}

	cmd.Flags().StringVar(&lang, "lang", "shell", "Scaffold language: shell, go, rust, python")
	return cmd
}

func newPluginRemoveCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "remove [name]",
		Short: "Remove a plugin",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			if name == "" && isTTY() {
				v, err := promptInstalledPlugin(app, "Plugin to remove")
				if err != nil {
					return err
				}
				name = v
			}
			if name == "" {
				return fmt.Errorf("name required\n\n  gdt plugin remove <name>")
			}
			if isTTY() {
				ok, err := promptConfirm(fmt.Sprintf("Remove plugin %s?", name))
				if err != nil {
					return err
				}
				if !ok {
					fmt.Fprintln(os.Stderr, "Aborted")
					return nil
				}
			}
			svc := plugins.NewService(app.PluginsDir())
			if err := svc.Remove(name); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Plugin %s removed\n", name)
			return nil
		},
	}
}
