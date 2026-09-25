package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/monkeymonk/gdt/internal/config"
	"github.com/monkeymonk/gdt/internal/platform"
)

// fakeLsRemoteServer serves a fake GitHub releases API response containing
// a single stable release, mirroring internal/metadata/metadata_test.go's
// fakeGitHubServer fixture shape.
func fakeLsRemoteServer(t *testing.T) *httptest.Server {
	t.Helper()
	type githubAsset struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	}
	type githubRelease struct {
		TagName string        `json:"tag_name"`
		Assets  []githubAsset `json:"assets"`
	}
	releases := []githubRelease{
		{
			TagName: "4.3-stable",
			Assets: []githubAsset{
				{Name: "Godot_v4.3-stable_linux.x86_64.zip", URL: "http://example.com/linux.zip"},
			},
		},
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(releases)
	}))
}

// TestLsRemote_PrintsReleasesWithStableLabel proves that releases fetched
// from the metadata cache are printed, and stable releases get the
// " stable" label suffix.
func TestLsRemote_PrintsReleasesWithStableLabel(t *testing.T) {
	srv := fakeLsRemoteServer(t)
	defer srv.Close()

	home := t.TempDir()
	app := &App{
		Home:     home,
		Config:   &config.Config{GodotAPI: srv.URL},
		Platform: platform.Info{OS: "linux", Arch: "amd64"},
	}

	cmd := newLsRemoteCmd(app)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	var runErr error
	output := captureStdout(t, func() {
		runErr = cmd.Execute()
	})
	if runErr != nil {
		t.Fatalf("expected nil error, got: %v", runErr)
	}
	if !strings.Contains(output, "4.3") {
		t.Errorf("expected output to contain release version 4.3, got:\n%s", output)
	}
	if !strings.Contains(output, "4.3 stable") {
		t.Errorf("expected stable release to carry the ' stable' label, got:\n%s", output)
	}
}

// TestLsRemote_EnsureCacheErrorPropagates proves that when the metadata API
// returns a non-200 status and no cache exists to fall back to, the command
// returns a non-nil error.
func TestLsRemote_EnsureCacheErrorPropagates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	home := t.TempDir()
	app := &App{
		Home:     home,
		Config:   &config.Config{GodotAPI: srv.URL},
		Platform: platform.Info{OS: "linux", Arch: "amd64"},
	}

	cmd := newLsRemoteCmd(app)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	var runErr error
	_ = captureStdout(t, func() {
		runErr = cmd.Execute()
	})
	if runErr == nil {
		t.Fatal("expected error when metadata API returns a non-200 status, got nil")
	}
}
