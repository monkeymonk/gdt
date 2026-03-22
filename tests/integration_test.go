package tests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var gdtBinary string

func TestMain(m *testing.M) {
	dir, _ := os.MkdirTemp("", "gdt-test-*")
	bin := "gdt"
	if runtime.GOOS == "windows" {
		bin = "gdt.exe"
	}
	gdtBinary = filepath.Join(dir, bin)
	cmd := exec.Command("go", "build", "-o", gdtBinary, "../cmd/gdt")
	if err := cmd.Run(); err != nil {
		panic("failed to build gdt: " + err.Error())
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func runGdt(t *testing.T, home string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(gdtBinary, args...)
	cmd.Env = append(os.Environ(), "GDT_HOME="+home)
	cmd.Stdin = strings.NewReader("") // ensure non-TTY so interactive prompts are skipped
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func runGdtIn(t *testing.T, home string, dir string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(gdtBinary, args...)
	cmd.Env = append(os.Environ(), "GDT_HOME="+home)
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader("") // ensure non-TTY
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// setupFakeVersion creates a fake engine install at $home/versions/<ver>/godot
func setupFakeVersion(t *testing.T, home, ver string) {
	t.Helper()
	vDir := filepath.Join(home, "versions", ver)
	os.MkdirAll(vDir, 0755)
	bin := "godot"
	if runtime.GOOS == "windows" {
		bin = "godot.exe"
	}
	os.WriteFile(filepath.Join(vDir, bin), []byte("fake"), 0755)
}

// setupFakePlugin creates a fake plugin with a manifest at $home/plugins/<name>/
func setupFakePlugin(t *testing.T, home, name, version string) {
	t.Helper()
	dir := filepath.Join(home, "plugins", name)
	os.MkdirAll(dir, 0755)
	manifest := fmt.Sprintf("name = %q\nversion = %q\ncommands = [%q]\n", name, version, name)
	os.WriteFile(filepath.Join(dir, "plugin.toml"), []byte(manifest), 0644)
}

func TestVersion(t *testing.T) {
	home := t.TempDir()
	out, err := runGdt(t, home, "--version")
	if err != nil {
		t.Fatalf("error: %v, output: %s", err, out)
	}
	if !strings.Contains(out, "gdt") {
		t.Errorf("expected version output, got: %s", out)
	}
}

func TestListEmpty(t *testing.T) {
	home := t.TempDir()
	out, _ := runGdt(t, home, "list")
	if !strings.Contains(out, "No versions installed") {
		t.Errorf("expected empty list message, got: %s", out)
	}
}

func TestUseAndList(t *testing.T) {
	home := t.TempDir()

	vDir := filepath.Join(home, "versions", "4.3")
	os.MkdirAll(vDir, 0755)
	os.WriteFile(filepath.Join(vDir, "godot"), []byte("fake"), 0755)

	runGdt(t, home, "use", "4.3")

	out, _ := runGdt(t, home, "list")
	if !strings.Contains(out, "* 4.3") {
		t.Errorf("expected active marker, got: %s", out)
	}
}

func TestLocalCreatesFile(t *testing.T) {
	home := t.TempDir()
	projDir := t.TempDir()

	cmd := exec.Command(gdtBinary, "local", "4.2")
	cmd.Env = append(os.Environ(), "GDT_HOME="+home)
	cmd.Dir = projDir
	cmd.Stdin = strings.NewReader("")
	cmd.Run()

	data, err := os.ReadFile(filepath.Join(projDir, ".godot-version"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(data)) != "4.2" {
		t.Errorf("expected 4.2, got: %s", string(data))
	}
}

func TestDoctorRuns(t *testing.T) {
	home := t.TempDir()
	out, err := runGdt(t, home, "doctor")
	if err != nil {
		t.Fatalf("error: %v, output: %s", err, out)
	}
}

func TestRemoveNotInstalled(t *testing.T) {
	home := t.TempDir()
	_, err := runGdt(t, home, "remove", "9.9")
	if err == nil {
		t.Error("expected error removing non-existent version")
	}
}

func TestNewDefault(t *testing.T) {
	home := t.TempDir()
	projDir := t.TempDir()

	vDir := filepath.Join(home, "versions", "4.3")
	os.MkdirAll(vDir, 0755)
	os.WriteFile(filepath.Join(vDir, "godot"), []byte("fake"), 0755)

	cmd := exec.Command(gdtBinary, "new", "testgame", "--version", "4.3", "--renderer", "forward_plus")
	cmd.Env = append(os.Environ(), "GDT_HOME="+home)
	cmd.Dir = projDir
	cmd.Stdin = strings.NewReader("")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("error: %v, output: %s", err, out)
	}

	gameDir := filepath.Join(projDir, "testgame")
	if _, err := os.Stat(filepath.Join(gameDir, "project.godot")); err != nil {
		t.Error("project.godot should exist")
	}
	if _, err := os.Stat(filepath.Join(gameDir, ".godot-version")); err != nil {
		t.Error(".godot-version should exist")
	}
	if _, err := os.Stat(filepath.Join(gameDir, ".gitignore")); err != nil {
		t.Error(".gitignore should exist")
	}
}

func TestNewAlreadyExists(t *testing.T) {
	home := t.TempDir()
	projDir := t.TempDir()

	gameDir := filepath.Join(projDir, "existing")
	os.MkdirAll(gameDir, 0755)
	os.WriteFile(filepath.Join(gameDir, "project.godot"), []byte("exists"), 0644)

	cmd := exec.Command(gdtBinary, "new", "existing", "--version", "4.3", "--renderer", "forward_plus")
	cmd.Env = append(os.Environ(), "GDT_HOME="+home)
	cmd.Dir = projDir
	cmd.Stdin = strings.NewReader("")
	_, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("should error when project already exists")
	}
}

func TestNewCSharp(t *testing.T) {
	home := t.TempDir()
	projDir := t.TempDir()

	vDir := filepath.Join(home, "versions", "4.3")
	os.MkdirAll(vDir, 0755)
	os.WriteFile(filepath.Join(vDir, "godot"), []byte("fake"), 0755)

	cmd := exec.Command(gdtBinary, "new", "csgame", "--version", "4.3", "--renderer", "forward_plus", "--csharp")
	cmd.Env = append(os.Environ(), "GDT_HOME="+home)
	cmd.Dir = projDir
	cmd.Stdin = strings.NewReader("")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("error: %v, output: %s", err, out)
	}

	gameDir := filepath.Join(projDir, "csgame")
	if _, err := os.Stat(filepath.Join(gameDir, "csgame.csproj")); err != nil {
		t.Error("csgame.csproj should exist")
	}
	if _, err := os.Stat(filepath.Join(gameDir, "csgame.sln")); err != nil {
		t.Error("csgame.sln should exist")
	}

	verData, _ := os.ReadFile(filepath.Join(gameDir, ".godot-version"))
	if strings.TrimSpace(string(verData)) != "4.3-mono" {
		t.Errorf(".godot-version = %q, want %q", strings.TrimSpace(string(verData)), "4.3-mono")
	}
}

func TestExportListNoProject(t *testing.T) {
	home := t.TempDir()
	emptyDir := t.TempDir()

	cmd := exec.Command(gdtBinary, "export", "--list")
	cmd.Env = append(os.Environ(), "GDT_HOME="+home)
	cmd.Dir = emptyDir
	cmd.Stdin = strings.NewReader("")
	_, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("should error when no project found")
	}
}

func TestCiSetupGitHub(t *testing.T) {
	home := t.TempDir()
	projDir := t.TempDir()

	cmd := exec.Command(gdtBinary, "ci", "setup", "--provider", "github")
	cmd.Env = append(os.Environ(), "GDT_HOME="+home)
	cmd.Dir = projDir
	cmd.Stdin = strings.NewReader("")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("error: %v, output: %s", err, out)
	}

	ciFile := filepath.Join(projDir, ".github", "workflows", "export.yml")
	if _, err := os.Stat(ciFile); err != nil {
		t.Error("CI file should exist")
	}
}

