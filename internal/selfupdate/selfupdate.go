package selfupdate

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/monkeymonk/gdt/internal/download"
	"github.com/monkeymonk/gdt/internal/metadata"
)

var osExecutable = os.Executable

// Result contains the outcome of an update attempt.
type Result struct {
	Updated    bool
	NewVersion string
}

func trimV(version string) string {
	return strings.TrimPrefix(version, "v")
}

// Update checks for and applies the latest gdt release.
func Update(ctx context.Context, currentVersion string, apiURL string) (*Result, error) {
	token := os.Getenv("GITHUB_TOKEN")

	release, err := metadata.FetchLatestRelease(apiURL, token)
	if err != nil {
		return nil, fmt.Errorf("fetch latest release: %w", err)
	}

	latestVersion := release.TagName
	if trimV(latestVersion) == trimV(currentVersion) {
		return &Result{Updated: false}, nil
	}

	ext := ".tar.gz"
	if runtime.GOOS == "windows" {
		ext = ".zip"
	}

	ver := strings.TrimPrefix(latestVersion, "v")
	primary := fmt.Sprintf("gdt-%s-%s-%s%s", ver, runtime.GOOS, runtime.GOARCH, ext)
	fallback := fmt.Sprintf("gdt-%s-%s-%s%s", latestVersion, runtime.GOOS, runtime.GOARCH, ext)

	var downloadURL, assetName string
	for _, candidate := range []string{primary, fallback} {
		if url, ok := release.Assets[candidate]; ok {
			downloadURL = url
			assetName = candidate
			break
		}
	}
	if downloadURL == "" {
		return nil, fmt.Errorf("no binary found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	tmp, err := os.MkdirTemp("", "gdt-selfupdate-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmp)

	archivePath := filepath.Join(tmp, "archive"+ext)
	if err := download.File(ctx, downloadURL, archivePath, download.DownloadOpts{}); err != nil {
		return nil, fmt.Errorf("download update: %w", err)
	}

	if checksumsURL, ok := release.Assets["checksums.txt"]; ok {
		checksumsPath := filepath.Join(tmp, "checksums.txt")
		if err := download.File(ctx, checksumsURL, checksumsPath, download.DownloadOpts{}); err != nil {
			return nil, fmt.Errorf("download checksums: %w", err)
		}
		content, err := os.ReadFile(checksumsPath)
		if err != nil {
			return nil, fmt.Errorf("read checksums: %w", err)
		}
		expected := metadata.FindChecksum(string(content), assetName)
		if expected == "" {
			fmt.Fprintf(os.Stderr, "  %s not listed in checksums.txt; skipping verification\n", assetName)
		} else if err := download.VerifySHA256(archivePath, expected); err != nil {
			return nil, fmt.Errorf("verify checksum: %w", err)
		}
	} else {
		fmt.Fprintln(os.Stderr, "  no checksums.txt in release; skipping verification")
	}

	if ext == ".zip" {
		err = download.ExtractZip(archivePath, tmp)
	} else {
		err = download.ExtractTarGz(archivePath, tmp)
	}
	if err != nil {
		return nil, fmt.Errorf("extract update: %w", err)
	}

	binName := "gdt"
	if runtime.GOOS == "windows" {
		binName = "gdt.exe"
	}
	newBin := filepath.Join(tmp, binName)
	if _, err := os.Stat(newBin); err != nil {
		return nil, fmt.Errorf("binary not found in archive")
	}

	exe, err := osExecutable()
	if err != nil {
		return nil, fmt.Errorf("resolve executable path: %w", err)
	}

	stagePath := exe + ".new"
	if err := copyFile(newBin, stagePath); err != nil {
		os.Remove(stagePath)
		return nil, fmt.Errorf("stage new binary: %w", err)
	}

	if err := os.Chmod(stagePath, 0755); err != nil {
		os.Remove(stagePath)
		return nil, fmt.Errorf("chmod: %w", err)
	}

	if err := replaceBinary(exe, stagePath); err != nil {
		os.Remove(stagePath)
		return nil, fmt.Errorf("failed to replace binary: %w\n\n  Try: sudo mv %s %s", err, stagePath, exe)
	}

	return &Result{Updated: true, NewVersion: latestVersion}, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)
	out.Close()
	return err
}
