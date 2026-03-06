package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/monkeymonk/gdt/internal/download"
	"github.com/monkeymonk/gdt/internal/metadata"
	"github.com/monkeymonk/gdt/internal/templates"
	"github.com/spf13/cobra"
)

func newTemplatesCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "templates",
		Short: "Manage export templates",
	}

	installCmd := newTemplatesInstallCmd(app)
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List installed templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			list, err := templates.List(app.TemplatesDir())
			if err != nil {
				return err
			}
			if len(list) == 0 {
				fmt.Fprintln(os.Stderr, "No templates installed\n\n  gdt templates install <version>")
				return nil
			}
			fmt.Println("Installed templates")
			for _, t := range list {
				fmt.Printf("  %s\n", t)
			}
			return nil
		},
	}

	cmd.AddCommand(installCmd, listCmd)
	return cmd
}

func newTemplatesInstallCmd(app *App) *cobra.Command {
	var mono bool
	var refresh bool

	cmd := &cobra.Command{
		Use:   "install <version>",
		Short: "Install export templates",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTemplatesInstall(app, args[0], mono, refresh)
		},
	}

	cmd.Flags().BoolVar(&mono, "mono", false, "Install Mono templates")
	cmd.Flags().BoolVar(&refresh, "refresh", false, "Refresh metadata cache")
	return cmd
}

func runTemplatesInstall(app *App, query string, mono bool, refresh bool) error {
	releases, err := loadMetadata(app, refresh)
	if err != nil {
		return err
	}

	release, err := metadata.ResolveVersion(releases, query)
	if err != nil {
		return err
	}

	versionName := release.Version
	if mono {
		versionName += "-mono"
	}

	artifactName := templates.ArtifactName(release.Version, mono)
	downloadURL, ok := release.Assets[artifactName]
	if !ok {
		for name, url := range release.Assets {
			if strings.Contains(name, "export_templates") {
				if mono && strings.Contains(name, "mono") {
					artifactName = name
					downloadURL = url
					ok = true
					break
				} else if !mono && !strings.Contains(name, "mono") {
					artifactName = name
					downloadURL = url
					ok = true
					break
				}
			}
		}
		if !ok {
			return fmt.Errorf("templates not found for version %s", release.Version)
		}
	}

	downloadDir := filepath.Join(app.CacheDir(), "downloads")
	archivePath := filepath.Join(downloadDir, artifactName)
	fmt.Fprintf(os.Stderr, "Installing templates for %s...\n", versionName)
	if err := download.File(downloadURL, archivePath); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	destDir := filepath.Join(app.TemplatesDir(), versionName)
	os.MkdirAll(filepath.Dir(destDir), 0755)
	os.RemoveAll(destDir)
	if err := download.ExtractZip(archivePath, destDir); err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Templates for %s installed\n", versionName)
	return nil
}
