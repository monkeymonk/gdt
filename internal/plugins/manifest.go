package plugins

import "github.com/BurntSushi/toml"

type Manifest struct {
	Name        string   `toml:"name"`
	Version     string   `toml:"version"`
	Commands    []string `toml:"commands"`
	RequiresGdt string   `toml:"requires_gdt"`
	Description string   `toml:"description"`
}

func ParseManifest(data []byte) (*Manifest, error) {
	var m Manifest
	if err := toml.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
