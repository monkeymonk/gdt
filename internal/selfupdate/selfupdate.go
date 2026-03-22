package selfupdate

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/monkeymonk/gdt/internal/download"
	"github.com/monkeymonk/gdt/internal/metadata"
)

// Result contains the outcome of an update attempt.
type Result struct {
	Updated    bool
	NewVersion string
}

// Update checks for and applies the latest gdt release.
func Update(ctx context.Context, currentVersion string, apiURL string) (*Result, error) {
	token := os.Getenv("GITHUB_TOKEN")

	release, err := metadata.FetchLatestRelease(apiURL, token)
	if err != nil {
		return nil, fmt.Errorf("fetch latest release: %w", err)
	}

	latestVersion := release.TagName
	if latestVersion == "v"+currentVersion || latestVersion == currentVersion {
		return &Result{Updated: false}, nil
	}

	artifact := fmt.Sprintf("gdt-%s-%s-%s", latestVersion, runtime.GOOS, runtime.GOARCH)
	var downloadURL string
	for name, url := range release.Assets {
		if name == artifact+".tar.gz" || name == artifact+".zip" {
			downloadURL = url
			break
		}
	}
	if downloadURL == "" {
		return nil, fmt.Errorf("no binary found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	exe, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("resolve executable path: %w", err)
	}

	tmpPath := exe + ".new"
	if err := download.File(ctx, downloadURL, tmpPath, download.DownloadOpts{}); err != nil {
		return nil, fmt.Errorf("download update: %w", err)
	}

	if err := os.Chmod(tmpPath, 0755); err != nil {
		os.Remove(tmpPath)
		return nil, fmt.Errorf("chmod: %w", err)
	}

	if err := os.Rename(tmpPath, exe); err != nil {
		return nil, fmt.Errorf("failed to replace binary: %w\n\n  Try: sudo mv %s %s", err, tmpPath, exe)
	}

	return &Result{Updated: true, NewVersion: latestVersion}, nil
}
