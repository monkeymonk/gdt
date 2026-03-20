package plugins

import (
	"os"

	"github.com/BurntSushi/toml"
)

// PluginConfig holds arbitrary plugin configuration loaded from a TOML file.
type PluginConfig map[string]interface{}

// LoadConfig reads a TOML config file from path.
// Returns an empty config if the file does not exist.
func (s *Service) LoadConfig(path string) (PluginConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return PluginConfig{}, nil
		}
		return nil, err
	}
	var cfg PluginConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
