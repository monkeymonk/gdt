package cli

import (
	"fmt"
	"os"

	"github.com/monkeymonk/gdt/internal/engine"
	"github.com/spf13/cobra"
)

func newLocalCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "local [version]",
		Short: "Pin version for current project",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc := engine.NewService(app.Home, app.Platform, app.Config)

			version := ""
			if len(args) > 0 {
				version = args[0]
			}
			if version == "" && isTTY() {
				v, err := promptVersion(app, "Pin project to version")
				if err != nil {
					return err
				}
				version = v
			}
			if version == "" {
				return fmt.Errorf("version required\n\n  gdt local <version>")
			}
			resolved, err := svc.ResolveInstalledVersion(version)
			if err != nil {
				if !svc.IsInstalled(version) {
					fmt.Fprintf(os.Stderr, "Warning: version %s is not installed\n\n  gdt install %s\n", version, version)
				}
			} else {
				version = resolved
			}
			if err := os.WriteFile(".godot-version", []byte(version+"\n"), 0644); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Project pinned to version %s\n", version)
			return nil
		},
	}
}