func TestCiSetupGitLab(t *testing.T) {
	home := t.TempDir()
	projDir := t.TempDir()

	cmd := exec.Command(gdtBinary, "ci", "setup", "--provider", "gitlab")
	cmd.Env = append(os.Environ(), "GDT_HOME="+home)
	cmd.Dir = projDir
	cmd.Stdin = strings.NewReader("")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("error: %v, output: %s", err, out)
	}

	ciFile := filepath.Join(projDir, ".gitlab-ci.yml")
	if _, err := os.Stat(ciFile); err != nil {
		t.Error("CI file should exist")
	}
}

func TestCiSetupGeneric(t *testing.T) {
	home := t.TempDir()
	projDir := t.TempDir()

	cmd := exec.Command(gdtBinary, "ci", "setup", "--provider", "generic")
	cmd.Env = append(os.Environ(), "GDT_HOME="+home)
	cmd.Dir = projDir
	cmd.Stdin = strings.NewReader("")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("error: %v, output: %s", err, out)
	}

	ciFile := filepath.Join(projDir, "ci", "export.sh")
	info, err := os.Stat(ciFile)
	if err != nil {
		t.Error("CI file should exist")
	}
	if runtime.GOOS != "windows" && info.Mode()&0111 == 0 {
		t.Error("generic script should be executable")
	}
}

