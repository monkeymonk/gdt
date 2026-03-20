package plugins

import (
	"fmt"
	"os/exec"
)

// UpdateResult holds the outcome of a single plugin update attempt.
type UpdateResult struct {
	Name string
	Err  error
}

// Update pulls the latest changes for plugins. If name is empty, all plugins
// are updated. Returns results for each plugin attempted.
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
		} else {
			results = append(results, UpdateResult{Name: p.Manifest.Name})
		}
	}
	return results, nil
}
