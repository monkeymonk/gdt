package plugins

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// UpdateResult holds the outcome of a single plugin update attempt.
type UpdateResult struct {
	Name string
	Err  error
}

// Update pulls the latest changes for plugins and rebuilds binaries.
// If name is empty, all plugins are updated.
func (s *Service) Update(name string) ([]UpdateResult, error) {
	pluginList, err := s.Discover()
	if err != nil {
		return nil, err
	}

	var results []UpdateResult
	for _, p := range pluginList {
		if name != "" && p.Manifest.Name != name {
			continue
		}
		gitCmd := exec.Command("git", "-C", p.Dir, "pull", "--ff-only")
		if err := gitCmd.Run(); err != nil {
			results = append(results, UpdateResult{Name: p.Manifest.Name, Err: fmt.Errorf("failed to update: %w", err)})
			continue
		}

		// Resolve binary after pull (re-download or rebuild)
		repoSlug := detectRepoSlug(p.Dir)
		if err := ResolveBinary(p.Dir, &p.Manifest, repoSlug); err != nil {
			results = append(results, UpdateResult{Name: p.Manifest.Name, Err: fmt.Errorf("binary resolve failed: %w", err)})
			continue
		}

		results = append(results, UpdateResult{Name: p.Manifest.Name})
	}
	return results, nil
}

// detectRepoSlug extracts the GitHub owner/repo from the git remote origin URL.
func detectRepoSlug(dir string) string {
	cmd := exec.Command("git", "-C", dir, "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	url := strings.TrimSpace(string(out))

	// Handle SSH format: git@github.com:owner/repo.git
	if strings.HasPrefix(url, "git@github.com:") {
		slug := strings.TrimPrefix(url, "git@github.com:")
		slug = strings.TrimSuffix(slug, ".git")
		return slug
	}

	// Handle HTTPS format
	url = strings.TrimPrefix(url, "https://github.com/")
	url = strings.TrimPrefix(url, "http://github.com/")
	url = strings.TrimSuffix(url, ".git")

	parts := strings.Split(url, "/")
	if len(parts) >= 2 {
		return parts[0] + "/" + parts[1]
	}
	return ""
}

// rebuildNeeded checks if the binary is missing or older than source files.
// For now, always attempt resolve if no binary exists.
func rebuildNeeded(dir string, m *Manifest) bool {
	binPath := filepath.Join(dir, m.Name)
	_, err := os.Stat(binPath)
	return os.IsNotExist(err)
}
