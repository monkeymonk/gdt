package plugins

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ResolveBinary ensures the plugin binary exists after cloning.
// Resolution order:
//  1. Binary already present (shell scripts, pre-built)
//  2. Download from GitHub release (cross-platform)
//  3. Build from source (auto-detect or manifest [build] command)
func ResolveBinary(dir string, m *Manifest, repoSlug string) error {
	binName := m.Name
	binPath := filepath.Join(dir, binName)

	// 1. Binary already present (e.g. shell script committed as the plugin name)
	if info, err := os.Stat(binPath); err == nil && !info.IsDir() {
		slog.Debug("plugin binary already present", "path", binPath)
		return nil
	}

	// 2. Try GitHub release download
	if repoSlug != "" {
		if err := downloadReleaseBinary(binPath, binName, repoSlug); err == nil {
			slog.Debug("plugin binary downloaded from release", "path", binPath)
			return nil
		} else {
			slog.Debug("no release binary available", "error", err)
		}
	}

	// 3. Build from source
	return buildFromSource(dir, m)
}

// downloadReleaseBinary attempts to download a pre-built binary from
// the latest GitHub release matching the current OS/arch.
func downloadReleaseBinary(binPath, binName, repoSlug string) error {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repoSlug)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("github api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("github api returned %d", resp.StatusCode)
	}

	var release struct {
		Assets []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("parse release: %w", err)
	}

	// Match asset by name pattern: {name}-{os}-{arch}
	// Also try common variants (amd64/x86_64, arm64/aarch64)
	candidates := buildAssetCandidates(binName, goos, goarch)
	var downloadURL string
	for _, asset := range release.Assets {
		lower := strings.ToLower(asset.Name)
		for _, candidate := range candidates {
			if lower == candidate {
				downloadURL = asset.BrowserDownloadURL
				break
			}
		}
		if downloadURL != "" {
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no release asset for %s/%s", goos, goarch)
	}

	return downloadFile(binPath, downloadURL)
}

func buildAssetCandidates(name, goos, goarch string) []string {
	archAliases := map[string][]string{
		"amd64": {"amd64", "x86_64"},
		"arm64": {"arm64", "aarch64"},
		"386":   {"386", "i386"},
	}
	arches := archAliases[goarch]
	if len(arches) == 0 {
		arches = []string{goarch}
	}

	osAliases := map[string][]string{
		"darwin":  {"darwin", "macos"},
		"linux":   {"linux"},
		"windows": {"windows"},
	}
	oses := osAliases[goos]
	if len(oses) == 0 {
		oses = []string{goos}
	}

	var candidates []string
	for _, o := range oses {
		for _, a := range arches {
			candidates = append(candidates, fmt.Sprintf("%s-%s-%s", name, o, a))
			if goos == "windows" {
				candidates = append(candidates, fmt.Sprintf("%s-%s-%s.exe", name, o, a))
			}
		}
	}
	return candidates
}

func downloadFile(dest, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download returned %d", resp.StatusCode)
	}

	f, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

// buildFromSource compiles the plugin binary using the detected or declared
// build system.
func buildFromSource(dir string, m *Manifest) error {
	// Use explicit build command from manifest if provided
	if m.Build.Command != "" {
		return runBuildCommand(dir, m.Build.Command)
	}

	// Auto-detect build system
	binName := m.Name
	checks := []struct {
		file string
		cmd  string
	}{
		{"Makefile", "make build"},
		{"go.mod", fmt.Sprintf("go build -o %s .", binName)},
		{"Cargo.toml", "cargo build --release"},
		{"build.sh", "./build.sh"},
	}

	for _, c := range checks {
		if _, err := os.Stat(filepath.Join(dir, c.file)); err == nil {
			slog.Debug("auto-detected build system", "file", c.file, "command", c.cmd)
			if err := runBuildCommand(dir, c.cmd); err != nil {
				return fmt.Errorf("build failed (%s): %w", c.file, err)
			}
			// For Cargo, copy binary from target/release to plugin dir
			if c.file == "Cargo.toml" {
				return cargoCopyBinary(dir, binName)
			}
			return nil
		}
	}

	return fmt.Errorf("no build system detected and no binary found; add [build] to plugin.toml or include a pre-built binary")
}

func runBuildCommand(dir, command string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}
	cmd.Dir = dir
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func cargoCopyBinary(dir, binName string) error {
	src := filepath.Join(dir, "target", "release", binName)
	if runtime.GOOS == "windows" {
		src += ".exe"
	}
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("cargo binary not found at %s", src)
	}
	dst := filepath.Join(dir, binName)
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0755)
}
