package cli

import (
	"github.com/spf13/cobra"
)

func newDapCmd(app *App) *cobra.Command {
	var port int
	var projectPath string

	cmd := &cobra.Command{
		Use:   "dap",
		Short: "Start DAP proxy (stdin/stdout to Godot debugger)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLanguageProxy(app, port, "--dap-port", projectPath)
		},
	}

	cmd.Flags().IntVar(&port, "port", 6006, "Godot DAP TCP port")
	cmd.Flags().StringVarP(&projectPath, "path", "C", "", "Path to Godot project directory")
	return cmd
}