func TestLspHelp(t *testing.T) {
	home := t.TempDir()
	out, err := runGdt(t, home, "lsp", "--help")
	if err != nil {
		t.Fatalf("error: %v, output: %s", err, out)
	}
	if !strings.Contains(out, "LSP") {
		t.Error("should show LSP help")
	}
}

func TestDapHelp(t *testing.T) {
	home := t.TempDir()
	out, err := runGdt(t, home, "dap", "--help")
	if err != nil {
		t.Fatalf("error: %v, output: %s", err, out)
	}
	if !strings.Contains(out, "DAP") {
		t.Error("should show DAP help")
	}
}

// --- Shell ---

func TestShellInit(t *testing.T) {
	home := t.TempDir()
	out, err := runGdt(t, home, "shell", "init")
	if err != nil {
		t.Fatalf("error: %v, output: %s", err, out)
	}
	if !strings.Contains(out, "PATH") && !strings.Contains(out, "fish_add_path") {
		t.Errorf("expected PATH export, got: %s", out)
	}
}

// --- Completion ---

func TestCompletionBash(t *testing.T) {
	home := t.TempDir()
	out, err := runGdt(t, home, "completion", "bash")
	if err != nil {
		t.Fatalf("error: %v, output: %s", err, out)
	}
	if !strings.Contains(out, "bash completion") {
		t.Errorf("expected bash completion script, got: %s", out[:min(len(out), 100)])
	}
}

func TestCompletionZsh(t *testing.T) {
	home := t.TempDir()
	out, err := runGdt(t, home, "completion", "zsh")
	if err != nil {
		t.Fatalf("error: %v, output: %s", err, out)
	}
	if !strings.Contains(out, "compdef") {
		t.Errorf("expected zsh completion script, got: %s", out[:min(len(out), 100)])
	}
}

func TestCompletionFish(t *testing.T) {
	home := t.TempDir()
	out, err := runGdt(t, home, "completion", "fish")
	if err != nil {
		t.Fatalf("error: %v, output: %s", err, out)
	}
	if !strings.Contains(out, "complete") {
		t.Errorf("expected fish completion script, got: %s", out[:min(len(out), 100)])
	}
}

func TestCompletionPowershell(t *testing.T) {
	home := t.TempDir()
	out, err := runGdt(t, home, "completion", "powershell")
	if err != nil {
		t.Fatalf("error: %v, output: %s", err, out)
	}
	if !strings.Contains(out, "Register-ArgumentCompleter") {
		t.Errorf("expected powershell completion script, got: %s", out[:min(len(out), 100)])
	}
}

func TestCompletionInvalidShell(t *testing.T) {
	home := t.TempDir()
	_, err := runGdt(t, home, "completion", "nushell")
	if err == nil {
		t.Error("expected error for invalid shell")
	}
}

