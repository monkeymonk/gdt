package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/monkeymonk/gdt/internal/download"
	"github.com/monkeymonk/gdt/internal/metadata"
)

const apiURL = "https://api.github.com/repos/godotengine/godot/releases"

// Install downloads and installs a Godot engine version.
func (s *Service) Install(ctx context.Context, version string, opts InstallOpts) (*InstallResult, error) {
	token := os.Getenv("GDT_GITHUB_TOKEN")

	// 1. Load releases
	releases, err := metadata.EnsureCache(s.cachePath(), apiURL, token, opts.Refresh)
	if err != nil {
		return nil, err
	}

	// 2. Resolve version
	release, err := metadata.ResolveVersion(releases, version)
	if err != nil {
		return nil, err
	}

	// 3. Build version name
	versionName := release.Version
	if opts.Mono {
		versionName += "-mono"
	}

	// 4. Check if already installed
	if !opts.Force && s.IsInstalled(versionName) {
		return &InstallResult{
			Version:     release.Version,
			VersionName: versionName,
			IsNew:       false,
		}, ErrAlreadyInstalled
	}

	// 5. Resolve artifact
	artifactName, err := metadata.ResolveEngineArtifact(release, s.Platform, opts.Mono)
	if err != nil {
		return nil, err
	}
	downloadURL, ok := release.Assets[artifactName]
	if !ok {
		return nil, fmt.Errorf("artifact %q not found for version %s", artifactName, release.Version)
	}

	// 6. Download checksum file (best-effort)
	downloadDir := filepath.Join(s.cacheDir(), "downloads")
	var expectedChecksum string
	if checksumURL, ok := release.Assets["SHA512-SUMS.txt"]; ok {
		checksumPath := filepath.Join(downloadDir, "SHA512-SUMS.txt")
		if err := download.File(ctx, checksumURL, checksumPath, download.DownloadOpts{}); err == nil {
			if data, err := os.ReadFile(checksumPath); err == nil {
				expectedChecksum = metadata.FindChecksum(string(data), artifactName)
			}
		}
	}

	// 7. Download engine artifact
	archivePath := filepath.Join(downloadDir, artifactName)
	if err := download.File(ctx, downloadURL, archivePath, download.DownloadOpts{}); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDownloadFailed, err)
	}

	// 8. Verify checksum
	if expectedChecksum != "" {
		if err := download.VerifyChecksum(archivePath, expectedChecksum); err != nil {
			os.Remove(archivePath)
			return nil, ErrChecksumMismatch
		}
	}

	// 9. Extract
	destDir := filepath.Join(s.versionsDir(), versionName)
	tmpDir := filepath.Join(s.cacheDir(), "tmp")
	os.MkdirAll(tmpDir, 0755)
	if err := download.ExtractZip(archivePath, tmpDir); err != nil {
		return nil, fmt.Errorf("extraction failed: %w", err)
	}

	os.MkdirAll(filepath.Dir(destDir), 0755)
	os.RemoveAll(destDir)
	if err := os.Rename(tmpDir, destDir); err != nil {
		return nil, fmt.Errorf("failed to install: %w", err)
	}

	// 10. Desktop integration (best-effort)
	s.installDesktop()

	// 11. Return result
	return &InstallResult{
		Version:      release.Version,
		VersionName:  versionName,
		ArtifactName: artifactName,
		IsNew:        true,
	}, nil
}
