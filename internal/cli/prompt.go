package cli

import (
	"os"

	"github.com/charmbracelet/huh"
	"github.com/monkeymonk/gdt/internal/engine"
	"github.com/monkeymonk/gdt/internal/metadata"
	"github.com/monkeymonk/gdt/internal/plugins"
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

// promptInstalledTemplate prompts the user to select from installed templates.
func promptInstalledTemplate(app *App, title string) (string, error) {
	svc := engine.NewService(app.Home, app.Platform, app.Config)
	list, _ := svc.ListTemplates()
	if len(list) == 0 {
		return "", nil
	}

	var version string
	options := make([]huh.Option[string], len(list))
	for i, v := range list {
		options[i] = huh.NewOption(v, v)
	}

	err := huh.NewSelect[string]().
		Title(title).
		Options(options...).
		Value(&version).
		Run()
	return version, err
}

// promptInput prompts the user for a text value.
func promptInput(title string, placeholder string) (string, error) {
	var value string
	err := huh.NewInput().
		Title(title).
		Placeholder(placeholder).
		Value(&value).
		Run()
	return value, err
}

// promptInstalledPlugin prompts the user to select from installed plugins.
func promptInstalledPlugin(app *App, title string) (string, error) {
	svc := plugins.NewService(app.PluginsDir())
	pluginList, _ := svc.Discover()
	if len(pluginList) == 0 {
		return "", nil
	}

	var name string
	options := make([]huh.Option[string], len(pluginList))
	for i, p := range pluginList {
		options[i] = huh.NewOption(p.Manifest.Name, p.Manifest.Name)
	}

	err := huh.NewSelect[string]().
		Title(title).
		Options(options...).
		Value(&name).
		Run()
	return name, err
}

// isTTY returns true if stdin is a terminal (interactive mode possible).
func isTTY() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
