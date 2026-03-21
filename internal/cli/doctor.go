package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/monkeymonk/gdt/internal/engine"
	"github.com/monkeymonk/gdt/internal/plugins"
	"github.com/monkeymonk/gdt/internal/project"
	"github.com/spf13/cobra"
)

func newDoctorCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose installation problems",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDoctor(app)
		},
	}
}

func runDoctor(app *App) error {
	svc := engine.NewService(app.Home, app.Platform, app.Config)
	issues := 0

	binDir := filepath.Dir(os.Args[0])
	if binDir == "." {
		if exe, err := os.Executable(); err == nil {
			binDir = filepath.Dir(exe)
		}
	}
	pathEnv := os.Getenv("PATH")
	if strings.Contains(pathEnv, binDir) {
		fmt.Println("  ok  PATH configured")
	} else {
		fmt.Println("  WARN  gdt not in PATH")
		fmt.Printf("        Add to your shell profile: eval \"$(gdt shell init)\"\n")
		issues++
	}

	installed, _ := svc.List()
	if len(installed) == 0 {
		fmt.Println("  WARN  no engine versions installed")
		fmt.Printf("        Run: gdt install <version>\n")
		issues++
	} else {
		for _, v := range installed {
			if _, err := svc.BinaryPath(v.Version); err == nil {
				fmt.Printf("  ok  engine %s valid\n", v.Version)
			} else {
				fmt.Printf("  FAIL  engine %s binary missing\n", v.Version)
				issues++
			}
		}
	}

	for _, v := range installed {
		if svc.TemplatesInstalled(v.Version) {
			fmt.Printf("  ok  templates for %s\n", v.Version)
		} else {
			fmt.Printf("  WARN  templates missing for %s\n", v.Version)
			fmt.Printf("        Run: gdt templates install %s\n", v.Version)
			issues++
		}
	}

	cwd, _ := os.Getwd()
	if root, err := project.DetectRoot(cwd); err == nil {
		resolved, _ := svc.Resolve(cwd)
		hasCSharp, _ := project.HasCSharp(root)
		if hasCSharp {
			ver := resolved.Version
			if ver != "" && !strings.HasSuffix(ver, "-mono") {
				fmt.Println("  WARN  project uses C# but mono engine not installed")
				baseVer := strings.TrimSuffix(ver, "-mono")
				fmt.Printf("        Run: gdt install %s --mono\n", baseVer)
				issues++
			} else if ver != "" {
				fmt.Println("  ok  C# project with mono engine")
			}
		}

		presets, presetsErr := project.ParsePresets(root)
		if presetsErr == nil && len(presets) > 0 {
			if !svc.TemplatesInstalled(resolved.Version) {
				issues++
				fmt.Printf("  [!] export presets found but no templates installed for %s\n", resolved.Version)
				fmt.Printf("      fix: gdt templates install %s\n", resolved.Version)
			}
		}
	}

	pluginList, _ := plugins.Discover(app.PluginsDir())
	for _, p := range pluginList {
		fmt.Printf("  ok  plugin %s v%s\n", p.Manifest.Name, p.Manifest.Version)
	}

	// Run plugin doctor checks (V2 protocol)
	pluginSvc := plugins.NewService(app.PluginsDir())
	doctorPlugins := pluginSvc.DiscoverDoctorPlugins()
	for _, p := range doctorPlugins {
		binPath := filepath.Join(p.Dir, p.Manifest.Name)
		if _, err := os.Stat(binPath); os.IsNotExist(err) {
			fmt.Printf("  WARN  [%s] doctor declared but binary missing\n", p.Manifest.Name)
			issues++
			continue
		}

		env := append(os.Environ(), plugins.BuildEnv(plugins.EnvContext{
			Home: app.Home,
		})...)
		out, runErr := plugins.RunPluginSubcommand(binPath, p.Dir, env, plugins.DefaultHookTimeout, "doctor", "check")
		if runErr != nil {
			fmt.Printf("  FAIL  [%s] doctor check failed: %s\n", p.Manifest.Name, runErr)
			issues++
			continue
		}

		for _, r := range plugins.ParseStatusLines(out) {
			switch r.Status {
			case "OK":
				fmt.Printf("  ok    [%s] %s\n", p.Manifest.Name, r.Message)
			case "WARN":
				fmt.Printf("  WARN  [%s] %s\n", p.Manifest.Name, r.Message)
				issues++
			case "FAIL":
				fmt.Printf("  FAIL  [%s] %s\n", p.Manifest.Name, r.Message)
				issues++
			}
		}
	}

	if issues == 0 {
		fmt.Println("\nAll checks passed")
	} else {
		fmt.Printf("\n%d issue(s) found\n", issues)
	}
	return nil
}
