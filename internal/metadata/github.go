package metadata

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type LatestRelease struct {
	TagName string
	Assets  map[string]string // name → download URL
}

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

func FetchLatestRelease(url string, token string) (*LatestRelease, error) {
	req, err := http.NewRequest("GET", url, nil)
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
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}
	var rel githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	assets := make(map[string]string, len(rel.Assets))
	for _, a := range rel.Assets {
		assets[a.Name] = a.URL
	}
	return &LatestRelease{TagName: rel.TagName, Assets: assets}, nil
}
