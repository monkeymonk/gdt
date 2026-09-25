package metadata

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func fakeGitHubServer(t *testing.T) *httptest.Server {
	t.Helper()
	releases := []githubRelease{
		{
			TagName: "4.3-stable",
			Assets: []githubAsset{
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
			Assets: []githubAsset{
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

func fakeGitHubServerScrambled(t *testing.T) *httptest.Server {
	t.Helper()
	// Deliberately not in newest-first order, and not alphabetical
	// order either ("4.1.0" < "4.10.0" < "4.2.2" lexicographically),
	// so this proves FetchReleases applies an explicit numeric sort
	// rather than relying on request or string order.
	releases := []githubRelease{
		{
			TagName: "4.1.0-stable",
			Assets: []githubAsset{
				{Name: "Godot_v4.1.0-stable_linux.x86_64.zip", URL: "http://example.com/linux410.zip"},
			},
		},
		{
			TagName: "4.10.0-stable",
			Assets: []githubAsset{
				{Name: "Godot_v4.10.0-stable_linux.x86_64.zip", URL: "http://example.com/linux4100.zip"},
			},
		},
		{
			TagName: "4.2.2-stable",
			Assets: []githubAsset{
				{Name: "Godot_v4.2.2-stable_linux.x86_64.zip", URL: "http://example.com/linux422.zip"},
			},
		},
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(releases)
	}))
}

func TestFetchReleasesSortedDescending(t *testing.T) {
	srv := fakeGitHubServerScrambled(t)
	defer srv.Close()

	releases, err := FetchReleases(srv.URL, "")
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"4.10.0", "4.2.2", "4.1.0"}
	if len(releases) != len(want) {
		t.Fatalf("expected %d releases, got %d", len(want), len(releases))
	}
	for i, w := range want {
		if releases[i].Version != w {
			t.Errorf("releases[%d].Version = %q, want %q", i, releases[i].Version, w)
		}
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want int // sign only: -1 negative, 0 zero, 1 positive
	}{
		{"standard ordering", "4.3", "4.2", 1},
		{"multi-digit segment is numeric not lexicographic", "4.10", "4.9", 1},
		{"reverse of multi-digit segment", "4.9", "4.10", -1},
		{"extra patch segment wins", "4.3.1", "4.3", 1},
		{"missing trailing segment treated as zero", "4.3", "4.3.0", 0},
		{"mono suffix stripped before comparing", "4.3-mono", "4.3", 0},
		{"mono suffix stripped on both sides", "4.3-mono", "4.3.0-mono", 0},
		{"equal versions", "4.3.0", "4.3.0", 0},
		{"older loses", "4.2", "4.3", -1},
		{"non-numeric segment parses as zero, does not panic", "4.x", "4.0", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CompareVersions(tt.a, tt.b)
			switch {
			case tt.want > 0 && got <= 0:
				t.Errorf("CompareVersions(%q, %q) = %d, want positive", tt.a, tt.b, got)
			case tt.want < 0 && got >= 0:
				t.Errorf("CompareVersions(%q, %q) = %d, want negative", tt.a, tt.b, got)
			case tt.want == 0 && got != 0:
				t.Errorf("CompareVersions(%q, %q) = %d, want 0", tt.a, tt.b, got)
			}
		})
	}
}

func TestEnsureCacheReturnsReleasesOnCacheWriteFailure(t *testing.T) {
	srv := fakeGitHubServer(t)
	defer srv.Close()

	dir := t.TempDir()
	// Make the cache path's parent a file, not a directory, so MkdirAll
	// (and thus SaveCache) fails while FetchReleases still succeeds.
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("not a dir"), 0644); err != nil {
		t.Fatal(err)
	}
	cachePath := filepath.Join(blocker, "releases.json")

	releases, err := EnsureCache(cachePath, srv.URL, "", true)
	if err != nil {
		t.Fatalf("EnsureCache returned error despite successful fetch: %v", err)
	}
	if len(releases) != 2 {
		t.Errorf("expected 2 releases, got %d", len(releases))
	}
}

