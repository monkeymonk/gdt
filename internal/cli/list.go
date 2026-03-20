package cli

import (
	"fmt"
	"os"

	"github.com/monkeymonk/gdt/internal/engine"
	"github.com/spf13/cobra"
)

func newListCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List installed versions",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc := engine.NewService(app.Home, app.Platform, app.Config)
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
