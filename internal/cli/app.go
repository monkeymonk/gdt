package cli

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/monkeymonk/gdt/internal/config"
	"github.com/monkeymonk/gdt/internal/engine"
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
	version = strings.TrimPrefix(version, "v")
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

func (a *App) PluginsDir() string {
	return filepath.Join(a.Home, "plugins")
}

func (a *App) EngineSvc() *engine.Service {
	return engine.NewService(a.Home, a.Platform, a.Config)
}
