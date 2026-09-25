package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

// TestServiceDiscoverCachesResult verifies that Service.Discover only scans
// the filesystem once: a plugin added to the directory after the first
// Discover() call must not show up on a second call against the same
// Service instance.
func TestServiceDiscoverCachesResult(t *testing.T) {
	dir := t.TempDir()

	firstDir := filepath.Join(dir, "gdt-first")
	if err := os.MkdirAll(firstDir, 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(firstDir, "plugin.toml"), []byte(`
name = "first"
version = "1.0.0"
commands = ["first"]
`), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	svc := NewService(dir)

	first, err := svc.Discover()
	if err != nil {
		t.Fatalf("first Discover: %v", err)
	}
	if len(first) != 1 {
		t.Fatalf("expected 1 plugin after first Discover, got %d", len(first))
	}

	// Mutate the plugin directory between calls: a cached Discover must
	// not observe this addition.
	secondDir := filepath.Join(dir, "gdt-second")
	if err := os.MkdirAll(secondDir, 0755); err != nil {
		t.Fatalf("mutate: %v", err)
	}
	if err := os.WriteFile(filepath.Join(secondDir, "plugin.toml"), []byte(`
name = "second"
version = "1.0.0"
commands = ["second"]
`), 0644); err != nil {
		t.Fatalf("mutate: %v", err)
	}

	second, err := svc.Discover()
	if err != nil {
		t.Fatalf("second Discover: %v", err)
	}
	if len(second) != 1 {
		t.Fatalf("expected cached Discover to still return 1 plugin, got %d", len(second))
	}
	if second[0].Manifest.Name != "first" {
		t.Errorf("expected cached plugin %q, got %q", "first", second[0].Manifest.Name)
	}
}

// TestServiceDiscoverCachesErr verifies that a Discover error is also
// cached: once the first call fails, a later call — even after the
// filesystem state that caused the failure has been fixed — must keep
// returning the same cached error rather than re-scanning.
func TestServiceDiscoverCachesErr(t *testing.T) {
	dir := t.TempDir()
	pluginsPath := filepath.Join(dir, "plugins")

	// Seed a *file* where the plugins directory is expected, so
	// os.ReadDir fails with a real (non-NotExist) error on first call.
	if err := os.WriteFile(pluginsPath, []byte("not a directory"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	svc := NewService(pluginsPath)

	_, firstErr := svc.Discover()
	if firstErr == nil {
		t.Fatalf("expected first Discover to fail against a file, got nil error")
	}

	// Fix the filesystem state: remove the file and put a real,
	// valid plugin directory in its place.
	if err := os.Remove(pluginsPath); err != nil {
		t.Fatalf("mutate: %v", err)
	}
	pluginDir := filepath.Join(pluginsPath, "gdt-fixed")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatalf("mutate: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "plugin.toml"), []byte(`
name = "fixed"
version = "1.0.0"
commands = ["fixed"]
`), 0644); err != nil {
		t.Fatalf("mutate: %v", err)
	}

	second, secondErr := svc.Discover()
	if secondErr == nil || secondErr.Error() != firstErr.Error() {
		t.Fatalf("expected cached error %v on second call, got plugins=%v err=%v", firstErr, second, secondErr)
	}
}
