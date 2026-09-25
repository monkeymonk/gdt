package plugins

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initGitFixture creates a minimal local git repository at dir containing a
// plugin.toml manifest with the given content, committed on the default
// branch, suitable for use as a `git clone` source in tests.
func initGitFixture(t *testing.T, dir, manifest string) {
	t.Helper()

	runGit := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=gdt-test",
			"GIT_AUTHOR_EMAIL=gdt-test@example.com",
			"GIT_COMMITTER_NAME=gdt-test",
			"GIT_COMMITTER_EMAIL=gdt-test@example.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	runGit("init", "-q")
	if err := os.WriteFile(filepath.Join(dir, ManifestFile), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit("add", "-A")
	runGit("commit", "-q", "-m", "initial", "--no-gpg-sign")
}

// TestInstallRejectsUnsatisfiedRequiresGdt proves Service.Install rejects a
// plugin whose manifest declares a requires_gdt constraint the running gdt
// version does not satisfy, and that the destination directory is fully
// cleaned up (os.RemoveAll ran) rather than left as a partial clone.
//
// Service.Install rewrites any repo argument lacking an "http" prefix into
// a github.com URL, so this fixture's directory is deliberately named with
// an "http" prefix and cloned via a relative path (after chdir into its
// parent) to exercise the real `git clone` code path without a network
// dependency.
func TestInstallRejectsUnsatisfiedRequiresGdt(t *testing.T) {
	base := t.TempDir()
	const repoName = "http-fixture"
	repoDir := filepath.Join(base, repoName)
	initGitFixture(t, repoDir, `
name = "needs-newer-gdt"
version = "1.0.0"
protocol = 2
commands = ["needs-newer-gdt"]
requires_gdt = ">=1.0"
`)

	pluginsDir := t.TempDir()
	svc := NewService(pluginsDir)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(base); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(cwd)

	_, err = svc.Install(repoName, "0.3.0")
	if err == nil {
		t.Fatal("Install() error = nil, want error rejecting unsatisfied requires_gdt")
	}
	const wantSubstr = "requires gdt >=1.0, running 0.3.0"
	if !strings.Contains(err.Error(), wantSubstr) {
		t.Errorf("Install() error = %q, want substring %q", err.Error(), wantSubstr)
	}

	destDir := filepath.Join(pluginsDir, repoName)
	if _, statErr := os.Stat(destDir); !os.IsNotExist(statErr) {
		t.Errorf("destDir %s = exists (stat err %v), want removed after rejected install", destDir, statErr)
	}
}
