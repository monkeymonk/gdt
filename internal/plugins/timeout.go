package plugins

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"syscall"
	"time"
)

const DefaultHookTimeout = 60 * time.Second

// RunPluginSubcommand executes a plugin binary with the given subcommand and args.
// It captures stdout and returns it. Stderr goes to the process stderr.
// On timeout, sends SIGKILL to the process group.
func RunPluginSubcommand(binPath string, workDir string, env []string, timeout time.Duration, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, binPath, args...)
	cmd.Dir = workDir
	if len(env) > 0 {
		cmd.Env = env
	}

	// Set process group so we can kill the whole tree on timeout.
	if runtime.GOOS != "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		cmd.Cancel = func() error {
			if cmd.Process != nil {
				return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			}
			return nil
		}
	}

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()

	if ctx.Err() == context.DeadlineExceeded {
		return stdout.String(), fmt.Errorf("plugin timed out after %s", timeout)
	}

	if err != nil {
		return stdout.String(), fmt.Errorf("plugin exited with error: %w", err)
	}

	return stdout.String(), nil
}
