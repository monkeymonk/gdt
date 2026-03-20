package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/monkeymonk/gdt/internal/ci"
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
			return runCiSetup(provider)
		},
	}

	cmd.Flags().StringVar(&provider, "provider", "", "CI provider: github, gitlab, generic")
	return cmd
}

func runCiSetup(provider string) error {
	if provider == "" {
		providers := ci.Providers()
		options := make([]huh.Option[string], len(providers))
		for i, p := range providers {
			options[i] = huh.NewOption(p.Label, p.Name)
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
	return nil
}
