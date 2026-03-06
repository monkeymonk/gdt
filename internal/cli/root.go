package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/monkeymonk/gdt/internal/plugins"
	"github.com/monkeymonk/gdt/internal/project"
	"github.com/monkeymonk/gdt/internal/versions"
	"github.com/spf13/cobra"
)

func NewRootCmd(app *App) *cobra.Command {
	root := &cobra.Command{
		Use:           "gdt",
		Short:         "Godot Developer Toolchain",
		Long:          "A cross-platform CLI to manage Godot Engine installations.",
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
		newDoctorCmd(app),
		newUpdateCmd(app),
		newShellCmd(app),
		newSelfUpdateCmd(app),
		newTemplatesCmd(app),
		newPluginCmd(app),
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
	installed, _ := versions.List(app.VersionsDir())
	ver, _ := versions.Resolve(cwd, os.Getenv("GDT_GODOT_VERSION"), app.Config.DefaultVersion, installed)
	enginePath := ""
	if ver != "" {
		enginePath, _ = versions.AbsoluteBinaryPath(app.VersionsDir(), ver, app.Platform.OS)
	}

	cmd.Env = append(os.Environ(),
		"GDT_HOME="+app.Home,
		"GDT_PROJECT_ROOT="+projectRoot,
		"GDT_GODOT_VERSION="+ver,
		"GDT_ENGINE_PATH="+enginePath,
	)

	cmd.Run()
}
