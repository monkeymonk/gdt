package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/monkeymonk/gdt/internal/metadata"
	"github.com/monkeymonk/gdt/internal/platform"
)

// ListTemplates returns the names of all installed template sets, sorted alphabetically.
func (s *Service) ListTemplates() ([]string, error) {
	return listDirectories(s.TemplatesDir())
}

// TemplatesInstalled checks whether templates for the given version are installed.
func (s *Service) TemplatesInstalled(version string) bool {
	_, err := os.Stat(filepath.Join(s.TemplatesDir(), version))
	return err == nil
}

// RemoveTemplates deletes installed export templates for the given version.
func (s *Service) RemoveTemplates(version string) error {
	dir := filepath.Join(s.TemplatesDir(), version)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return &ActionableError{
			Err:        fmt.Errorf("templates for %s are not installed", version),
			Suggestion: "gdt templates list",
		}
	}
	return os.RemoveAll(dir)
}

// InstallTemplates downloads and extracts export templates for the given version.
func (s *Service) InstallTemplates(ctx context.Context, query string, opts InstallOpts) (*InstallResult, error) {
	return s.downloadAndInstall(ctx, downloadSpec{
		CachePath: s.CachePath(),
		APIURL:    s.Config.GodotAPIURL(),
		Token:     os.Getenv("GITHUB_TOKEN"),
		Query:     query,
		Mono:      opts.Mono,
		Force:     opts.Force,
		Refresh:   opts.Refresh,
		DestDir:   s.TemplatesDir(),
		ResolveArtifact: func(release *metadata.Release, plat platform.Info, mono bool) (string, error) {
			name := metadata.TemplateArtifactName(release.Version, mono)
			if _, ok := release.Assets[name]; ok {
				return name, nil
			}
			// Fallback: scan assets for export_templates
			for assetName := range release.Assets {
				if strings.Contains(assetName, "export_templates") {
					if mono && strings.Contains(assetName, "mono") {
						return assetName, nil
					} else if !mono && !strings.Contains(assetName, "mono") {
						return assetName, nil
					}
				}
			}
			return "", fmt.Errorf("templates not found for version %s", release.Version)
		},
		PostInstall:    nil,
		VerifyChecksum: false,
	})
}
