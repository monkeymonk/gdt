package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/monkeymonk/gdt/internal/metadata"
)

func loadMetadata(app *App, forceRefresh bool) ([]metadata.Release, error) {
	cachePath := app.CachePath()

	if !forceRefresh {
		cache, err := metadata.LoadCache(cachePath)
		if err == nil && !cache.IsStale() {
			return cache.Releases, nil
		}
	}

	apiURL := "https://api.github.com/repos/godotengine/godot/releases"
	token := os.Getenv("GDT_GITHUB_TOKEN")
	fmt.Fprintln(os.Stderr, "Fetching release metadata...")

	releases, err := metadata.FetchReleases(apiURL, token)
	if err != nil {
		return nil, err
	}

	cache := &metadata.Cache{
		UpdatedAt: time.Now(),
		Releases:  releases,
	}
	metadata.SaveCache(cachePath, cache)

	return releases, nil
}

func resolveEngineArtifact(release *metadata.Release, platformArtifact string, mono bool) string {
	prefix := fmt.Sprintf("Godot_v%s-stable_", release.Version)
	if mono {
		prefix = fmt.Sprintf("Godot_v%s-stable_mono_", release.Version)
	}

	for name := range release.Assets {
		if strings.HasPrefix(name, prefix) && strings.Contains(name, platformArtifact) && !strings.Contains(name, "export_templates") {
			return name
		}
	}
	return prefix + platformArtifact + ".zip"
}

func findChecksum(checksumFile string, artifactName string) string {
	data, err := os.ReadFile(checksumFile)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		parts := strings.Fields(line)
		if len(parts) == 2 && parts[1] == artifactName {
			return parts[0]
		}
	}
	return ""
}
