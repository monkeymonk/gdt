package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monkeymonk/gdt/internal/download"
	"github.com/monkeymonk/gdt/internal/metadata"
)

// ListTemplates returns the names of all installed template sets, sorted alphabetically.
func (s *Service) ListTemplates() ([]string, error) {
	entries, err := os.ReadDir(s.templatesDir())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

// TemplatesInstalled checks whether templates for the given version are installed.
func (s *Service) TemplatesInstalled(version string) bool {
	_, err := os.Stat(filepath.Join(s.templatesDir(), version))
	return err == nil
}

// InstallTemplates downloads and extracts export templates for the given version.
func (s *Service) InstallTemplates(ctx context.Context, query string, opts InstallOpts) (*InstallResult, error) {
	apiURL := "https://api.github.com/repos/godotengine/godot/releases"
	token := os.Getenv("GDT_GITHUB_TOKEN")

	releases, err := metadata.EnsureCache(s.cachePath(), apiURL, token, opts.Refresh)
	if err != nil {
		return nil, err
	}

	release, err := metadata.ResolveVersion(releases, query)
	if err != nil {
		return nil, err
	}

	versionName := release.Version
	if opts.Mono {
		versionName += "-mono"
	}

	// Check if already installed
	if s.TemplatesInstalled(versionName) && !opts.Force {
		return &InstallResult{
			Version:     release.Version,
			VersionName: versionName,
			IsNew:       false,
		}, ErrAlreadyInstalled
	}

	artifactName := metadata.TemplateArtifactName(release.Version, opts.Mono)
	downloadURL, ok := release.Assets[artifactName]
	if !ok {
		// Fallback: scan assets for export_templates
		for name, url := range release.Assets {
			if strings.Contains(name, "export_templates") {
				if opts.Mono && strings.Contains(name, "mono") {
					artifactName = name
					downloadURL = url
					ok = true
					break
				} else if !opts.Mono && !strings.Contains(name, "mono") {
					artifactName = name
					downloadURL = url
					ok = true
					break
				}
			}
		}
		if !ok {
			return nil, fmt.Errorf("templates not found for version %s", release.Version)
		}
	}

	downloadDir := filepath.Join(s.cacheDir(), "downloads")
	archivePath := filepath.Join(downloadDir, artifactName)

	dlOpts := download.DownloadOpts{
		Resume:  true,
		Mirrors: s.Config.Mirrors,
	}
	if err := download.File(ctx, downloadURL, archivePath, dlOpts); err != nil {
		return nil, fmt.Errorf("download failed: %w", err)
	}

	destDir := filepath.Join(s.templatesDir(), versionName)
	os.MkdirAll(filepath.Dir(destDir), 0o755)
	os.RemoveAll(destDir)

	if err := download.ExtractZip(archivePath, destDir); err != nil {
		return nil, fmt.Errorf("extraction failed: %w", err)
	}

	return &InstallResult{
		Version:      release.Version,
		VersionName:  versionName,
		ArtifactName: artifactName,
		IsNew:        true,
	}, nil
}
