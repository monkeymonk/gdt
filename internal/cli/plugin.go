package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/monkeymonk/gdt/internal/plugins"
	"github.com/spf13/cobra"
)

func newPluginCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage plugins",
	}

	cmd.AddCommand(
		newPluginInstallCmd(app),
		newPluginListCmd(app),
		newPluginRemoveCmd(app),
	)

	return cmd
}

func newPluginInstallCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "install <repository>",
		Short: "Install a plugin from a Git repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPluginInstall(app, args[0])
		},
	}
}

func runPluginInstall(app *App, repo string) error {
	repoURL := repo
	if !strings.HasPrefix(repo, "http") {
		repoURL = "https://github.com/" + repo
	}

	parts := strings.Split(strings.TrimSuffix(repo, "/"), "/")
	name := parts[len(parts)-1]

	destDir := filepath.Join(app.PluginsDir(), name)
	if _, err := os.Stat(destDir); err == nil {
		return fmt.Errorf("plugin %s is already installed\n\n  gdt plugin remove %s", name, name)
	}

	fmt.Fprintf(os.Stderr, "Installing plugin %s...\n", name)

	gitCmd := exec.Command("git", "clone", "--depth", "1", repoURL, destDir)
	gitCmd.Stdout = os.Stderr
	gitCmd.Stderr = os.Stderr
	if err := gitCmd.Run(); err != nil {
		return fmt.Errorf("failed to clone %s: %w", repoURL, err)
	}

	manifestPath := filepath.Join(destDir, plugins.ManifestFile)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		os.RemoveAll(destDir)
		return fmt.Errorf("plugin missing %s manifest", plugins.ManifestFile)
	}

	m, err := plugins.ParseManifest(data)
	if err != nil {
		os.RemoveAll(destDir)
		return fmt.Errorf("invalid plugin manifest: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Plugin %s v%s installed\n", m.Name, m.Version)
	return nil
}

func newPluginListCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed plugins",
		RunE: func(cmd *cobra.Command, args []string) error {
			pluginList, err := plugins.Discover(app.PluginsDir())
			if err != nil {
				return err
			}
			if len(pluginList) == 0 {
				fmt.Fprintln(os.Stderr, "No plugins installed")
				return nil
			}
			fmt.Println("Installed plugins\n")
			for _, p := range pluginList {
				fmt.Printf("  %s v%s\n", p.Manifest.Name, p.Manifest.Version)
			}
			return nil
		},
	}
}

func newPluginRemoveCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			dir := filepath.Join(app.PluginsDir(), name)
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				return fmt.Errorf("plugin %s not found", name)
			}
			if err := os.RemoveAll(dir); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Plugin %s removed\n", name)
			return nil
		},
	}
}
