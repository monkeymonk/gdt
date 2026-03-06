package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/monkeymonk/gdt/internal/plugins"
	"github.com/monkeymonk/gdt/internal/project"
	"github.com/monkeymonk/gdt/internal/templates"
	"github.com/monkeymonk/gdt/internal/versions"
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
	issues := 0

	shimPath := filepath.Join(app.ShimsDir(), "godot")
	if _, err := os.Lstat(shimPath); err == nil {
		fmt.Println("  ok  shim installed")
	} else {
		fmt.Println("  FAIL  shim not installed")
		fmt.Printf("        Run: gdt shell init\n")
		issues++
	}

	pathEnv := os.Getenv("PATH")
	if strings.Contains(pathEnv, app.ShimsDir()) {
		fmt.Println("  ok  PATH configured")
	} else {
		fmt.Println("  WARN  shim directory not in PATH")
		fmt.Printf("        Add to your shell profile: eval \"$(gdt shell init)\"\n")
		issues++
	}

	installed, _ := versions.List(app.VersionsDir())
	if len(installed) == 0 {
		fmt.Println("  WARN  no engine versions installed")
		fmt.Printf("        Run: gdt install <version>\n")
		issues++
	} else {
		for _, v := range installed {
			binPath := filepath.Join(app.VersionsDir(), versions.BinaryPath(v, app.Platform.OS))
			if _, err := os.Stat(binPath); err == nil {
				fmt.Printf("  ok  engine %s valid\n", v)
			} else {
				fmt.Printf("  FAIL  engine %s binary missing\n", v)
				issues++
			}
		}
	}

	for _, v := range installed {
		if templates.IsInstalled(app.TemplatesDir(), v) {
			fmt.Printf("  ok  templates for %s\n", v)
		} else {
			fmt.Printf("  WARN  templates missing for %s\n", v)
			fmt.Printf("        Run: gdt templates install %s\n", v)
			issues++
		}
	}

	cwd, _ := os.Getwd()
	if root, err := project.DetectRoot(cwd); err == nil {
		hasCSharp, _ := project.HasCSharp(root)
		if hasCSharp {
			envVer := os.Getenv("GDT_GODOT_VERSION")
			ver, _ := versions.Resolve(cwd, envVer, app.Config.DefaultVersion, installed)
			if ver != "" && !strings.HasSuffix(ver, "-mono") {
				fmt.Println("  WARN  project uses C# but mono engine not installed")
				baseVer := strings.TrimSuffix(ver, "-mono")
				fmt.Printf("        Run: gdt install %s --mono\n", baseVer)
				issues++
			} else if ver != "" {
				fmt.Println("  ok  C# project with mono engine")
			}
		}
	}

	pluginList, _ := plugins.Discover(app.PluginsDir())
	for _, p := range pluginList {
		fmt.Printf("  ok  plugin %s v%s\n", p.Manifest.Name, p.Manifest.Version)
	}

	if issues == 0 {
		fmt.Println("\nAll checks passed")
	} else {
		fmt.Printf("\n%d issue(s) found\n", issues)
	}
	return nil
}