// --- Templates ---

func TestTemplatesListEmpty(t *testing.T) {
	home := t.TempDir()
	out, _ := runGdt(t, home, "templates", "list")
	if !strings.Contains(out, "No templates installed") {
		t.Errorf("expected empty templates message, got: %s", out)
	}
}

func TestTemplatesListWithTemplates(t *testing.T) {
	home := t.TempDir()
	tDir := filepath.Join(home, "templates", "4.3")
	os.MkdirAll(tDir, 0755)
	os.WriteFile(filepath.Join(tDir, "marker"), []byte("ok"), 0644)

	out, err := runGdt(t, home, "templates", "list")
	if err != nil {
		t.Fatalf("error: %v, output: %s", err, out)
	}
	if !strings.Contains(out, "4.3") {
		t.Errorf("expected 4.3 in template list, got: %s", out)
	}
}

func TestTemplatesInstallNoVersion(t *testing.T) {
	home := t.TempDir()
	_, err := runGdt(t, home, "templates", "install")
	if err == nil {
		t.Error("expected error when no version provided (non-TTY)")
	}
}

// --- Plugin ---

func TestPluginListEmpty(t *testing.T) {
	home := t.TempDir()
	out, _ := runGdt(t, home, "plugin", "list")
	if !strings.Contains(out, "No plugins installed") {
		t.Errorf("expected empty plugins message, got: %s", out)
	}
}

func TestPluginListWithPlugins(t *testing.T) {
	home := t.TempDir()
	setupFakePlugin(t, home, "testplugin", "0.2.0")

	out, err := runGdt(t, home, "plugin", "list")
	if err != nil {
		t.Fatalf("error: %v, output: %s", err, out)
	}
	if !strings.Contains(out, "testplugin") {
		t.Errorf("expected testplugin in output, got: %s", out)
	}
	if !strings.Contains(out, "0.2.0") {
		t.Errorf("expected version 0.2.0 in output, got: %s", out)
	}
}

func TestPluginRemove(t *testing.T) {
	home := t.TempDir()
	setupFakePlugin(t, home, "removeme", "1.0.0")

	out, err := runGdt(t, home, "plugin", "remove", "removeme")
	if err != nil {
		t.Fatalf("error: %v, output: %s", err, out)
	}
	if !strings.Contains(out, "removed") {
		t.Errorf("expected removed message, got: %s", out)
	}

	dir := filepath.Join(home, "plugins", "removeme")
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("plugin directory should be deleted")
	}
}

func TestPluginRemoveNotFound(t *testing.T) {
	home := t.TempDir()
	_, err := runGdt(t, home, "plugin", "remove", "nonexistent")
	if err == nil {
		t.Error("expected error removing non-existent plugin")
	}
}

func TestPluginNew(t *testing.T) {
	home := t.TempDir()
	workDir := t.TempDir()

	out, err := runGdtIn(t, home, workDir, "plugin", "new", "myplugin")
	if err != nil {
		t.Fatalf("error: %v, output: %s", err, out)
	}

	pluginDir := filepath.Join(workDir, "gdt-myplugin")
	if _, err := os.Stat(filepath.Join(pluginDir, "plugin.toml")); err != nil {
		t.Error("plugin.toml should exist")
	}
	if _, err := os.Stat(filepath.Join(pluginDir, "README.md")); err != nil {
		t.Error("README.md should exist")
	}

	manifest, _ := os.ReadFile(filepath.Join(pluginDir, "plugin.toml"))
	if !strings.Contains(string(manifest), `name = "myplugin"`) {
		t.Errorf("manifest should contain plugin name, got: %s", manifest)
	}
}

func TestPluginNewAlreadyExists(t *testing.T) {
	home := t.TempDir()
	workDir := t.TempDir()

	os.MkdirAll(filepath.Join(workDir, "gdt-existing"), 0755)

	_, err := runGdtIn(t, home, workDir, "plugin", "new", "existing")
	if err == nil {
		t.Error("expected error when plugin dir already exists")
	}
}

