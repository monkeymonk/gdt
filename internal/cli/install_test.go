package cli

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/monkeymonk/gdt/internal/config"
	"github.com/monkeymonk/gdt/internal/engine"
	"github.com/monkeymonk/gdt/internal/platform"
)

// testInstallZipBytes builds a minimal valid zip archive containing a single
// "godot" binary file.
func testInstallZipBytes(t *testing.T) []byte {
	t.Helper()
	buf := &bytes.Buffer{}
	zw := zip.NewWriter(buf)
	w, err := zw.Create("godot")
	if err != nil {
		t.Fatalf("zip create: %v", err)
	}
	if _, err := w.Write([]byte("fake binary")); err != nil {
		t.Fatalf("zip write: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
	return buf.Bytes()
}

// testInstallServer serves a fake GitHub releases API for a single stable
// release plus a matching zip asset, returning the releases API URL.
func testInstallServer(t *testing.T, version string) (apiURL string) {
	t.Helper()
	zipData := testInstallZipBytes(t)
	artifactName := fmt.Sprintf("Godot_v%s-stable_linux.x86_64.zip", version)

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	mux.HandleFunc("/releases", func(w http.ResponseWriter, r *http.Request) {
		releases := []map[string]any{
			{
				"tag_name": version + "-stable",
				"assets": []map[string]string{
					{"name": artifactName, "browser_download_url": srv.URL + "/assets/" + artifactName},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(releases)
	})
	mux.HandleFunc("/assets/"+artifactName, func(w http.ResponseWriter, r *http.Request) {
		w.Write(zipData)
	})

	return srv.URL + "/releases"
}

func TestInstall_AfterHookFailure_ReturnsError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows: fake plugin binary is a #!/bin/sh script")
	}
	home := t.TempDir()
	apiURL := testInstallServer(t, "9.9.9")
	writeFailingV2HookPlugin(t, filepath.Join(home, "plugins"), "failplugin", "after_install")

	app := &App{
		Home:     home,
		Platform: platform.Info{OS: "linux", Arch: "amd64"},
		Config:   &config.Config{GodotAPI: apiURL},
	}

	cmd := newInstallCmd(app)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{"9.9.9"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when after_install hook fails, got nil")
	}

	var ae *engine.ActionableError
	if !errors.As(err, &ae) {
		t.Fatalf("expected *engine.ActionableError, got %T: %v", err, err)
	}
	if ae.Suggestion == "" {
		t.Error("expected non-empty Suggestion on ActionableError")
	}
	if !strings.Contains(err.Error(), "after_install") {
		t.Errorf("expected error to mention after_install hook, got: %v", err)
	}
	if !strings.Contains(err.Error(), "installed successfully") {
		t.Errorf("expected error to make clear the install itself succeeded, got: %v", err)
	}

	// The engine must have actually been installed before the hook ran.
	if _, statErr := os.Stat(filepath.Join(home, "versions", "9.9.9", "godot")); statErr != nil {
		t.Errorf("expected engine binary to exist despite hook failure: %v", statErr)
	}
}
