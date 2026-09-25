package cli

import (
	"fmt"
	"os"

	"github.com/monkeymonk/gdt/internal/config"
	"github.com/monkeymonk/gdt/internal/engine"
	"github.com/monkeymonk/gdt/internal/plugins"
	"github.com/spf13/cobra"
)

// newUseCmd builds the "gdt use [version]" command, which sets the global default version.
func newUseCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "use [version]",
		Short: "Set global default version",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc := app.EngineSvc()

			version := ""
			if len(args) > 0 {
				version = args[0]
			}
			if version == "" && isTTY() {
				v, err := promptVersion(app, "Set global default")
				if err != nil {
					return err
				}
				version = v
			}
			if version == "" {
				return fmt.Errorf("version required\n\n  gdt use <version>")
			}
			resolved, err := svc.ResolveInstalledVersion(version)
			if err != nil {
				if !svc.IsInstalled(version) {
					fmt.Fprintf(os.Stderr, "Warning: version %s is not installed\n\n  gdt install %s\n", version, version)
				}
			} else {
				version = resolved
			}
			app.Config.DefaultVersion = version
			if err := config.Save(app.ConfigPath, app.Config); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Default version set to %s\n", version)

			pluginSvc := app.PluginSvc()
			hookCtx := plugins.HookContext{
				GodotVersion: version,
			}
			if err := pluginSvc.RunHooks(plugins.AfterUse, hookCtx); err != nil {
				return engine.Actionable(
					fmt.Errorf("default version set to %s, but the after_use hook failed: %w", version, err),
					"check the plugin's hook script for errors; the default version change is unaffected",
				)
			}
			return nil
		},
	}
}
