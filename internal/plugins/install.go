package plugins

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Install clones a plugin from a Git repository into the plugins directory.
// Returns the parsed manifest on success.
func (s *Service) Install(repo string) (*Manifest, error) {
	repoURL := repo
	if !strings.HasPrefix(repo, "http") {
		repoURL = "https://github.com/" + repo
	}

	parts := strings.Split(strings.TrimSuffix(repo, "/"), "/")
	name := parts[len(parts)-1]

	destDir := filepath.Join(s.Dir, name)
	if _, err := os.Stat(destDir); err == nil {
		return nil, fmt.Errorf("plugin %s is already installed\n\n  gdt plugin remove %s", name, name)
	}

	gitCmd := exec.Command("git", "clone", "--depth", "1", repoURL, destDir)
	gitCmd.Stdout = os.Stderr
	gitCmd.Stderr = os.Stderr
	if err := gitCmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to clone %s: %w", repoURL, err)
	}

	manifestPath := filepath.Join(destDir, ManifestFile)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		os.RemoveAll(destDir)
		return nil, fmt.Errorf("plugin missing %s manifest", ManifestFile)
	}

	m, err := ParseManifest(data)
	if err != nil {
		os.RemoveAll(destDir)
		return nil, fmt.Errorf("invalid plugin manifest: %w", err)
	}

	return m, nil
}
