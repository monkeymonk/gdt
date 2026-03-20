package cli

import (
	"os"

	"github.com/charmbracelet/huh"
	"github.com/monkeymonk/gdt/internal/engine"
	"github.com/monkeymonk/gdt/internal/metadata"
)

// promptVersion prompts the user to select from installed versions.
// Returns empty string if no versions are installed.
func promptVersion(app *App, title string) (string, error) {
	svc := engine.NewService(app.Home, app.Platform, app.Config)
	installed, _ := svc.ListVersionStrings()
	if len(installed) == 0 {
		return "", nil
	}

	var version string
	options := make([]huh.Option[string], len(installed))
	for i, v := range installed {
		options[i] = huh.NewOption(v, v)
	}

	err := huh.NewSelect[string]().
		Title(title).
		Options(options...).
		Value(&version).
		Run()
	return version, err
}

// promptConfirm asks a yes/no question. Returns false on error.
func promptConfirm(title string) (bool, error) {
	var confirm bool
	err := huh.NewConfirm().
		Title(title).
		Value(&confirm).
		Run()
	return confirm, err
}

// promptRemoteVersion prompts the user to select from available remote releases.
func promptRemoteVersion(releases []metadata.Release) (string, error) {
	options := make([]huh.Option[string], 0, len(releases))
	for _, r := range releases {
		label := r.Version
		if r.Stable {
			label += " (stable)"
		}
		options = append(options, huh.NewOption(label, r.Version))
	}

	var version string
	err := huh.NewSelect[string]().
		Title("Engine version").
		Options(options...).
		Value(&version).
		Run()
	return version, err
}

// promptPreset prompts the user to select from available export presets.
func promptPreset(presets []string) (string, error) {
	options := make([]huh.Option[string], len(presets))
	for i, p := range presets {
		options[i] = huh.NewOption(p, p)
	}

	var preset string
	err := huh.NewSelect[string]().
		Title("Export preset").
		Options(options...).
		Value(&preset).
		Run()
	return preset, err
}

// isTTY returns true if stdin is a terminal (interactive mode possible).
func isTTY() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
