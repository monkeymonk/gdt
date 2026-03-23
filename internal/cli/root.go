package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/monkeymonk/gdt/internal/engine"
	"github.com/monkeymonk/gdt/internal/plugins"
	"github.com/spf13/cobra"
)

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
	pluginSvc := plugins.NewService(app.PluginsDir())
	if pluginList, err := pluginSvc.Discover(); err == nil {
		for _, p := range pluginList {
			for _, cmdName := range p.Manifest.Commands {
				plug := p // capture loop variable
				root.AddCommand(&cobra.Command{
					Use:                cmdName,
					Short:              plug.Manifest.Description,
					DisableFlagParsing: true,
					RunE: func(cmd *cobra.Command, args []string) error {
						dispatchPlugin(app, plug, args)
						return nil
					},
				})
			}
		}
	}

	return root
}

func dispatchPlugin(app *App, p plugins.Plugin, args []string) {
	binName := p.Manifest.Name
	binPath := filepath.Join(p.Dir, binName)

	cmd := exec.Command(binPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cwd, _ := os.Getwd()
	svc := engine.NewService(app.Home, app.Platform, app.Config)
	projectRoot, rv, _ := svc.ResolveProject(cwd)

	cmd.Env = append(os.Environ(), plugins.BuildEnv(plugins.EnvContext{
		Home:         app.Home,
		ProjectRoot:  projectRoot,
		GodotVersion: rv.Version,
		EnginePath:   rv.BinaryPath,
	})...)

	cmd.Run()
}

// resolveProjectVersion detects the project root and resolves the engine version.
// Used by lsp, dap, and export commands.
func resolveProjectVersion(app *App) (root string, version string, binPath string, err error) {
	cwd, _ := os.Getwd()
	svc := engine.NewService(app.Home, app.Platform, app.Config)
	projectRoot, rv, resolveErr := svc.ResolveProject(cwd)
	if resolveErr != nil {
		return "", "", "", resolveErr
	}
	return projectRoot, rv.Version, rv.BinaryPath, nil
}
