package cli

import (
	"fmt"
	"os"

	"github.com/monkeymonk/gdt/internal/engine"
	"github.com/monkeymonk/gdt/internal/plugins"
	"github.com/monkeymonk/gdt/internal/project"
	"github.com/spf13/cobra"
)

// newRunCmd builds the "gdt run [version] [-- <args>]" command, which runs a Godot engine version.
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

// newEditCmd builds the "gdt edit [version] [-- <args>]" command, which opens the Godot editor.
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
	svc := app.EngineSvc()

	var version string
	var engineArgs []string

	if len(args) > 0 {
		if resolved, err := svc.ResolveInstalledVersion(args[0]); err == nil {
			version = resolved
			engineArgs = args[1:]
		}
	}

	if version == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("determining working directory: %w", err)
		}
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
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("determining working directory: %w", err)
	}
	projectRoot, err := project.DetectRoot(cwd)
	if err != nil {
		return engine.Actionable(
			fmt.Errorf("detecting project root: %w", err),
			"run gdt run from within a Godot project directory (containing project.godot)",
		)
	}
	hookCtx := plugins.HookContext{
		ProjectRoot:  projectRoot,
		GodotVersion: version,
		EnginePath:   binPath,
	}
	if err := pluginSvc.RunHooks(plugins.BeforeRun, hookCtx); err != nil {
		return engine.Actionable(
			fmt.Errorf("a before_run hook failed: %w", err),
			"check the plugin's hook script for errors, or remove/disable the plugin and retry",
		)
	}

	return engine.ExecBinary(binPath, engineArgs)
}
