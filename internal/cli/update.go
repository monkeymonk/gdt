package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/monkeymonk/gdt/internal/metadata"
	"github.com/spf13/cobra"
)

func newUpdateCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Refresh release metadata cache",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiURL := "https://api.github.com/repos/godotengine/godot/releases"
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
