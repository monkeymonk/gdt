package cli

import (
	"os"

	"github.com/monkeymonk/gdt/internal/engine"
	"github.com/monkeymonk/gdt/internal/plugins"
	"github.com/monkeymonk/gdt/internal/project"
	"github.com/spf13/cobra"
)

func newRunCmd(app *App) *cobra.Command {
	var editor bool

	cmd := &cobra.Command{
		Use:   "run [version] [-- <args>]",
		Short: "Run a Godot engine version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGodot(app, args, editor)
		},
	}

	cmd.Flags().BoolVarP(&editor, "editor", "e", false, "Open the editor instead of running the game")

	return cmd
}

func newEditCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "edit [version] [-- <args>]",
		Short: "Open the Godot editor",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGodot(app, args, true)
		},
	}
}

func runGodot(app *App, args []string, editor bool) error {
	svc := engine.NewService(app.Home, app.Platform, app.Config)

	var version string
	var engineArgs []string

	if len(args) > 0 {
		if resolved, err := svc.ResolveInstalledVersion(args[0]); err == nil {
			version = resolved
			engineArgs = args[1:]
		}
	}

	if version == "" {
		cwd, _ := os.Getwd()
		resolved, err := svc.Resolve(cwd)
		if err != nil {
			return err
		}
		version = resolved.Version
		engineArgs = args
	}

	if editor {
		engineArgs = append([]string{"--editor"}, engineArgs...)
	}

	binPath, err := svc.BinaryPath(version)
	if err != nil {
		return err
	}

	pluginSvc := plugins.NewService(app.PluginsDir())
	cwd, _ := os.Getwd()
	projectRoot, _ := project.DetectRoot(cwd)
	hookCtx := plugins.HookContext{
		ProjectRoot:  projectRoot,
		GodotVersion: version,
		EnginePath:   binPath,
	}
	if err := pluginSvc.RunHooks(plugins.BeforeRun, hookCtx); err != nil {
		return err
	}

	return engine.ExecBinary(binPath, engineArgs)
}
