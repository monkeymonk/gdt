package cli

import (
	"github.com/monkeymonk/gdt/internal/shim"
	"github.com/monkeymonk/gdt/internal/versions"
	"github.com/spf13/cobra"
)

func newRunCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "run <version> [-- <args>]",
		Short: "Run a specific engine version",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]
			binPath, err := versions.AbsoluteBinaryPath(app.VersionsDir(), version, app.Platform.OS)
			if err != nil {
				return err
			}
			engineArgs := args[1:]
			return shim.Exec(binPath, engineArgs)
		},
	}
}
