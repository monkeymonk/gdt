package plugins

import (
	"fmt"
	"os"
	"path/filepath"
)

// Remove deletes an installed plugin by name.
func (s *Service) Remove(name string) error {
	dir := filepath.Join(s.Dir, name)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("plugin %s not found", name)
	}
	return os.RemoveAll(dir)
}
