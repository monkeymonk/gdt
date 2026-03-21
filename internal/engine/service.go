package engine

import (
	"path/filepath"

	"github.com/monkeymonk/gdt/internal/config"
	"github.com/monkeymonk/gdt/internal/platform"
)

type Service struct {
	Home     string
	Platform platform.Info
	Config   *config.Config
}

type InstallOpts struct {
	Mono    bool
	Force   bool
	Refresh bool
}

type InstallResult struct {
	Version      string
	VersionName  string
	ArtifactName string
	IsNew        bool
}

type ResolvedVersion struct {
	Version    string
	BinaryPath string
	Source     string // "file", "env", "config", "latest"
}

type InstalledVersion struct {
	Version   string
	IsDefault bool
}

func NewService(home string, plat platform.Info, cfg *config.Config) *Service {
	return &Service{Home: home, Platform: plat, Config: cfg}
}

func (s *Service) versionsDir() string  { return filepath.Join(s.Home, "versions") }
func (s *Service) templatesDir() string { return filepath.Join(s.Home, "templates") }
func (s *Service) cacheDir() string     { return filepath.Join(s.Home, "cache") }
func (s *Service) cachePath() string    { return filepath.Join(s.Home, "cache", "releases.json") }