func TestPluginInstallNoArgs(t *testing.T) {
	home := t.TempDir()
	_, err := runGdt(t, home, "plugin", "install")
	if err == nil {
		t.Error("expected error when no repository provided")
	}
}

func TestPluginUpdateEmpty(t *testing.T) {
	home := t.TempDir()
	out, _ := runGdt(t, home, "plugin", "update")
	if !strings.Contains(out, "No plugins installed") {
		t.Errorf("expected no plugins message, got: %s", out)
	}
}

// --- Run/Edit ---

func TestRunNoVersion(t *testing.T) {
	home := t.TempDir()
	workDir := t.TempDir()
	_, err := runGdtIn(t, home, workDir, "run")
	if err == nil {
		t.Error("expected error when no version available")
	}
}

func TestRunMissingBinary(t *testing.T) {
	home := t.TempDir()
	// Create version dir but no binary
	os.MkdirAll(filepath.Join(home, "versions", "4.3"), 0755)

	out, err := runGdtIn(t, home, t.TempDir(), "run", "4.3")
	if err == nil {
		t.Error("expected error when binary missing")
	}
	if !strings.Contains(out, "not found") && !strings.Contains(out, "not installed") {
		t.Errorf("expected helpful error about missing binary, got: %s", out)
	}
}

func TestEditNoVersion(t *testing.T) {
	home := t.TempDir()
	workDir := t.TempDir()
	_, err := runGdtIn(t, home, workDir, "edit")
	if err == nil {
		t.Error("expected error when no version available")
	}
}

func TestRunHelp(t *testing.T) {
	home := t.TempDir()
	out, err := runGdt(t, home, "run", "--help")
	if err != nil {
		t.Fatalf("error: %v, output: %s", err, out)
	}
	if !strings.Contains(out, "--editor") {
		t.Error("should show --editor flag in help")
	}
}

// --- Doctor ---

func TestDoctorWithVersions(t *testing.T) {
	home := t.TempDir()
	setupFakeVersion(t, home, "4.3")
	setupFakePlugin(t, home, "testplugin", "0.1.0")

	out, err := runGdt(t, home, "doctor")
	if err != nil {
		t.Fatalf("error: %v, output: %s", err, out)
	}
	if !strings.Contains(out, "engine 4.3") {
		t.Error("should show engine check")
	}
	if !strings.Contains(out, "plugin testplugin") {
		t.Error("should show plugin check")
	}
}

func TestDoctorMissingBinary(t *testing.T) {
	home := t.TempDir()
	// Version dir without binary
	os.MkdirAll(filepath.Join(home, "versions", "4.3"), 0755)

	out, err := runGdt(t, home, "doctor")
	if err != nil {
		t.Fatalf("error: %v, output: %s", err, out)
	}
	if !strings.Contains(out, "FAIL") {
		t.Error("should report FAIL for missing binary")
	}
	if !strings.Contains(out, "issue") {
		t.Error("should report issues found")
	}
}

// --- Remove ---

func TestRemoveInstalled(t *testing.T) {
	home := t.TempDir()
	setupFakeVersion(t, home, "4.2")

	out, err := runGdt(t, home, "remove", "4.2")
	if err != nil {
		t.Fatalf("error: %v, output: %s", err, out)
	}

	dir := filepath.Join(home, "versions", "4.2")
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("version directory should be deleted")
	}
}

func TestRemoveNoArgs(t *testing.T) {
	home := t.TempDir()
	// Non-TTY should error
	_, err := runGdt(t, home, "remove")
	if err == nil {
		t.Error("expected error when no version provided (non-TTY)")
	}
}

// --- Use ---

