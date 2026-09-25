package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/monkeymonk/gdt/internal/plugins"
	"github.com/spf13/cobra"
)

// NewRootCmd builds the "gdt" command, which is the Godot Developer Toolchain.
func NewRootCmd(app *App) *cobra.Command {
	root := &cobra.Command{
		Use:           "gdt",
		Short:         "Godot Developer Toolchain",
		Long:          fmt.Sprintf("Godot Developer Toolchain (v%s)", app.Version),
		Version:       app.Version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.SetVersionTemplate(fmt.Sprintf("gdt %s\n", app.Version))

	root.AddCommand(
		newInstallCmd(app),
		newRemoveCmd(app),
		newListCmd(app),
		newLsRemoteCmd(app),
		newUseCmd(app),
		newLocalCmd(app),
		newRunCmd(app),
		newEditCmd(app),
		newDoctorCmd(app),
		newUpdateCmd(app),
		newShellCmd(app),
		newSelfUpdateCmd(app),
		newTemplatesCmd(app),
		newPluginCmd(app),
		newNewCmd(app),
		newLspCmd(app),
		newDapCmd(app),
		newExportCmd(app),
		newCiCmd(app),
		newCompletionCmd(app),
	)

	// Register plugin commands as cobra subcommands
	pluginSvc := app.PluginSvc()
	pluginList, err := pluginSvc.Discover()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: plugin discovery failed: %v\n", err)
	} else {
		for _, p := range pluginList {
			for _, cmdName := range p.Manifest.Commands {
				plug := p // capture loop variable
				root.AddCommand(&cobra.Command{
					Use:                cmdName,
					Short:              plug.Manifest.Description,
					DisableFlagParsing: true,
					RunE: func(cmd *cobra.Command, args []string) error {
						return dispatchPlugin(app, plug, args)
					},
				})
			}
		}
	}

	return root
}

func dispatchPlugin(app *App, p plugins.Plugin, args []string) error {
	binName := p.Manifest.Name
	binPath := filepath.Join(p.Dir, binName)

	cmd := exec.Command(binPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("dispatching plugin %q: determining working directory: %w", p.Manifest.Name, err)
	}
	svc := app.EngineSvc()
	projectRoot, rv, err := svc.ResolveProject(cwd)
	if err != nil {
		return fmt.Errorf("dispatching plugin %q: %w", p.Manifest.Name, err)
	}

	cmd.Env = append(os.Environ(), plugins.BuildEnv(plugins.EnvContext{
		Home:         app.Home,
		ProjectRoot:  projectRoot,
		GodotVersion: rv.Version,
		EnginePath:   rv.BinaryPath,
	})...)

	cmd.Run()
	return nil
}

// resolveProjectVersion detects the project root and resolves the engine version.
// Used by lsp, dap, and export commands.
func resolveProjectVersion(app *App) (root string, version string, binPath string, err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", "", fmt.Errorf("determining working directory: %w", err)
	}
	svc := app.EngineSvc()
	projectRoot, rv, resolveErr := svc.ResolveProject(cwd)
	if resolveErr != nil {
		return "", "", "", resolveErr
	}
	return projectRoot, rv.Version, rv.BinaryPath, nil
}
