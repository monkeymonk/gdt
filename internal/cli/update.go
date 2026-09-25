package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/monkeymonk/gdt/internal/metadata"
	"github.com/spf13/cobra"
)

// newUpdateCmd builds the "gdt update" command, which refreshes the release metadata cache.
func newUpdateCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Refresh release metadata cache",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiURL := app.Config.GodotAPIURL()
			token := os.Getenv("GITHUB_TOKEN")
			fmt.Fprintln(os.Stderr, "Refreshing release metadata...")

			releases, err := metadata.FetchReleases(apiURL, token)
			if err != nil {
				return err
			}

			cache := &metadata.Cache{
				UpdatedAt: time.Now(),
				Releases:  releases,
			}
			if err := metadata.SaveCache(app.EngineSvc().CachePath(), cache); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Metadata updated (%d releases)\n", len(releases))
			return nil
		},
	}
}