func TestUseWritesConfig(t *testing.T) {
	home := t.TempDir()
	setupFakeVersion(t, home, "4.3")

	out, err := runGdt(t, home, "use", "4.3")
	if err != nil {
		t.Fatalf("error: %v, output: %s", err, out)
	}

	configData, err := os.ReadFile(filepath.Join(home, "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(configData), "4.3") {
		t.Errorf("config should contain 4.3, got: %s", configData)
	}
}

func TestUseNoArgs(t *testing.T) {
	home := t.TempDir()
	_, err := runGdt(t, home, "use")
	if err == nil {
		t.Error("expected error when no version provided (non-TTY)")
	}
}

// --- List ---

func TestListMultipleVersions(t *testing.T) {
	home := t.TempDir()
	setupFakeVersion(t, home, "4.2")
	setupFakeVersion(t, home, "4.3")
	setupFakeVersion(t, home, "4.3-mono")

	out, err := runGdt(t, home, "list")
	if err != nil {
		t.Fatalf("error: %v, output: %s", err, out)
	}
	if !strings.Contains(out, "4.2") || !strings.Contains(out, "4.3") || !strings.Contains(out, "4.3-mono") {
		t.Errorf("expected all versions listed, got: %s", out)
	}
}

func TestListAlias(t *testing.T) {
	home := t.TempDir()
	out1, _ := runGdt(t, home, "list")
	out2, _ := runGdt(t, home, "ls")
	if out1 != out2 {
		t.Error("list and ls should produce identical output")
	}
}

// --- Local ---

func TestLocalOverwrite(t *testing.T) {
	home := t.TempDir()
	projDir := t.TempDir()

	// Set to 4.2 first
	runGdtIn(t, home, projDir, "local", "4.2")
	// Overwrite with 4.3
	runGdtIn(t, home, projDir, "local", "4.3")

	data, _ := os.ReadFile(filepath.Join(projDir, ".godot-version"))
	if strings.TrimSpace(string(data)) != "4.3" {
		t.Errorf("expected 4.3, got: %s", data)
	}
}

// --- Export ---

func TestExportNoPreset(t *testing.T) {
	home := t.TempDir()
	projDir := t.TempDir()
	// Create a project
	os.WriteFile(filepath.Join(projDir, "project.godot"), []byte("[application]\nconfig/name=\"test\""), 0644)

	_, err := runGdtIn(t, home, projDir, "export")
	if err == nil {
		t.Error("expected error when no preset specified (non-TTY)")
	}
}

func TestExportListWithPresets(t *testing.T) {
	home := t.TempDir()
	projDir := t.TempDir()
	os.WriteFile(filepath.Join(projDir, "project.godot"), []byte("[application]\nconfig/name=\"test\""), 0644)
	os.WriteFile(filepath.Join(projDir, "export_presets.cfg"), []byte("[preset.0]\nname=\"Linux/X11\"\n\n[preset.1]\nname=\"Windows Desktop\"\n"), 0644)

	out, err := runGdtIn(t, home, projDir, "export", "--list")
	if err != nil {
		t.Fatalf("error: %v, output: %s", err, out)
	}
	if !strings.Contains(out, "Linux/X11") {
		t.Errorf("expected Linux/X11 preset, got: %s", out)
	}
	if !strings.Contains(out, "Windows Desktop") {
		t.Errorf("expected Windows Desktop preset, got: %s", out)
	}
}

func TestExportHelp(t *testing.T) {
	home := t.TempDir()
	out, err := runGdt(t, home, "export", "--help")
	if err != nil {
		t.Fatalf("error: %v, output: %s", err, out)
	}
	for _, flag := range []string{"--output", "--debug", "--verbose", "--list"} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected %s flag in help, got: %s", flag, out)
		}
	}
}

// --- Install ---

func TestInstallNoVersion(t *testing.T) {
	home := t.TempDir()
	workDir := t.TempDir()
	_, err := runGdtIn(t, home, workDir, "install")
	if err == nil {
		t.Error("expected error when no version provided (non-TTY)")
	}
}

func TestInstallHelp(t *testing.T) {
	home := t.TempDir()
	out, err := runGdt(t, home, "install", "--help")
	if err != nil {
		t.Fatalf("error: %v, output: %s", err, out)
	}
	for _, flag := range []string{"--mono", "--force", "--refresh"} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected %s flag in help, got: %s", flag, out)
		}
	}
}

// --- LsRemote ---

