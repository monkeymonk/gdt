package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/monkeymonk/gdt/internal/engine"
	"github.com/monkeymonk/gdt/internal/project"
	"github.com/spf13/cobra"
)

func newExportCmd(app *App) *cobra.Command {
	var outputDir string
	var debug bool
	var verbose bool
	var listPresets bool

	cmd := &cobra.Command{
		Use:   "export [preset]",
		Short: "Export project for a platform",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if listPresets {
				return runExportList()
			}
			preset := ""
			if len(args) > 0 {
				preset = args[0]
			}
			if preset == "" && isTTY() {
				cwd, _ := os.Getwd()
				root, err := project.DetectRoot(cwd)
				if err != nil {
					return err
				}
				presets, err := project.ParsePresets(root)
				if err != nil {
					return err
				}
				if len(presets) == 0 {
					return fmt.Errorf("no export presets found\n\n  Configure them in the Godot editor: Project > Export")
				}
				p, err := promptPreset(presets)
				if err != nil {
					return err
				}
				preset = p
			}
			if preset == "" {
				return fmt.Errorf("preset name required\n\n  List available: gdt export --list")
			}
			return runExport(app, preset, outputDir, debug, verbose)
		},
	}

	cmd.Flags().StringVar(&outputDir, "output", "", "Output directory (default: dist/<preset>)")
	cmd.Flags().BoolVar(&debug, "debug", false, "Use debug export instead of release")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show Godot engine output")
	cmd.Flags().BoolVar(&listPresets, "list", false, "List available export presets")

	return cmd
}

func runExportList() error {
	cwd, _ := os.Getwd()
	root, err := project.DetectRoot(cwd)
	if err != nil {
		return err
	}

	presets, err := project.ParsePresets(root)
	if err != nil {
		return err
	}

	fmt.Println("Available export presets")
	for _, p := range presets {
		fmt.Printf("  %s\n", p)
	}
	return nil
}

func runExport(app *App, preset string, outputDir string, debug bool, verbose bool) error {
	root, version, binPath, err := resolveProjectVersion(app)
	if err != nil {
		return err
	}

	presets, err := project.ParsePresets(root)
	if err != nil {
		return err
	}

	found := false
	for _, p := range presets {
		if strings.EqualFold(p, preset) {
			preset = p
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("preset %q not found\n\n  Available: %s", preset, strings.Join(presets, ", "))
	}

	svc := engine.NewService(app.Home, app.Platform, app.Config)
	if !svc.TemplatesInstalled(version) {
		fmt.Fprintf(os.Stderr, "Export templates not installed for %s\n", version)
		fmt.Fprintf(os.Stderr, "\n  gdt templates install %s\n", version)
		return fmt.Errorf("missing export templates")
	}

	if outputDir == "" {
		outputDir = project.DefaultOutputDir(preset)
	}
	os.MkdirAll(outputDir, 0755)

	outputFile := filepath.Join(outputDir, "game")

	exportFlag := "--export-release"
	if debug {
		exportFlag = "--export-debug"
	}

	fmt.Fprintf(os.Stderr, "Exporting %q...\n", preset)
	godotCmd := exec.Command(binPath, "--headless", exportFlag, preset, outputFile)
	godotCmd.Dir = root

	if verbose {
		godotCmd.Stdout = os.Stderr
		godotCmd.Stderr = os.Stderr

		if err := godotCmd.Run(); err != nil {
			return fmt.Errorf("export failed: %w", err)
		}
	} else {
		var stdout, stderr bytes.Buffer
		godotCmd.Stdout = &stdout
		godotCmd.Stderr = &stderr

		if err := godotCmd.Run(); err != nil {
			combined := strings.TrimSpace(stderr.String() + stdout.String())
			if combined != "" {
				return fmt.Errorf("export failed: %w\n\n%s", err, combined)
			}
			return fmt.Errorf("export failed: %w", err)
		}
	}

	fmt.Fprintf(os.Stderr, "Export complete: %s\n", outputDir)
	return nil
}
