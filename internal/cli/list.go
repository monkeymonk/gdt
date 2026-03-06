package cli

import (
	"fmt"
	"os"

	"github.com/monkeymonk/gdt/internal/versions"
	"github.com/spf13/cobra"
)

func newListCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List installed versions",
		RunE: func(cmd *cobra.Command, args []string) error {
			installed, err := versions.List(app.VersionsDir())
			if err != nil {
				return err
			}
			if len(installed) == 0 {
				fmt.Fprintln(os.Stderr, "No versions installed\n\n  gdt install <version>")
				return nil
			}
			fmt.Println("Installed versions\n")
			for _, v := range installed {
				marker := "  "
				if v == app.Config.DefaultVersion {
					marker = "* "
				}
				fmt.Printf("%s%s\n", marker, v)
			}
			return nil
		},
	}
}
