package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/monkeymonk/gdt/internal/engine"
	"github.com/monkeymonk/gdt/internal/plugins"
	"github.com/monkeymonk/gdt/internal/project"
	"github.com/spf13/cobra"
)

var (
	styleBold    = lipgloss.NewStyle().Bold(true)
	stylePrimary = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	styleDim     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

func banner(version string) string {
	return "\n" +
		stylePrimary.Render("  ┌─┐┌┬┐┌┬┐") + "\n" +
		stylePrimary.Render("  │ ┬ ││ │ ") + "\n" +
		stylePrimary.Render("  └─┘─┴┘ ┴ ") + "  " + styleDim.Render("v"+version) + "\n\n" +
		styleBold.Render("  Godot Developer Toolchain") + "\n"
}

func NewRootCmd(app *App) *cobra.Command {
	root := &cobra.Command{
		Use:           "gdt",
		Short:         "Godot Developer Toolchain",
		Long:          banner(app.Version),
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
		newCompletionCmd(),
	)

	// Plugin dispatch for unknown commands
	origHelpFunc := root.HelpFunc()
	root.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			pluginList, _ := plugins.Discover(app.PluginsDir())
			if p, ok := plugins.FindForCommand(pluginList, args[0]); ok {
				dispatchPlugin(app, p, args[1:])
				return
			}
		}
		origHelpFunc(cmd, args)
	})

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
	projectRoot, _ := project.DetectRoot(cwd)

	svc := engine.NewService(app.Home, app.Platform, app.Config)
	resolved, _ := svc.Resolve(cwd)

	cmd.Env = append(os.Environ(), plugins.BuildEnv(plugins.EnvContext{
		Home:         app.Home,
		ProjectRoot:  projectRoot,
		GodotVersion: resolved.Version,
		EnginePath:   resolved.BinaryPath,
	})...)

	cmd.Run()
}

// resolveProjectVersion detects the project root and resolves the engine version.
// Used by lsp, dap, and export commands.
func resolveProjectVersion(app *App) (root string, version string, binPath string, err error) {
	cwd, _ := os.Getwd()
	root, err = project.DetectRoot(cwd)
	if err != nil {
		err = fmt.Errorf("no Godot project found\n\n  Run from a directory containing project.godot")
		return
	}

	svc := engine.NewService(app.Home, app.Platform, app.Config)
	resolved, resolveErr := svc.Resolve(cwd)
	if resolveErr != nil {
		err = resolveErr
		return
	}

	version = resolved.Version
	binPath = resolved.BinaryPath
	return
}
