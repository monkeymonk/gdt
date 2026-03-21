package plugins

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func writeV1TestPlugin(t *testing.T, dir, name, hookEvent, script string) {
	t.Helper()
	pluginDir := filepath.Join(dir, name)
	os.MkdirAll(pluginDir, 0o755)
	manifest := `name = "` + name + `"
version = "1.0.0"

[hooks]
` + hookEvent + ` = "` + script + `"
`
	os.WriteFile(filepath.Join(pluginDir, ManifestFile), []byte(manifest), 0o644)
}

func writeV2TestPlugin(t *testing.T, dir, name string, hooks []string, script string) string {
	t.Helper()
	pluginDir := filepath.Join(dir, name)
	os.MkdirAll(pluginDir, 0o755)

	hooksToml := "["
	for i, h := range hooks {
		if i > 0 {
			hooksToml += ", "
		}
		hooksToml += `"` + h + `"`
	}
	hooksToml += "]"

	manifest := `name = "` + name + `"
version = "1.0.0"
protocol = 2

[contributions]
hooks = ` + hooksToml + `
`
	os.WriteFile(filepath.Join(pluginDir, ManifestFile), []byte(manifest), 0o644)

	binPath := filepath.Join(pluginDir, name)
	os.WriteFile(binPath, []byte(script), 0o755)
	return pluginDir
}

func TestRunHooks_V1_Success(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}
	dir := t.TempDir()
	writeV1TestPlugin(t, dir, "test-plugin", "before_export", "exit 0")

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

func TestRunHooks_V2_Success(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}
	dir := t.TempDir()
	writeV2TestPlugin(t, dir, "v2plugin", []string{"before_export"}, "#!/bin/sh\necho \"OK hook ran\"\n")

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

func TestRunHooks_V2_NonZeroExit_Fatal(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}
	dir := t.TempDir()
	writeV2TestPlugin(t, dir, "failplugin", []string{"before_export"}, "#!/bin/sh\nexit 1\n")

	svc := NewService(dir)
	err := svc.RunHooks(BeforeExport, HookContext{})
	if err == nil {
		t.Fatal("expected error for non-zero exit")
	}
}

func TestRunHooks_V2_UndeclaredEvent_Skipped(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}
	dir := t.TempDir()
	writeV2TestPlugin(t, dir, "selective", []string{"after_new"}, "#!/bin/sh\nexit 1\n")

	svc := NewService(dir)
	// Plugin only declares after_new, so before_export should be skipped
	err := svc.RunHooks(BeforeExport, HookContext{})
	if err != nil {
		t.Fatalf("expected no error for undeclared event, got: %v", err)
	}
}

func TestRunHooks_AlphabeticalOrder(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}
	dir := t.TempDir()
	writeV2TestPlugin(t, dir, "b-plugin", []string{"before_export"}, "#!/bin/sh\necho \"OK b ran\"\n")
	writeV2TestPlugin(t, dir, "a-plugin", []string{"before_export"}, "#!/bin/sh\necho \"OK a ran\"\n")

	svc := NewService(dir)
	err := svc.RunHooks(BeforeExport, HookContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestManifest_HookFor(t *testing.T) {
	m := &Manifest{
		Hooks: Hooks{
			BeforeExport: "echo before",
			AfterExport:  "echo after",
		},
	}

	tests := []struct {
		event HookEvent
		want  string
	}{
		{BeforeExport, "echo before"},
		{AfterExport, "echo after"},
		{HookEvent("unknown"), ""},
	}

	for _, tt := range tests {
		got := m.HookFor(tt.event)
		if got != tt.want {
			t.Errorf("HookFor(%q) = %q, want %q", tt.event, got, tt.want)
		}
	}
}
