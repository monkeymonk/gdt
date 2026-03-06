package cli

import (
	"fmt"
	"os"

	"github.com/monkeymonk/gdt/internal/config"
	"github.com/monkeymonk/gdt/internal/versions"
	"github.com/spf13/cobra"
)

func newUseCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "use <version>",
		Short: "Set global default version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]
			if !versions.IsInstalled(app.VersionsDir(), version) {
				fmt.Fprintf(os.Stderr, "Warning: version %s is not installed\n\n  gdt install %s\n", version, version)
			}
			app.Config.DefaultVersion = version
			if err := config.Save(app.ConfigPath, app.Config); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Default version set to %s\n", version)
			return nil
		},
	}
}
