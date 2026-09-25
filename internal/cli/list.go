package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// newListCmd builds the "gdt list" command, which lists installed versions.
func newListCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List installed versions",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc := app.EngineSvc()
			installed, err := svc.List()
			if err != nil {
				return err
			}
			if len(installed) == 0 {
				fmt.Fprintln(os.Stderr, "No versions installed\n\n  gdt install <version>")
				return nil
			}
			fmt.Println("Installed versions")
			for _, v := range installed {
				marker := "  "
				if v.IsDefault {
					marker = "* "
				}
				fmt.Printf("%s%s\n", marker, v.Version)
			}
			return nil
		},
	}
}
