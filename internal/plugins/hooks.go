package plugins

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
)

// HookContext holds runtime context passed to hook scripts via environment.
type HookContext struct {
	ProjectRoot  string
	GodotVersion string
	EnginePath   string
}

// RunHooks discovers all plugins and executes any hook registered for the given event.
// Exit code 2 from a hook is treated as a fatal error; other non-zero exits are warnings.
func (s *Service) RunHooks(event HookEvent, ctx HookContext) error {
	plugins, err := s.Discover()
	if err != nil {
		return fmt.Errorf("discover plugins: %w", err)
	}

	env := BuildEnv(EnvContext{
		Home:         s.Dir,
		ProjectRoot:  ctx.ProjectRoot,
		GodotVersion: ctx.GodotVersion,
		EnginePath:   ctx.EnginePath,
	})
	env = append(env, "GDT_HOOK_EVENT="+string(event))
	// Inherit current environment.
	env = append(os.Environ(), env...)

	for _, p := range plugins {
		script := p.Manifest.HookFor(event)
		if script == "" {
			continue
		}

		slog.Debug("running hook", "plugin", p.Manifest.Name, "event", event)

		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd", "/C", script)
		} else {
			cmd = exec.Command("sh", "-c", script)
		}
		cmd.Dir = p.Dir
		cmd.Env = env
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if runErr := cmd.Run(); runErr != nil {
			var exitErr *exec.ExitError
			if ok := isExitError(runErr, &exitErr); ok && exitErr.ExitCode() == 2 {
				return fmt.Errorf("hook %s from plugin %s failed (exit 2): %w", event, p.Manifest.Name, runErr)
			}
			slog.Warn("hook returned non-zero exit", "plugin", p.Manifest.Name, "event", event, "error", runErr)
		}
	}

	return nil
}

// isExitError checks if err is an *exec.ExitError and assigns it.
func isExitError(err error, target **exec.ExitError) bool {
	if ee, ok := err.(*exec.ExitError); ok {
		*target = ee
		return true
	}
	return false
}
