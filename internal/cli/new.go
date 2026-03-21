package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/monkeymonk/gdt/internal/engine"
	"github.com/monkeymonk/gdt/internal/plugins"
	"github.com/monkeymonk/gdt/internal/project"
	"github.com/spf13/cobra"
)

func newNewCmd(app *App) *cobra.Command {
	var templateURL string
	var version string
	var renderer string
	var csharp bool
	var listTemplates bool

	cmd := &cobra.Command{
		Use:   "new [name]",
		Short: "Create a new Godot project",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			csharpExplicit := cmd.Flags().Changed("csharp")
			rendererExplicit := cmd.Flags().Changed("renderer")
			return runNew(app, listTemplates, name, templateURL, version, renderer, csharp, csharpExplicit, rendererExplicit)
		},
	}

	cmd.Flags().StringVar(&templateURL, "template", "", "Clone from a template repository (GitHub URL or user/repo)")
	cmd.Flags().StringVar(&version, "version", "", "Engine version to pin")
	cmd.Flags().StringVar(&renderer, "renderer", "", "Renderer: forward_plus, mobile, gl_compatibility")
	cmd.Flags().BoolVar(&csharp, "csharp", false, "Create a C# project (uses Mono build)")
	cmd.Flags().BoolVar(&listTemplates, "list-templates", false, "List available templates")

	return cmd
}

func runListTemplates(app *App) error {
	fmt.Println("Built-in:")
	for _, t := range project.AvailableTemplates() {
		fmt.Printf("  %s\n", t)
	}

	svc := plugins.NewService(app.PluginsDir())
	templates, err := svc.DiscoverTemplates()
	if err != nil {
		return err
	}
	if len(templates) > 0 {
		fmt.Println("\nPlugins:")
		for _, t := range templates {
			fmt.Printf("  %s:%s\n", t.PluginName, t.Name)
		}
	}
	return nil
}

func runNew(app *App, listTemplates bool, name string, templateURL string, version string, renderer string, csharp bool, csharpExplicit bool, rendererExplicit bool) error {
	if listTemplates {
		return runListTemplates(app)
	}

	svc := engine.NewService(app.Home, app.Platform, app.Config)
	installed, _ := svc.ListVersionStrings()
	interactive := false

	if name == "" {
		interactive = true
		err := huh.NewInput().
			Title("Project name").
			Value(&name).
			Validate(func(s string) error {
				if s == "" {
					return fmt.Errorf("name is required")
				}
				return nil
			}).
			Run()
		if err != nil {
			return err
		}
	}

	if version == "" {
		interactive = true
		if len(installed) > 0 {
			options := make([]huh.Option[string], len(installed))
			for i, v := range installed {
				options[i] = huh.NewOption(v, v)
			}
			err := huh.NewSelect[string]().
				Title("Engine version").
				Options(options...).
				Value(&version).
				Run()
			if err != nil {
				return err
			}
		} else {
			err := huh.NewInput().
				Title("Engine version").
				Placeholder("4.3").
				Value(&version).
				Run()
			if err != nil {
				return err
			}
		}
	}

	if templateURL == "" && !rendererExplicit {
		interactive = true
		err := huh.NewSelect[string]().
			Title("Renderer").
			Options(
				huh.NewOption("Forward+ (best quality, desktop)", "forward_plus"),
				huh.NewOption("Mobile (balanced)", "mobile"),
				huh.NewOption("Compatibility (widest support, GL)", "gl_compatibility"),
			).
			Value(&renderer).
			Run()
		if err != nil {
			return err
		}
	}

	if templateURL == "" && !csharpExplicit && interactive {
		err := huh.NewConfirm().
			Title("Use C# (.NET)?").
			Value(&csharp).
			Run()
		if err != nil {
			return err
		}
	}

	projectDir := filepath.Join(".", name)

	pluginSvc := plugins.NewService(app.PluginsDir())
	hookCtx := plugins.HookContext{
		ProjectRoot:  filepath.Join(".", name),
		GodotVersion: version,
	}
	if err := pluginSvc.RunHooks(plugins.BeforeNew, hookCtx); err != nil {
		return err
	}

	// Check if template is from a plugin (core built-ins take priority)
	isBuiltin := false
	for _, bt := range project.AvailableTemplates() {
		if templateURL == bt {
			isBuiltin = true
			break
		}
	}
	if templateURL != "" && !isBuiltin && !strings.Contains(templateURL, "/") && !strings.Contains(templateURL, "http") {
		pluginTemplates, _ := pluginSvc.DiscoverTemplates()
		var items []plugins.NamespacedItem
		for _, t := range pluginTemplates {
			items = append(items, plugins.NamespacedItem{
				ShortName:     t.Name,
				QualifiedName: t.PluginName + ":" + t.Name,
				Data:          t,
			})
		}
		if resolved, resolveErr := plugins.ResolveNamespace(templateURL, items); resolveErr == nil {
			pt := resolved.Data.(plugins.PluginTemplate)
			fmt.Fprintf(os.Stderr, "Creating project from plugin template %s:%s...\n", pt.PluginName, pt.Name)
			if err := project.CopyTemplate(pt.Dir, projectDir, name, version); err != nil {
				return err
			}
			templateURL = "" // skip built-in template handling
		}
	}

	if templateURL == "2d" || templateURL == "3d" {
		fmt.Fprintf(os.Stderr, "Creating %s project from built-in template...\n", templateURL)
		if err := project.GenerateFromTemplate(templateURL, projectDir, name, version); err != nil {
			return err
		}
	} else if templateURL != "" {
		fmt.Fprintf(os.Stderr, "Creating project from template...\n")
		if err := project.CloneTemplate(templateURL, projectDir, version); err != nil {
			return err
		}
	} else {
		if renderer == "" {
			renderer = "forward_plus"
		}
		if err := project.Generate(project.ScaffoldOptions{
			Name:     name,
			Version:  version,
			Renderer: renderer,
			Dir:      projectDir,
			CSharp:   csharp,
		}); err != nil {
			return err
		}
	}

	if err := pluginSvc.RunHooks(plugins.AfterNew, hookCtx); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Project %s created\n", name)
	fmt.Fprintf(os.Stderr, "\n  cd %s\n  godot --editor\n", name)
	return nil
}
