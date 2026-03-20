package cli

import (
	"fmt"
	"os"

	"github.com/monkeymonk/gdt/internal/engine"
	"github.com/spf13/cobra"
)

func newRemoveCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:     "remove [version]",
		Aliases: []string{"rm", "uninstall"},
		Short:   "Remove an installed version",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := ""
			if len(args) > 0 {
				version = args[0]
			}
			if version == "" && isTTY() {
				v, err := promptVersion(app, "Version to remove")
				if err != nil {
					return err
				}
				version = v
			}
			if version == "" {
				return fmt.Errorf("version required\n\n  gdt remove <version>")
			}
			if version == app.Config.DefaultVersion {
				fmt.Fprintf(os.Stderr, "Warning: %s is the current global default\n", version)
			}
			if isTTY() {
				ok, err := promptConfirm(fmt.Sprintf("Remove Godot %s?", version))
				if err != nil {
					return err
				}
				if !ok {
					fmt.Fprintln(os.Stderr, "Aborted")
					return nil
				}
			}

			svc := engine.NewService(app.Home, app.Platform, app.Config)
			if err := svc.Remove(cmd.Context(), version); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Version %s removed\n", version)
			return nil
		},
	}
}
