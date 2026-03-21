package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/monkeymonk/gdt/internal/ci"
	"github.com/monkeymonk/gdt/internal/plugins"
	"github.com/monkeymonk/gdt/internal/project"
	"github.com/spf13/cobra"
)

func newCiCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ci",
		Short: "CI integration tools",
	}

	cmd.AddCommand(newCiSetupCmd(app))
	return cmd
}

func newCiSetupCmd(app *App) *cobra.Command {
	var provider string

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Generate CI pipeline configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCiSetup(app, provider)
		},
	}

	cmd.Flags().StringVar(&provider, "provider", "", "CI provider: github, gitlab, generic")
	return cmd
}

func runCiSetup(app *App, provider string) error {
	pluginSvc := plugins.NewService(app.PluginsDir())
	pluginProviders, _ := pluginSvc.DiscoverCIProviders()

	if provider == "" {
		// Build options from built-in providers
		builtinProviders := ci.Providers()
		options := make([]huh.Option[string], 0, len(builtinProviders)+len(pluginProviders))
		for _, p := range builtinProviders {
			options = append(options, huh.NewOption(p.Label, p.Name))
		}
		// Add plugin providers with "plugin:" prefix to distinguish them
		for _, p := range pluginProviders {
			options = append(options, huh.NewOption(p.PluginName+": "+p.Name, "plugin:"+p.PluginName+":"+p.Name))
		}

		err := huh.NewSelect[string]().
			Title("CI Provider").
			Options(options...).
			Value(&provider).
			Run()
		if err != nil {
			return err
		}
	}

	// Handle plugin-provided CI files
	if strings.HasPrefix(provider, "plugin:") {
		parts := strings.SplitN(provider, ":", 3)
		if len(parts) != 3 {
			return fmt.Errorf("invalid plugin provider format: %s", provider)
		}
		pluginName, providerName := parts[1], parts[2]
		// Find the plugin provider
		var found *plugins.PluginCIProvider
		for i := range pluginProviders {
			if pluginProviders[i].PluginName == pluginName && pluginProviders[i].Name == providerName {
				found = &pluginProviders[i]
				break
			}
		}
		if found == nil {
			return fmt.Errorf("plugin CI provider not found: %s:%s", pluginName, providerName)
		}
		content, readErr := os.ReadFile(found.FilePath)
		if readErr != nil {
			return readErr
		}
		// Use provider name as output key
		outPath := filepath.Join(".ci", providerName+filepath.Ext(found.FilePath))
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(outPath, content, 0644); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "CI configuration written to %s\n", outPath)
		cwd, _ := os.Getwd()
		root, _ := project.DetectRoot(cwd)
		_ = pluginSvc.RunHooks(plugins.AfterCISetup, plugins.HookContext{ProjectRoot: root})
		return nil
	}

	content := ci.Generate(provider)
	if content == "" {
		return fmt.Errorf("unknown provider: %s", provider)
	}

	outPath := ci.OutputPath(provider)

	if _, err := os.Stat(outPath); err == nil {
		var confirm bool
		err := huh.NewConfirm().
			Title(fmt.Sprintf("%s already exists. Overwrite?", outPath)).
			Value(&confirm).
			Run()
		if err != nil {
			return err
		}
		if !confirm {
			fmt.Fprintln(os.Stderr, "Aborted")
			return nil
		}
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
		return err
	}

	if provider == "generic" {
		os.Chmod(outPath, 0755)
	}

	fmt.Fprintf(os.Stderr, "CI configuration written to %s\n", outPath)
	cwd, _ := os.Getwd()
	root, _ := project.DetectRoot(cwd)
	_ = pluginSvc.RunHooks(plugins.AfterCISetup, plugins.HookContext{ProjectRoot: root})
	return nil
}