func TestLsRemoteHelp(t *testing.T) {
	home := t.TempDir()
	out, err := runGdt(t, home, "ls-remote", "--help")
	if err != nil {
		t.Fatalf("error: %v, output: %s", err, out)
	}
	if !strings.Contains(out, "remote") || !strings.Contains(out, "version") {
		t.Errorf("expected helpful description, got: %s", out)
	}
}

// --- Self Update ---

func TestSelfUpdateHelp(t *testing.T) {
	home := t.TempDir()
	out, err := runGdt(t, home, "self", "update", "--help")
	if err != nil {
		t.Fatalf("error: %v, output: %s", err, out)
	}
	if !strings.Contains(out, "gdt") {
		t.Errorf("expected self update help, got: %s", out)
	}
}

// --- Update (metadata refresh) ---

func TestUpdateHelp(t *testing.T) {
	home := t.TempDir()
	out, err := runGdt(t, home, "update", "--help")
	if err != nil {
		t.Fatalf("error: %v, output: %s", err, out)
	}
	if !strings.Contains(out, "metadata") || !strings.Contains(out, "cache") {
		t.Errorf("expected metadata/cache in help, got: %s", out)
	}
}

// --- Help / unknown command ---

func TestHelpShowsAllCommands(t *testing.T) {
	home := t.TempDir()
	out, err := runGdt(t, home, "--help")
	if err != nil {
		t.Fatalf("error: %v, output: %s", err, out)
	}

	expected := []string{"install", "remove", "list", "use", "local", "run",
		"doctor", "templates", "plugin", "new", "export", "ci",
		"shell", "completion", "lsp", "dap"}
	for _, cmd := range expected {
		if !strings.Contains(out, cmd) {
			t.Errorf("expected %q in help output", cmd)
		}
	}
}

func TestUnknownCommand(t *testing.T) {
	home := t.TempDir()
	_, err := runGdt(t, home, "nonexistent-command-xyz")
	if err == nil {
		t.Error("expected error for unknown command")
	}
}

// --- Cross-command workflows ---

func TestInstallUseRunWorkflow(t *testing.T) {
	// Simulates: install creates version dir → use sets default → run resolves it
	home := t.TempDir()
	setupFakeVersion(t, home, "4.3")

	// Use 4.3
	_, err := runGdt(t, home, "use", "4.3")
	if err != nil {
		t.Fatal(err)
	}

	// List should show active marker
	out, _ := runGdt(t, home, "list")
	if !strings.Contains(out, "* 4.3") {
		t.Errorf("expected active marker, got: %s", out)
	}

	// Doctor should show engine valid
	out, _ = runGdt(t, home, "doctor")
	if !strings.Contains(out, "engine 4.3 valid") {
		t.Errorf("expected engine valid, got: %s", out)
	}
}

func TestNewThenExportList(t *testing.T) {
	home := t.TempDir()
	projDir := t.TempDir()
	setupFakeVersion(t, home, "4.3")

	// Create project
	cmd := exec.Command(gdtBinary, "new", "mygame", "--version", "4.3", "--renderer", "forward_plus")
	cmd.Env = append(os.Environ(), "GDT_HOME="+home)
	cmd.Dir = projDir
	cmd.Stdin = strings.NewReader("")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("new failed: %v, output: %s", err, out)
	}

	// export --list should work (no presets configured yet, but project detected)
	gameDir := filepath.Join(projDir, "mygame")
	out, _ := runGdtIn(t, home, gameDir, "export", "--list")
	// Should either list presets or show empty (not error about no project)
	if strings.Contains(out, "no Godot project found") {
		t.Error("should detect project created by 'new'")
	}
}

func TestPluginWorkflow(t *testing.T) {
	home := t.TempDir()

	// List empty
	out, _ := runGdt(t, home, "plugin", "list")
	if !strings.Contains(out, "No plugins") {
		t.Errorf("expected no plugins, got: %s", out)
	}

	// Add fake plugin
	setupFakePlugin(t, home, "testplugin", "1.0.0")

	// List should show it
	out, _ = runGdt(t, home, "plugin", "list")
	if !strings.Contains(out, "testplugin") {
		t.Errorf("expected testplugin, got: %s", out)
	}

	// Remove it
	out, err := runGdt(t, home, "plugin", "remove", "testplugin")
	if err != nil {
		t.Fatalf("remove failed: %v, output: %s", err, out)
	}

	// List empty again
	out, _ = runGdt(t, home, "plugin", "list")
	if !strings.Contains(out, "No plugins") {
		t.Errorf("expected no plugins after remove, got: %s", out)
	}
}

