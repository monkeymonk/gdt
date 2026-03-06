package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var gdtBinary string

func TestMain(m *testing.M) {
	dir, _ := os.MkdirTemp("", "gdt-test-*")
	gdtBinary = filepath.Join(dir, "gdt")
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
	out, err := cmd.CombinedOutput()
	return string(out), err
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
