package plugins

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Install clones a plugin from a Git repository into the plugins directory,
// then resolves the binary (download release or build from source).
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

	// Resolve binary: download release → build from source → already present
	repoSlug := extractRepoSlug(repo)
	if err := ResolveBinary(destDir, m, repoSlug); err != nil {
		os.RemoveAll(destDir)
		return nil, fmt.Errorf("failed to resolve plugin binary: %w", err)
	}

	return m, nil
}

// extractRepoSlug extracts "owner/repo" from a repository reference.
// Returns empty string if the format is unrecognized.
func extractRepoSlug(repo string) string {
	// Already in owner/repo format
	repo = strings.TrimSuffix(repo, "/")
	repo = strings.TrimSuffix(repo, ".git")

	if strings.HasPrefix(repo, "http") {
		// https://github.com/owner/repo → owner/repo
		repo = strings.TrimPrefix(repo, "https://github.com/")
		repo = strings.TrimPrefix(repo, "http://github.com/")
	}

	parts := strings.Split(repo, "/")
	if len(parts) == 2 {
		return parts[0] + "/" + parts[1]
	}
	return ""
}
