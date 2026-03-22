package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/monkeymonk/gdt/internal/config"
	"github.com/monkeymonk/gdt/internal/download"
	"github.com/monkeymonk/gdt/internal/metadata"
	"github.com/monkeymonk/gdt/internal/platform"
	"github.com/monkeymonk/gdt/internal/project"
)

type Service struct {
	Home     string
	Platform platform.Info
	Config   *config.Config
}

type InstallOpts struct {
	Mono    bool
	Force   bool
	Refresh bool
}

type InstallResult struct {
	Version      string
	VersionName  string
	ArtifactName string
	IsNew        bool
}

type ResolvedVersion struct {
	Version    string
	BinaryPath string
	Source     string // "file", "env", "config", "latest"
}

type InstalledVersion struct {
	Version   string
	IsDefault bool
}

func (s *Service) ResolveProject(cwd string) (projectRoot string, resolved ResolvedVersion, err error) {
	projectRoot, err = project.DetectRoot(cwd)
	if err != nil {
		return "", ResolvedVersion{}, fmt.Errorf("no Godot project found\n\n  Run from a directory containing project.godot")
	}
	resolved, err = s.Resolve(cwd)
	if err != nil {
		return "", ResolvedVersion{}, err
	}
	return projectRoot, resolved, nil
}

func NewService(home string, plat platform.Info, cfg *config.Config) *Service {
	return &Service{Home: home, Platform: plat, Config: cfg}
}

func (s *Service) VersionsDir() string  { return filepath.Join(s.Home, "versions") }
func (s *Service) TemplatesDir() string { return filepath.Join(s.Home, "templates") }
func (s *Service) CacheDir() string     { return filepath.Join(s.Home, "cache") }
func (s *Service) CachePath() string    { return filepath.Join(s.Home, "cache", "releases.json") }

// listDirectories returns sorted names of all subdirectories in dir.
// Returns nil, nil if dir does not exist.
func listDirectories(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
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

type downloadSpec struct {
	CachePath       string
	APIURL          string
	Token           string
	Query           string
	Mono            bool
	Force           bool
	Refresh         bool
	DestDir         string
	ResolveArtifact func(release *metadata.Release, plat platform.Info, mono bool) (name string, err error)
	PostInstall     func()
	VerifyChecksum  bool
}

func (s *Service) downloadAndInstall(ctx context.Context, spec downloadSpec) (*InstallResult, error) {
	// 1. Load releases
	releases, err := metadata.EnsureCache(spec.CachePath, spec.APIURL, spec.Token, spec.Refresh)
	if err != nil {
		return nil, err
	}

	// 2. Resolve version
	release, err := metadata.ResolveVersion(releases, spec.Query)
	if err != nil {
		return nil, err
	}

	// 3. Build version name
	versionName := release.Version
	if spec.Mono {
		versionName += "-mono"
	}

	// 4. Check if already installed
	destDir := filepath.Join(spec.DestDir, versionName)
	if !spec.Force {
		if _, err := os.Stat(destDir); err == nil {
			return &InstallResult{
				Version:     release.Version,
				VersionName: versionName,
				IsNew:       false,
			}, ErrAlreadyInstalled
		}
	}

	// 5. Resolve artifact
	artifactName, err := spec.ResolveArtifact(release, s.Platform, spec.Mono)
	if err != nil {
		return nil, err
	}
	downloadURL, ok := release.Assets[artifactName]
	if !ok {
		return nil, fmt.Errorf("artifact %q not found for version %s", artifactName, release.Version)
	}

	downloadDir := filepath.Join(s.CacheDir(), "downloads")

	// 6. Download checksum file (best-effort, only if VerifyChecksum)
	var expectedChecksum string
	if spec.VerifyChecksum {
		if checksumURL, ok := release.Assets["SHA512-SUMS.txt"]; ok {
			checksumPath := filepath.Join(downloadDir, "SHA512-SUMS.txt")
			if err := download.File(ctx, checksumURL, checksumPath, download.DownloadOpts{}); err == nil {
				if data, err := os.ReadFile(checksumPath); err == nil {
					expectedChecksum = metadata.FindChecksum(string(data), artifactName)
				}
			}
		}
	}

	// 7. Download artifact
	dlOpts := download.DownloadOpts{
		Resume:  true,
		Mirrors: s.Config.Mirrors,
	}
	archivePath := filepath.Join(downloadDir, artifactName)
	if err := download.File(ctx, downloadURL, archivePath, dlOpts); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDownloadFailed, err)
	}

	// 8. Verify checksum
	if expectedChecksum != "" {
		if err := download.VerifyChecksum(archivePath, expectedChecksum); err != nil {
			os.Remove(archivePath)
			return nil, ErrChecksumMismatch
		}
	}

	// 9. Extract to tmpDir, then rename to destDir (atomic)
	tmpDir := filepath.Join(s.CacheDir(), "tmp")
	os.MkdirAll(tmpDir, 0755)
	if err := download.ExtractZip(archivePath, tmpDir); err != nil {
		return nil, fmt.Errorf("extraction failed: %w", err)
	}

	os.MkdirAll(filepath.Dir(destDir), 0755)
	os.RemoveAll(destDir)
	if err := os.Rename(tmpDir, destDir); err != nil {
		return nil, fmt.Errorf("failed to install: %w", err)
	}

	// 10. Post-install hook (best-effort)
	if spec.PostInstall != nil {
		spec.PostInstall()
	}

	return &InstallResult{
		Version:      release.Version,
		VersionName:  versionName,
		ArtifactName: artifactName,
		IsNew:        true,
	}, nil
}
