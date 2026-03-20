package plugins

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func writeTestPlugin(t *testing.T, dir, name, hookEvent, script string) {
	t.Helper()
	pluginDir := filepath.Join(dir, name)
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `name = "` + name + `"
version = "1.0.0"

[hooks]
` + hookEvent + ` = "` + script + `"
`
	if err := os.WriteFile(filepath.Join(pluginDir, ManifestFile), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRunHooks_Success(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}

	dir := t.TempDir()
	writeTestPlugin(t, dir, "test-plugin", "before_export", "exit 0")

	svc := NewService(dir)
	err := svc.RunHooks(BeforeExport, HookContext{
		ProjectRoot:  dir,
		GodotVersion: "4.2.1",
		EnginePath:   "/usr/bin/godot",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestRunHooks_ExitCode2_Fatal(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}

	dir := t.TempDir()
	writeTestPlugin(t, dir, "fatal-plugin", "after_export", "exit 2")

	svc := NewService(dir)
	err := svc.RunHooks(AfterExport, HookContext{
		ProjectRoot:  dir,
		GodotVersion: "4.2.1",
		EnginePath:   "/usr/bin/godot",
	})
	if err == nil {
		t.Fatal("expected error for exit code 2, got nil")
	}
}

func TestRunHooks_NonZeroExit_Warning(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}

	dir := t.TempDir()
	writeTestPlugin(t, dir, "warn-plugin", "before_build", "exit 1")

	svc := NewService(dir)
	err := svc.RunHooks(BeforeBuild, HookContext{
		ProjectRoot:  dir,
		GodotVersion: "4.2.1",
		EnginePath:   "/usr/bin/godot",
	})
	if err != nil {
		t.Fatalf("expected warning only (no error), got: %v", err)
	}
}

func TestRunHooks_NoMatchingHook(t *testing.T) {
	dir := t.TempDir()
	writeTestPlugin(t, dir, "no-hook-plugin", "before_export", "exit 0")

	svc := NewService(dir)
	// Ask for AfterExport but plugin only has BeforeExport
	err := svc.RunHooks(AfterExport, HookContext{})
	if err != nil {
		t.Fatalf("expected no error for unmatched hook, got: %v", err)
	}
}

func TestManifest_HookFor(t *testing.T) {
	m := &Manifest{
		Hooks: Hooks{
			BeforeExport: "echo before",
			AfterExport:  "echo after",
			BeforeBuild:  "echo build",
		},
	}

	tests := []struct {
		event HookEvent
		want  string
	}{
		{BeforeExport, "echo before"},
		{AfterExport, "echo after"},
		{BeforeBuild, "echo build"},
		{HookEvent("unknown"), ""},
	}

	for _, tt := range tests {
		got := m.HookFor(tt.event)
		if got != tt.want {
			t.Errorf("HookFor(%q) = %q, want %q", tt.event, got, tt.want)
		}
	}
}
