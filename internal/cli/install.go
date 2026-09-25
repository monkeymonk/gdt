package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/monkeymonk/gdt/internal/engine"
	"github.com/monkeymonk/gdt/internal/metadata"
	"github.com/monkeymonk/gdt/internal/plugins"
	"github.com/spf13/cobra"
)

// newInstallCmd builds the "gdt install [version]" command, which installs a Godot engine version.
func newInstallCmd(app *App) *cobra.Command {
	var mono bool
	var force bool
	var refresh bool

	cmd := &cobra.Command{
		Use:   "install [version]",
		Short: "Install a Godot engine version",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc := app.EngineSvc()

			query := ""
			if len(args) > 0 {
				query = args[0]
			}
			if query == "" {
				// Try .godot-version file first
				if cwd, err := os.Getwd(); err == nil {
					if resolved, err := svc.Resolve(cwd); err == nil && resolved.Version != "" {
						query = resolved.Version
					}
				}
			}
			if query == "" && isTTY() {
				releases, err := metadata.EnsureCache(svc.CachePath(), app.Config.GodotAPIURL(), os.Getenv("GITHUB_TOKEN"), refresh)
				if err != nil {
					return err
				}
				selected, err := promptRemoteVersion(releases)
				if err != nil {
					return err
				}
				query = selected
			}
			if query == "" {
				return fmt.Errorf("version required\n\n  gdt install <version>\n  gdt ls-remote")
			}

			fmt.Fprintf(os.Stderr, "Installing Godot %s...\n", query)
			result, err := svc.Install(cmd.Context(), query, engine.InstallOpts{
				Mono:    mono,
				Force:   force,
				Refresh: refresh,
			})
			if errors.Is(err, engine.ErrAlreadyInstalled) {
				fmt.Fprintf(os.Stderr, "Version %s is already installed (use --force to reinstall)\n", result.VersionName)
				return nil
			}
			if err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Godot %s installed\n", result.VersionName)
			fmt.Fprintf(os.Stderr, "\n  Hint: install export templates with: gdt templates install %s\n", result.Version)

			pluginSvc := plugins.NewService(app.PluginsDir())
			enginePath, _ := svc.BinaryPath(result.VersionName)
			hookCtx := plugins.HookContext{
				GodotVersion: result.VersionName,
				EnginePath:   enginePath,
			}
			if err := pluginSvc.RunHooks(plugins.AfterInstall, hookCtx); err != nil {
				return engine.Actionable(
					fmt.Errorf("godot %s installed successfully, but the after_install hook failed: %w", result.VersionName, err),
					"check the plugin's hook script for errors; the installed engine is unaffected",
				)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&mono, "mono", false, "Install Mono/C# build")
	cmd.Flags().BoolVar(&force, "force", false, "Force reinstall")
	cmd.Flags().BoolVar(&refresh, "refresh", false, "Refresh metadata cache before resolving")

	return cmd
}
