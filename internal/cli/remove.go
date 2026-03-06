package cli

import (
	"fmt"
	"os"

	"github.com/monkeymonk/gdt/internal/versions"
	"github.com/spf13/cobra"
)

func newRemoveCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:     "remove <version>",
		Aliases: []string{"rm", "uninstall"},
		Short:   "Remove an installed version",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]
			if version == app.Config.DefaultVersion {
				fmt.Fprintf(os.Stderr, "Warning: %s is the current global default\n", version)
			}
			if err := versions.Remove(app.VersionsDir(), version); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Version %s removed\n", version)
			return nil
		},
	}
}
