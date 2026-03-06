package cli

import (
	"os"

	"github.com/monkeymonk/gdt/internal/shim"
	"github.com/monkeymonk/gdt/internal/versions"
	"github.com/spf13/cobra"
)

func newRunCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "run [version] [-- <args>]",
		Short: "Run a Godot engine version",
		RunE: func(cmd *cobra.Command, args []string) error {
			installed, _ := versions.List(app.VersionsDir())

			var version string
			var engineArgs []string

			if len(args) > 0 {
				// Try to resolve as installed version or alias
				version = resolveInstalledVersion(args[0], installed)
				engineArgs = args[1:]
			}

			if version == "" {
				// Fall back to standard resolution chain
				cwd, _ := os.Getwd()
				v, err := versions.Resolve(cwd, os.Getenv("GDT_GODOT_VERSION"), app.Config.DefaultVersion, installed)
				if err != nil {
					return err
				}
				version = v
				engineArgs = args
			}

			binPath, err := versions.AbsoluteBinaryPath(app.VersionsDir(), version, app.Platform.OS)
			if err != nil {
				return err
			}
			return shim.Exec(binPath, engineArgs)
		},
	}
}

func resolveInstalledVersion(query string, installed []string) string {
	// Exact match
	for _, v := range installed {
		if v == query {
			return v
		}
	}

	// "latest" → last installed (sorted)
	if query == "latest" || query == "stable" {
		if len(installed) > 0 {
			return installed[len(installed)-1]
		}
		return ""
	}

	// Prefix match (4.3 → 4.3-mono if 4.3 not found, or 4 → 4.3)
	for i := len(installed) - 1; i >= 0; i-- {
		if len(installed[i]) >= len(query) && installed[i][:len(query)] == query {
			return installed[i]
		}
	}

	return ""
}
