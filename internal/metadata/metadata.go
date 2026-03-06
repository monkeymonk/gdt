package metadata

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const CacheTTL = 24 * time.Hour

type Release struct {
	Version string            `json:"version"`
	Tag     string            `json:"tag"`
	Stable  bool              `json:"stable"`
	Assets  map[string]string `json:"assets"`
}

type Cache struct {
	UpdatedAt time.Time `json:"updated_at"`
	Releases  []Release `json:"releases"`
}

func (c *Cache) IsStale() bool {
	return time.Since(c.UpdatedAt) > CacheTTL
}

type ghRelease struct {
	TagName string    `json:"tag_name"`
	Assets  []ghAsset `json:"assets"`
}

type ghAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func FetchReleases(apiURL string, token string) ([]Release, error) {
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error %d: %s", resp.StatusCode, string(body))
	}

	var ghReleases []ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&ghReleases); err != nil {
		return nil, err
	}

	var releases []Release
	for _, ghr := range ghReleases {
		r := parseRelease(ghr)
		if r != nil {
			releases = append(releases, *r)
		}
	}
	return releases, nil
}

func parseRelease(ghr ghRelease) *Release {
	tag := ghr.TagName
	if !strings.HasSuffix(tag, "-stable") {
		return nil
	}
	version := strings.TrimSuffix(tag, "-stable")

	assets := make(map[string]string)
	for _, a := range ghr.Assets {
		assets[a.Name] = a.BrowserDownloadURL
	}

	return &Release{
		Version: version,
		Tag:     tag,
		Stable:  true,
		Assets:  assets,
	}
}

func LoadCache(path string) (*Cache, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cache Cache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}
	return &cache, nil
}

func SaveCache(path string, cache *Cache) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func ResolveVersion(releases []Release, query string) (*Release, error) {
	if query == "latest" || query == "stable" {
		for i := range releases {
			if releases[i].Stable {
				return &releases[i], nil
			}
		}
		return nil, fmt.Errorf("no stable version found")
	}

	// Exact match
	for i := range releases {
		if releases[i].Version == query {
			return &releases[i], nil
		}
	}

	// Prefix match (4.2 -> 4.2.2, 4 -> 4.3)
	for i := range releases {
		if releases[i].Stable && strings.HasPrefix(releases[i].Version, query) {
			return &releases[i], nil
		}
	}

	return nil, fmt.Errorf("version %q not found\n\n  Run: gdt ls-remote", query)
}
