package cli

import (
	"fmt"
	"os"

	"github.com/monkeymonk/gdt/internal/versions"
	"github.com/spf13/cobra"
)

func newLocalCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "local <version>",
		Short: "Pin version for current project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]
			if !versions.IsInstalled(app.VersionsDir(), version) {
				fmt.Fprintf(os.Stderr, "Warning: version %s is not installed\n\n  gdt install %s\n", version, version)
			}
			if err := os.WriteFile(".godot-version", []byte(version+"\n"), 0644); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Project pinned to version %s\n", version)
			return nil
		},
	}
}
