package metadata

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

func fakeGitHubServer(t *testing.T) *httptest.Server {
	t.Helper()
	releases := []GitHubRelease{
		{
			TagName: "4.3-stable",
			Assets: []GitHubAsset{
				{Name: "Godot_v4.3-stable_linux.x86_64.zip", URL: "http://example.com/linux.zip"},
				{Name: "Godot_v4.3-stable_mono_linux_x86_64.zip", URL: "http://example.com/linux_mono.zip"},
				{Name: "Godot_v4.3-stable_macos.universal.zip", URL: "http://example.com/macos.zip"},
				{Name: "Godot_v4.3-stable_win64.exe.zip", URL: "http://example.com/win.zip"},
				{Name: "Godot_v4.3-stable_export_templates.tpz", URL: "http://example.com/templates.zip"},
				{Name: "SHA512-SUMS.txt", URL: "http://example.com/sha512.txt"},
			},
		},
		{
			TagName: "4.2.2-stable",
			Assets: []GitHubAsset{
				{Name: "Godot_v4.2.2-stable_linux.x86_64.zip", URL: "http://example.com/linux422.zip"},
			},
		},
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(releases)
	}))
}

func TestFetchReleases(t *testing.T) {
	srv := fakeGitHubServer(t)
	defer srv.Close()

	releases, err := FetchReleases(srv.URL, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(releases) != 2 {
		t.Errorf("expected 2 releases, got %d", len(releases))
	}
	if releases[0].Version != "4.3" {
		t.Errorf("version = %q, want %q", releases[0].Version, "4.3")
	}
}

func TestCacheSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "releases.json")

	cache := &Cache{
		UpdatedAt: time.Now(),
		Releases: []Release{
			{Version: "4.3", Stable: true},
		},
	}

	err := SaveCache(path, cache)
	if err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadCache(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Releases) != 1 {
		t.Errorf("expected 1 release, got %d", len(loaded.Releases))
	}
}

func TestCacheIsStale(t *testing.T) {
	fresh := &Cache{UpdatedAt: time.Now()}
	if fresh.IsStale() {
		t.Error("fresh cache should not be stale")
	}

	old := &Cache{UpdatedAt: time.Now().Add(-25 * time.Hour)}
	if !old.IsStale() {
		t.Error("25h old cache should be stale")
	}
}

func TestResolveVersion(t *testing.T) {
	releases := []Release{
		{Version: "4.3", Stable: true},
		{Version: "4.2.2", Stable: true},
		{Version: "4.2.1", Stable: true},
		{Version: "4.1.4", Stable: true},
	}

	tests := []struct {
		query string
		want  string
	}{
		{"4.3", "4.3"},
		{"4.2.2", "4.2.2"},
		{"4.2", "4.2.2"},
		{"4", "4.3"},
		{"latest", "4.3"},
		{"stable", "4.3"},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			r, err := ResolveVersion(releases, tt.query)
			if err != nil {
				t.Fatal(err)
			}
			if r.Version != tt.want {
				t.Errorf("version = %q, want %q", r.Version, tt.want)
			}
		})
	}
}

func TestResolveVersionNotFound(t *testing.T) {
	releases := []Release{{Version: "4.3", Stable: true}}
	_, err := ResolveVersion(releases, "3.0")
	if err == nil {
		t.Error("should error for unknown version")
	}
}
