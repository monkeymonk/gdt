package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/monkeymonk/gdt/internal/plugins"
)

// runGit runs a git command with the given working directory, failing the
// test immediately (with combined output) on error.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// setupUpdatablePluginGitRepo creates a local git repository, clones it into
// pluginsDir/name, and drops a plugin.toml plus a same-named "binary" file
// so `gdt plugin update` can pull it (against the local clone, no network)
// and ResolveBinary's "binary already present" branch short-circuits
// without building or downloading anything.
func setupUpdatablePluginGitRepo(t *testing.T, pluginsDir, name string) {
	t.Helper()

	origin := t.TempDir()
	runGit(t, origin, "init")
	if err := os.WriteFile(filepath.Join(origin, "README.md"), []byte("origin\n"), 0o644); err != nil {
		t.Fatalf("seed origin file: %v", err)
	}
	runGit(t, origin, "add", "-A")
	runGit(t, origin, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-m", "init")

	pluginDir := filepath.Join(pluginsDir, name)
	runGit(t, "", "clone", origin, pluginDir)

	manifest := "name = \"" + name + "\"\nversion = \"1.0.0\"\nprotocol = 2\n"
	if err := os.WriteFile(filepath.Join(pluginDir, "plugin.toml"), []byte(manifest), 0o644); err != nil {
		t.Fatalf("write plugin manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, name), []byte("#!/bin/sh\necho ok\n"), 0o755); err != nil {
		t.Fatalf("write plugin binary: %v", err)
	}
}

// --- install ---

// TestPluginInstall_EmptyRepoArg_ReturnsError proves that running
// `gdt plugin install` with no repository argument and no TTY to prompt
// on returns an actionable error instead of proceeding with an empty repo.
func TestPluginInstall_EmptyRepoArg_ReturnsError(t *testing.T) {
	forceNonTTYStdin(t)
	app := &App{Home: t.TempDir()}
	cmd := newPluginInstallCmd(app)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for empty repository argument, got nil")
	}
	if !strings.Contains(err.Error(), "repository required") {
		t.Errorf("expected error to mention 'repository required', got: %v", err)
	}
}

// TestPluginInstall_AlreadyInstalled_ReturnsError proves that svc.Install's
// error (plugin already installed) propagates through the command as-is,
// without a real network clone: the "already installed" check in
// plugins.Service.Install happens before the git clone.
func TestPluginInstall_AlreadyInstalled_ReturnsError(t *testing.T) {
	home := t.TempDir()
	pluginsDir := filepath.Join(home, "plugins")
	if err := os.MkdirAll(filepath.Join(pluginsDir, "myplugin"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	app := &App{Home: home}
	cmd := newPluginInstallCmd(app)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{"someowner/myplugin"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for already-installed plugin, got nil")
	}
	if !strings.Contains(err.Error(), "already installed") {
		t.Errorf("expected error to mention 'already installed', got: %v", err)
	}
}

// --- list ---

// TestPluginList_WithContributions_FormatsSummaryLine proves that the list
// command's formatted line includes the contribution counts and flags
// (templates, presets, doctor) declared in a plugin's manifest.
func TestPluginList_WithContributions_FormatsSummaryLine(t *testing.T) {
	home := t.TempDir()
	pluginDir := filepath.Join(home, "plugins", "myplugin")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	manifest := `name = "myplugin"
version = "1.2.0"
protocol = 2

[contributions]
templates = ["a", "b"]
presets = ["c"]
doctor = true
`
	if err := os.WriteFile(filepath.Join(pluginDir, "plugin.toml"), []byte(manifest), 0o644); err != nil {
		t.Fatalf("write plugin manifest: %v", err)
	}

	app := &App{Home: home}
	cmd := newPluginListCmd(app)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	stdout := captureStdout(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	wantLine := "  myplugin v1.2.0 (2 templates, 1 presets, doctor)"
	if !strings.Contains(stdout, wantLine) {
		t.Errorf("expected line %q in output, got:\n%s", wantLine, stdout)
	}
}

// TestPluginList_NoPluginsInstalled_PrintsMessage proves that the empty-list
// branch prints "No plugins installed" instead of an empty "Installed
// plugins" header.
func TestPluginList_NoPluginsInstalled_PrintsMessage(t *testing.T) {
	app := &App{Home: t.TempDir()}
	cmd := newPluginListCmd(app)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	stderr := captureStderr(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(stderr, "No plugins installed") {
		t.Errorf("expected 'No plugins installed' message, got:\n%s", stderr)
	}
}

// --- update ---

// TestPluginUpdate_MixedResults_FormatsSuccessAndFailureLines proves that
// newPluginUpdateCmd's per-result formatting covers both branches (success
// and failure) in a single run: one plugin that is a real, pullable local
// git clone (updates cleanly) and one plugin directory with no .git (pull
// fails). This exercises plugin.go's own result-line formatting, not
// plugins.Service.Update's internals, which are covered separately.
func TestPluginUpdate_MixedResults_FormatsSuccessAndFailureLines(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows: fixture uses a real git clone and a #!/bin/sh script file")
	}

	home := t.TempDir()
	pluginsDir := filepath.Join(home, "plugins")
	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	setupUpdatablePluginGitRepo(t, pluginsDir, "goodplugin")

	badDir := filepath.Join(pluginsDir, "badplugin")
	if err := os.MkdirAll(badDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	badManifest := `name = "badplugin"
version = "1.0.0"
protocol = 2
`
	if err := os.WriteFile(filepath.Join(badDir, "plugin.toml"), []byte(badManifest), 0o644); err != nil {
		t.Fatalf("write plugin manifest: %v", err)
	}

	app := &App{Home: home}
	cmd := newPluginUpdateCmd(app)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	stderr := captureStderr(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(stderr, "  goodplugin updated") {
		t.Errorf("expected success line for goodplugin, got:\n%s", stderr)
	}
	if !strings.Contains(stderr, "  failed to update badplugin:") {
		t.Errorf("expected failure line for badplugin, got:\n%s", stderr)
	}
}

// --- new ---

// TestPluginNew_EmptyNameArg_ReturnsError proves that running
// `gdt plugin new` with no name argument and no TTY to prompt on returns an
// actionable error instead of proceeding with an empty name.
func TestPluginNew_EmptyNameArg_ReturnsError(t *testing.T) {
	forceNonTTYStdin(t)
	app := &App{Home: t.TempDir()}
	cmd := newPluginNewCmd(app)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for empty name argument, got nil")
	}
	if !strings.Contains(err.Error(), "name required") {
		t.Errorf("expected error to mention 'name required', got: %v", err)
	}
}

// TestPluginNew_DirectoryAlreadyExists_ReturnsError proves that
// svc.ScaffoldV2's error (target directory already exists) propagates
// through the command as-is.
func TestPluginNew_DirectoryAlreadyExists_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.MkdirAll(filepath.Join(dir, "gdt-myplugin"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	app := &App{Home: t.TempDir()}
	cmd := newPluginNewCmd(app)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{"myplugin"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for already-existing scaffold directory, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected error to mention 'already exists', got: %v", err)
	}
}

// TestPluginNew_Success_PrintsScaffoldMessageWithJoinedPath proves that a
// successful scaffold prints the "scaffolded at" and "Edit ..." hint lines,
// with the manifest hint path built via filepath.Join (not a hardcoded "/").
func TestPluginNew_Success_PrintsScaffoldMessageWithJoinedPath(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	app := &App{Home: t.TempDir()}
	cmd := newPluginNewCmd(app)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{"myplugin"})

	stderr := captureStderr(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	wantScaffoldDir := filepath.Join(".", "gdt-myplugin")
	if !strings.Contains(stderr, "Plugin scaffolded at "+wantScaffoldDir) {
		t.Errorf("expected scaffold message for %q, got:\n%s", wantScaffoldDir, stderr)
	}
	wantHint := filepath.Join(wantScaffoldDir, plugins.ManifestFile)
	if !strings.Contains(stderr, wantHint) {
		t.Errorf("expected hint to reference %q (via filepath.Join), got:\n%s", wantHint, stderr)
	}
	if _, err := os.Stat(filepath.Join(dir, "gdt-myplugin", "plugin.toml")); err != nil {
		t.Errorf("expected scaffolded manifest to exist: %v", err)
	}
}

// --- remove ---

// TestPluginRemove_EmptyNameArg_ReturnsError proves that running
// `gdt plugin remove` with no name argument and no TTY to prompt on returns
// an actionable error instead of proceeding with an empty name.
//
// Note: the confirm-abort branch (isTTY() true, user declines the removal
// prompt) is not covered here. This sandbox's `go test` stdin is, somewhat
// unusually, a real char-device (confirmed empirically — even explicit
// shell-level stdin redirection to a regular file does not change what
// the compiled test binary sees), so isTTY() reports true by default
// unless forced false via forceNonTTYStdin below. That default is not
// representative of a typical headless CI runner (GitHub Actions'
// ubuntu/macos/windows matrix), where stdin is normally not a terminal at
// all — a test relying on this sandbox's specific stdin plumbing to reach
// the confirm branch would be exercising an artifact of this environment,
// not portable behavior. promptConfirm also has no injectable test seam.
// This mirrors the same accepted gap noted for the sibling remove/plugin
// subcommands in this round.
func TestPluginRemove_EmptyNameArg_ReturnsError(t *testing.T) {
	forceNonTTYStdin(t)
	app := &App{Home: t.TempDir()}
	cmd := newPluginRemoveCmd(app)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for empty name argument, got nil")
	}
	if !strings.Contains(err.Error(), "name required") {
		t.Errorf("expected error to mention 'name required', got: %v", err)
	}
}