// TestEnsureCacheLogsCacheWriteFailureWhenDebugEnabled proves the
// SaveCache failure above is observable via slog.Warn when GDT_DEBUG=1,
// matching the internal/engine/desktop.go best-effort-logging precedent.
func TestEnsureCacheLogsCacheWriteFailureWhenDebugEnabled(t *testing.T) {
	srv := fakeGitHubServer(t)
	defer srv.Close()

	t.Setenv("GDT_DEBUG", "1")

	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("not a dir"), 0644); err != nil {
		t.Fatal(err)
	}
	cachePath := filepath.Join(blocker, "releases.json")

	var logBuf bytes.Buffer
	origLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logBuf, nil)))
	t.Cleanup(func() { slog.SetDefault(origLogger) })

	if _, err := EnsureCache(cachePath, srv.URL, "", true); err != nil {
		t.Fatalf("EnsureCache returned error despite successful fetch: %v", err)
	}

	output := logBuf.String()
	if output == "" {
		t.Fatal("expected EnsureCache to log a warning for the cache-write failure, got no log output")
	}
	if !bytes.Contains(logBuf.Bytes(), []byte("save cache failed")) {
		t.Errorf("expected log output to mention the cache-write failure, got: %s", output)
	}
}

// TestEnsureCacheNoLogWhenDebugDisabled proves the logging added above
// stays silent by default (GDT_DEBUG unset).
func TestEnsureCacheNoLogWhenDebugDisabled(t *testing.T) {
	srv := fakeGitHubServer(t)
	defer srv.Close()

	t.Setenv("GDT_DEBUG", "")

	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("not a dir"), 0644); err != nil {
		t.Fatal(err)
	}
	cachePath := filepath.Join(blocker, "releases.json")

	var logBuf bytes.Buffer
	origLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logBuf, nil)))
	t.Cleanup(func() { slog.SetDefault(origLogger) })

	if _, err := EnsureCache(cachePath, srv.URL, "", true); err != nil {
		t.Fatalf("EnsureCache returned error despite successful fetch: %v", err)
	}

	if logBuf.Len() != 0 {
		t.Errorf("expected no log output with GDT_DEBUG unset, got: %s", logBuf.String())
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

// TestEnsureCache_SortsPreExistingUnsortedCache proves that EnsureCache
// sorts its result even when returning data straight from a fresh
// on-disk cache written before this ordering existed (or by any other
// path that doesn't itself sort) — the cache-hit branch at line ~134
// must not bypass sortReleasesDescending. A prior version of this fix
// only sorted inside FetchReleases, so a pre-existing cache file (the
// common case for any user who already ran `gdt ls-remote` before this
// change) stayed silently unsorted until the cache next expired.
func TestEnsureCache_SortsPreExistingUnsortedCache(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "releases.json")

	// Deliberately unsorted, fresh (not stale) cache — simulates a cache
	// file written before descending order was introduced.
	cache := &Cache{
		UpdatedAt: time.Now(),
		Releases: []Release{
			{Version: "3.6.3", Stable: true},
			{Version: "4.7.2", Stable: true},
			{Version: "4.6.1", Stable: true},
			{Version: "4.2", Stable: true},
		},
	}
	if err := SaveCache(path, cache); err != nil {
		t.Fatal(err)
	}

	releases, err := EnsureCache(path, "http://unused.invalid", "", false)
	if err != nil {
		t.Fatalf("EnsureCache returned error for a fresh cache hit: %v", err)
	}

	want := []string{"4.7.2", "4.6.1", "4.2", "3.6.3"}
	if len(releases) != len(want) {
		t.Fatalf("expected %d releases, got %d: %v", len(want), len(releases), releases)
	}
	for i, r := range releases {
		if r.Version != want[i] {
			t.Errorf("index %d: expected %q, got %q (full order: %v)", i, want[i], r.Version, releases)
		}
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
