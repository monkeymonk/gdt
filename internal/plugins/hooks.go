package plugins

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// HookContext holds runtime context passed to hook scripts via environment.
type HookContext struct {
	ProjectRoot  string
	GodotVersion string
	EnginePath   string
}

// RunHooks discovers all plugins and executes hooks for the given event.
// V2 plugins (with [contributions]) use binary subcommand protocol.
// V1 plugins (with [hooks] shell strings) use legacy shell execution.
// Plugins run in alphabetical order by name.
func (s *Service) RunHooks(event HookEvent, ctx HookContext) error {
	plugins, err := s.Discover()
	if err != nil {
		return fmt.Errorf("discover plugins: %w", err)
	}

	// Sort alphabetically for deterministic order
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Manifest.Name < plugins[j].Manifest.Name
	})

	env := BuildEnv(EnvContext{
		Home:         s.Dir,
		ProjectRoot:  ctx.ProjectRoot,
		GodotVersion: ctx.GodotVersion,
		EnginePath:   ctx.EnginePath,
		HookEvent:    string(event),
	})
	env = append(os.Environ(), env...)

	for _, p := range plugins {
		if p.Manifest.Protocol >= 2 {
			if err := s.runV2Hook(p, event, env); err != nil {
				return err
			}
		} else {
			if err := s.runV1Hook(p, event, env); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Service) runV2Hook(p Plugin, event HookEvent, env []string) error {
	declared := false
	for _, h := range p.Manifest.Contributions.Hooks {
		if h == string(event) {
			declared = true
			break
		}
	}
	if !declared {
		return nil
	}

	binPath := filepath.Join(p.Dir, p.Manifest.Name)
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		slog.Warn("plugin declares hook but binary missing",
			"plugin", p.Manifest.Name, "event", event)
		return nil
	}

	slog.Debug("running v2 hook", "plugin", p.Manifest.Name, "event", event)

	out, err := RunPluginSubcommand(binPath, p.Dir, env, DefaultHookTimeout, "hook", string(event))

	var failures []string
	for _, r := range ParseStatusLines(out) {
		switch r.Status {
		case "WARN":
			slog.Warn("hook warning", "plugin", p.Manifest.Name, "message", r.Message)
		case "FAIL":
			slog.Error("hook failure", "plugin", p.Manifest.Name, "message", r.Message)
			failures = append(failures, r.Message)
		}
	}

	if err != nil {
		return fmt.Errorf("hook %s from plugin %s failed: %w", event, p.Manifest.Name, err)
	}
	if len(failures) > 0 {
		return fmt.Errorf("hook %s from plugin %s reported failures: %s", event, p.Manifest.Name, strings.Join(failures, "; "))
	}
	return nil
}

func (s *Service) runV1Hook(p Plugin, event HookEvent, env []string) error {
	script := p.Manifest.HookFor(event)
	if script == "" {
		return nil
	}

	slog.Debug("running v1 hook", "plugin", p.Manifest.Name, "event", event)

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
		if errors.As(runErr, &exitErr) && exitErr.ExitCode() == 2 {
			return fmt.Errorf("hook %s from plugin %s failed (exit 2): %w", event, p.Manifest.Name, runErr)
		}
		slog.Warn("v1 hook returned non-zero exit", "plugin", p.Manifest.Name, "event", event, "error", runErr)
	}
	return nil
}