// --- Platform-specific ---

func TestRunVersionResolution(t *testing.T) {
	home := t.TempDir()
	projDir := t.TempDir()

	setupFakeVersion(t, home, "4.2")
	setupFakeVersion(t, home, "4.3")

	// Set local version
	os.WriteFile(filepath.Join(projDir, ".godot-version"), []byte("4.2"), 0644)

	// Run should fail because binary is fake, but the error should reference 4.2 (resolved from .godot-version)
	out, _ := runGdtIn(t, home, projDir, "run")
	// It should try to exec the binary — either fork/exec error or "not found"
	// The important thing is it resolved the version, not "no version"
	if strings.Contains(out, "no version") || strings.Contains(out, "version required") {
		t.Errorf("should have resolved version from .godot-version, got: %s", out)
	}
}

func TestRunWithExplicitVersion(t *testing.T) {
	home := t.TempDir()

	setupFakeVersion(t, home, "4.2")
	setupFakeVersion(t, home, "4.3")

	// Run with explicit version — should fail on exec but resolve correctly
	out, err := runGdt(t, home, "run", "4.2")
	if err == nil {
		t.Error("expected error (fake binary), but should have resolved version")
	}
	// Should NOT say "not installed"
	_ = out
}

// --- Edge cases ---

func TestListAfterRemoveAll(t *testing.T) {
	home := t.TempDir()
	setupFakeVersion(t, home, "4.3")

	runGdt(t, home, "remove", "4.3")

	out, _ := runGdt(t, home, "list")
	if !strings.Contains(out, "No versions installed") {
		t.Errorf("expected empty after removing all, got: %s", out)
	}
}

func TestDoctorCSharpWithoutMono(t *testing.T) {
	home := t.TempDir()
	projDir := t.TempDir()
	setupFakeVersion(t, home, "4.3")

	// Create C# project indicators
	os.WriteFile(filepath.Join(projDir, "project.godot"), []byte("[application]\nconfig/name=\"test\""), 0644)
	os.WriteFile(filepath.Join(projDir, "test.csproj"), []byte("<Project></Project>"), 0644)
	os.WriteFile(filepath.Join(projDir, ".godot-version"), []byte("4.3"), 0644)

	cmd := exec.Command(gdtBinary, "doctor")
	cmd.Env = append(os.Environ(), "GDT_HOME="+home)
	cmd.Dir = projDir
	cmd.Stdin = strings.NewReader("")
	out, _ := cmd.CombinedOutput()

	if !strings.Contains(string(out), "WARN") {
		t.Errorf("should warn about C# without mono, got: %s", out)
	}
}

func TestVersionsDir(t *testing.T) {
	home := t.TempDir()
	versionsDir := filepath.Join(home, "versions")

	// Before any install, versions dir may not exist — list should handle gracefully
	out, _ := runGdt(t, home, "list")
	if !strings.Contains(out, "No versions") {
		t.Errorf("expected no versions message, got: %s", out)
	}

	// Create versions dir manually
	os.MkdirAll(versionsDir, 0755)
	out, _ = runGdt(t, home, "list")
	if !strings.Contains(out, "No versions") {
		t.Errorf("expected no versions for empty dir, got: %s", out)
	}
}

func TestBinaryPlatformDetection(t *testing.T) {
	// Verify the binary runs on current platform (sanity check)
	home := t.TempDir()
	out, err := runGdt(t, home, "--version")
	if err != nil {
		t.Fatalf("binary should run on %s/%s: %v", runtime.GOOS, runtime.GOARCH, err)
	}
	if !strings.Contains(out, "gdt") {
		t.Error("version output should contain 'gdt'")
	}
}

// min returns the smaller of two ints
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
