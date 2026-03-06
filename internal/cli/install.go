package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/monkeymonk/gdt/internal/download"
	"github.com/monkeymonk/gdt/internal/metadata"
	"github.com/monkeymonk/gdt/internal/versions"
	"github.com/spf13/cobra"
)

func newInstallCmd(app *App) *cobra.Command {
	var mono bool
	var force bool
	var refresh bool

	cmd := &cobra.Command{
		Use:   "install <version>",
		Short: "Install a Godot engine version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInstall(app, args[0], mono, force, refresh)
		},
	}

	cmd.Flags().BoolVar(&mono, "mono", false, "Install Mono/C# build")
	cmd.Flags().BoolVar(&force, "force", false, "Force reinstall")
	cmd.Flags().BoolVar(&refresh, "refresh", false, "Refresh metadata cache before resolving")

	return cmd
}

func runInstall(app *App, query string, mono bool, force bool, refresh bool) error {
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

	if !force && versions.IsInstalled(app.VersionsDir(), versionName) {
		fmt.Fprintf(os.Stderr, "Version %s is already installed (use --force to reinstall)\n", versionName)
		return nil
	}

	artifactName := resolveEngineArtifact(release, app.Platform.ArtifactName(), mono)
	downloadURL, ok := release.Assets[artifactName]
	if !ok {
		return fmt.Errorf("artifact %q not found for version %s", artifactName, release.Version)
	}

	downloadDir := filepath.Join(app.CacheDir(), "downloads")
	archivePath := filepath.Join(downloadDir, artifactName)
	fmt.Fprintf(os.Stderr, "Installing Godot %s...\n", versionName)
	if err := download.File(downloadURL, archivePath); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	if checksumURL, ok := release.Assets["SHA512-SUMS.txt"]; ok {
		checksumPath := filepath.Join(downloadDir, "SHA512-SUMS.txt")
		if err := download.File(checksumURL, checksumPath); err == nil {
			if checksum := findChecksum(checksumPath, artifactName); checksum != "" {
				if err := download.VerifyChecksum(archivePath, checksum); err != nil {
					os.Remove(archivePath)
					return fmt.Errorf("checksum verification failed: %w", err)
				}
				fmt.Fprintln(os.Stderr, "  checksum verified")
			}
		}
	}

	tmpDir := filepath.Join(app.CacheDir(), "tmp")
	os.MkdirAll(tmpDir, 0755)
	if err := download.ExtractZip(archivePath, tmpDir); err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	destDir := filepath.Join(app.VersionsDir(), versionName)
	os.MkdirAll(filepath.Dir(destDir), 0755)
	os.RemoveAll(destDir)
	if err := os.Rename(tmpDir, destDir); err != nil {
		return fmt.Errorf("failed to install: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Godot %s installed\n", versionName)
	fmt.Fprintf(os.Stderr, "\n  Hint: install export templates with: gdt templates install %s\n", release.Version)
	return nil
}
