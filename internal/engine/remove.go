package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// Remove deletes an installed Godot engine version.
func (s *Service) Remove(_ context.Context, version string) error {
	if !s.IsInstalled(version) {
		return Actionable(
			fmt.Errorf("version %s is not installed", version),
			"gdt ls",
		)
	}

	dir := filepath.Join(s.VersionsDir(), version)
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("failed to remove %s: %w", version, err)
	}

	// Remove desktop launcher if no versions remain
	remaining, _ := s.List()
	if len(remaining) == 0 {
		s.removeDesktop()
	}

	return nil
}
