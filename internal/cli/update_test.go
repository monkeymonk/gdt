package cli

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/monkeymonk/gdt/internal/config"
	"github.com/monkeymonk/gdt/internal/platform"
)

// captureStderr runs fn while redirecting os.Stderr to a pipe and returns
// everything written to it, mirroring doctor_test.go's captureStdout.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	orig := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe(): %v", err)
	}
	os.Stderr = w
	t.Cleanup(func() { os.Stderr = orig })

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("closing pipe writer: %v", err)
	}
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("reading captured stderr: %v", err)
	}
	return string(out)
}

// fakeUpdateServer serves a fake GitHub releases API response containing a
// single stable release, mirroring internal/metadata/metadata_test.go's
// fakeGitHubServer fixture shape.
func fakeUpdateServer(t *testing.T) *httptest.Server {
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

// TestUpdate_SuccessMessageReportsReleaseCount proves that a successful
// metadata refresh prints a confirmation to stderr naming the number of
// releases fetched. The apiURL is fully overridable via
// app.Config.GodotAPI (config.Config.GodotAPIURL()), so this is hermetic:
// no hardcoded-URL test gap exists for this path.
func TestUpdate_SuccessMessageReportsReleaseCount(t *testing.T) {
	srv := fakeUpdateServer(t)
	defer srv.Close()

	home := t.TempDir()
	app := &App{
		Home:     home,
		Config:   &config.Config{GodotAPI: srv.URL},
		Platform: platform.Info{OS: "linux", Arch: "amd64"},
	}

	cmd := newUpdateCmd(app)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	var runErr error
	output := captureStderr(t, func() {
		runErr = cmd.Execute()
	})
	if runErr != nil {
		t.Fatalf("expected nil error, got: %v", runErr)
	}
	if !strings.Contains(output, "Metadata updated (1 releases)") {
		t.Errorf("expected release-count confirmation on stderr, got:\n%s", output)
	}
}

// TestUpdate_FetchReleasesErrorPropagates proves that when the metadata API
// returns a non-200 status, the command returns a non-nil error instead of
// silently succeeding.
func TestUpdate_FetchReleasesErrorPropagates(t *testing.T) {
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

	cmd := newUpdateCmd(app)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	var runErr error
	_ = captureStderr(t, func() {
		runErr = cmd.Execute()
	})
	if runErr == nil {
		t.Fatal("expected error when metadata API returns a non-200 status, got nil")
	}
}
