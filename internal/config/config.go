package config

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	DefaultVersion string   `toml:"default_version"`
	Mirrors        []string `toml:"mirrors,omitempty"`
	GodotAPI       string   `toml:"godot_api,omitempty"`
	SelfUpdateAPI  string   `toml:"selfupdate_api,omitempty"`
}

const (
	defaultGodotAPI      = "https://api.github.com/repos/godotengine/godot/releases"
	defaultSelfUpdateAPI = "https://api.github.com/repos/monkeymonk/gdt/releases/latest"
)

func (c *Config) GodotAPIURL() string {
	if c.GodotAPI != "" {
		return c.GodotAPI
	}
	return defaultGodotAPI
}

func (c *Config) SelfUpdateAPIURL() string {
	if c.SelfUpdateAPI != "" {
		return c.SelfUpdateAPI
	}
	return defaultSelfUpdateAPI
}

func Load(path string) (*Config, error) {
	cfg := &Config{}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func Save(path string, cfg *Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(cfg); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0644)
}
