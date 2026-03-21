package cli

import (
	"os"
	"path/filepath"

	"github.com/monkeymonk/gdt/internal/config"
	"github.com/monkeymonk/gdt/internal/platform"
)

type App struct {
	Version    string
	Home       string
	Config     *config.Config
	ConfigPath string
	Platform   platform.Info
	Debug      bool
}

func NewApp(version string) (*App, error) {
	plat := platform.Detect()
	home := config.ResolveHome()

	configPath := filepath.Join(home, "config.toml")
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}

	debug := os.Getenv("GDT_DEBUG") == "1"

	return &App{
		Version:    version,
		Home:       home,
		Config:     cfg,
		ConfigPath: configPath,
		Platform:   plat,
		Debug:      debug,
	}, nil
}

func (a *App) VersionsDir() string {
	return filepath.Join(a.Home, "versions")
}

func (a *App) TemplatesDir() string {
	return filepath.Join(a.Home, "templates")
}

func (a *App) PluginsDir() string {
	return filepath.Join(a.Home, "plugins")
}

func (a *App) CacheDir() string {
	return filepath.Join(a.Home, "cache")
}

func (a *App) CachePath() string {
	return filepath.Join(a.CacheDir(), "releases.json")
}
