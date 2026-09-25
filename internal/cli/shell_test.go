package cli

import (
	"strings"
	"testing"

	"github.com/monkeymonk/gdt/internal/config"
	"github.com/monkeymonk/gdt/internal/platform"
)

// TestDetectShell proves detectShell classifies the SHELL environment
// variable into "fish", "zsh", or a "bash" default for everything else
// (including bash itself, dash, empty, and unrecognized values).
func TestDetectShell(t *testing.T) {
	cases := []struct {
		name  string
		shell string
		want  string
	}{
		{"fish path", "/usr/bin/fish", "fish"},
		{"fish substring", "/opt/homebrew/bin/fish", "fish"},
		{"zsh path", "/bin/zsh", "zsh"},
		{"zsh substring", "/usr/local/bin/zsh", "zsh"},
		{"bash path", "/bin/bash", "bash"},
		{"dash falls back to bash", "/bin/dash", "bash"},
		{"empty falls back to bash", "", "bash"},
		{"unrecognized shell falls back to bash", "/usr/bin/csh", "bash"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("SHELL", tc.shell)
			if got := detectShell(); got != tc.want {
				t.Errorf("detectShell() with SHELL=%q = %q, want %q", tc.shell, got, tc.want)
			}
		})
	}
}

// TestShellInit_PrintsFishOrDefaultExport proves that `gdt shell init`
// branches on the detected shell: fish gets a `fish_add_path` line, and
// every other shell gets a POSIX `export PATH=...` line.
func TestShellInit_PrintsFishOrDefaultExport(t *testing.T) {
	app := &App{Home: t.TempDir(), Config: &config.Config{}, Platform: platform.Info{OS: "linux", Arch: "amd64"}}

	cases := []struct {
		name        string
		shell       string
		wantContain string
		wantAbsent  string
	}{
		{"fish shell", "/usr/bin/fish", "fish_add_path ", "export PATH="},
		{"non-fish shell", "/bin/bash", "export PATH=", "fish_add_path "},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("SHELL", tc.shell)

			cmd := newShellCmd(app)
			cmd.SetArgs([]string{"init"})
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			var runErr error
			output := captureStdout(t, func() {
				runErr = cmd.Execute()
			})
			if runErr != nil {
				t.Fatalf("shell init returned error: %v", runErr)
			}
			if !strings.Contains(output, tc.wantContain) {
				t.Errorf("expected output to contain %q, got: %q", tc.wantContain, output)
			}
			if strings.Contains(output, tc.wantAbsent) {
				t.Errorf("expected output NOT to contain %q, got: %q", tc.wantAbsent, output)
			}
		})
	}
}

// Note: the os.Executable() failure branch (added to keep binDir as "."
// instead of silently blanking it) has no portable, non-fragile way to
// force os.Executable() to fail from a Go test — it only errors on
// platform-specific edge cases (e.g. the executable being deleted out from
// under a running process, or unreadable /proc/self/exe on Linux), none of
// which can be triggered deterministically without root or OS-specific
// process surgery. Matching round 001's precedent for similar
// unforceable-without-root paths, this branch is left covered by code
// review rather than an automated test.
