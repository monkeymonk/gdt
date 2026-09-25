package engine

import (
	"archive/zip"
	"bytes"
	"context"
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
	"github.com/monkeymonk/gdt/internal/metadata"
	"github.com/monkeymonk/gdt/internal/platform"
)

func TestActionableError(t *testing.T) {
	base := errors.New("boom")
	ae := &ActionableError{Err: base, Suggestion: "do X"}

	if ae.Error() != "boom" {
		t.Errorf("expected Error() to be %q, got %q", "boom", ae.Error())
	}
	if ae.Unwrap() != base {
		t.Errorf("expected Unwrap() to return base error, got %v", ae.Unwrap())
	}
	if !errors.Is(ae, base) {
		t.Error("expected errors.Is(ae, base) to be true")
	}
}

func TestCacheDir(t *testing.T) {
	svc := NewService("/home/x", platform.Info{OS: "linux", Arch: "amd64"}, &config.Config{})
	want := filepath.Join("/home/x", "cache")
	if got := svc.CacheDir(); got != want {
		t.Errorf("CacheDir() = %q, want %q", got, want)
	}
}

func TestCachePath(t *testing.T) {
	svc := NewService("/home/x", platform.Info{OS: "linux", Arch: "amd64"}, &config.Config{})
	want := filepath.Join("/home/x", "cache", "releases.json")
	if got := svc.CachePath(); got != want {
		t.Errorf("CachePath() = %q, want %q", got, want)
	}
}

// testZipBytes builds a minimal valid zip archive containing a single file.
func testZipBytes(t *testing.T) []byte {
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

// testDownloadServer serves a fake GitHub releases API for a single stable
// release plus a matching zip asset, returning the releases API URL.
func testDownloadServer(t *testing.T, version string) (apiURL string) {
	t.Helper()
	zipData := testZipBytes(t)
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

// testDownloadSpec builds a downloadSpec that resolves and downloads a
// single fake "1.0.0" release from a test server, skipping checksum
// verification.
func testDownloadSpec(svc *Service, apiURL string) downloadSpec {
	return downloadSpec{
		CachePath: svc.CachePath(),
		APIURL:    apiURL,
		Query:     "1.0.0",
		DestDir:   svc.VersionsDir(),
		ResolveArtifact: func(release *metadata.Release, plat platform.Info, mono bool) (string, error) {
			return metadata.ResolveEngineArtifact(release, plat, mono)
		},
		VerifyChecksum: false,
	}
}

func TestDownloadAndInstall_TmpDirCreateFails(t *testing.T) {
	svc := testService(t)
	apiURL := testDownloadServer(t, "1.0.0")

	// Pre-create a regular file where the extraction tmp dir needs to be a directory.
	if err := os.MkdirAll(svc.CacheDir(), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	tmpDirPath := filepath.Join(svc.CacheDir(), "tmp")
	if err := os.WriteFile(tmpDirPath, []byte("blocker"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, err := svc.downloadAndInstall(context.Background(), testDownloadSpec(svc, apiURL))
	if err == nil {
		t.Fatal("expected error")
	}
	var ae *ActionableError
	if !errors.As(err, &ae) {
		t.Fatalf("expected *ActionableError, got %T: %v", err, err)
	}
	if ae.Suggestion == "" {
		t.Error("expected non-empty suggestion")
	}
	if !strings.Contains(ae.Err.Error(), tmpDirPath) {
		t.Errorf("expected error to mention %s, got: %v", tmpDirPath, ae.Err)
	}
}

func TestDownloadAndInstall_DestParentCreateFails(t *testing.T) {
	svc := testService(t)
	apiURL := testDownloadServer(t, "1.0.0")

	// Replace the versions dir with a file so creating it as a parent directory fails.
	if err := os.RemoveAll(svc.VersionsDir()); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(svc.VersionsDir(), []byte("blocker"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, err := svc.downloadAndInstall(context.Background(), testDownloadSpec(svc, apiURL))
	if err == nil {
		t.Fatal("expected error")
	}
	var ae *ActionableError
	if !errors.As(err, &ae) {
		t.Fatalf("expected *ActionableError, got %T: %v", err, err)
	}
	if !strings.Contains(ae.Err.Error(), svc.VersionsDir()) {
		t.Errorf("expected error to mention %s, got: %v", svc.VersionsDir(), ae.Err)
	}
}

func TestDownloadAndInstall_RemoveExistingFails(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission-based removal failures aren't reliable on windows")
	}
	svc := testService(t)
	apiURL := testDownloadServer(t, "1.0.0")

	destDir := filepath.Join(svc.VersionsDir(), "1.0.0")
	blockedDir := filepath.Join(destDir, "locked")
	if err := os.MkdirAll(blockedDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(blockedDir, "file.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	// Read+execute but no write: os.RemoveAll can list entries but cannot unlink them.
	if err := os.Chmod(blockedDir, 0o555); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Cleanup(func() { os.Chmod(blockedDir, 0o755) })

	spec := testDownloadSpec(svc, apiURL)
	spec.Force = true
	_, err := svc.downloadAndInstall(context.Background(), spec)
	if err == nil {
		t.Fatal("expected error")
	}
	var ae *ActionableError
	if !errors.As(err, &ae) {
		t.Fatalf("expected *ActionableError, got %T: %v", err, err)
	}
	if !strings.Contains(ae.Suggestion, destDir) {
		t.Errorf("expected suggestion to mention %s, got: %s", destDir, ae.Suggestion)
	}
}

// testChecksumMismatchDownloadServer serves a fake GitHub releases API for a
// single stable release plus a matching zip asset AND a SHA512-SUMS.txt
// checksum file whose entry for the artifact is deliberately wrong, so
// downloadAndInstall's checksum-verification step (spec.VerifyChecksum)
// always fails for it.
func testChecksumMismatchDownloadServer(t *testing.T, version string) (apiURL string) {
	t.Helper()
	zipData := testZipBytes(t)
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
					{"name": "SHA512-SUMS.txt", "browser_download_url": srv.URL + "/assets/SHA512-SUMS.txt"},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(releases)
	})
	mux.HandleFunc("/assets/"+artifactName, func(w http.ResponseWriter, r *http.Request) {
		w.Write(zipData)
	})
	mux.HandleFunc("/assets/SHA512-SUMS.txt", func(w http.ResponseWriter, r *http.Request) {
		// Deliberately wrong checksum for artifactName.
		fmt.Fprintf(w, "0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000  %s\n", artifactName)
	})

	return srv.URL + "/releases"
}

// TestDownloadAndInstall_ChecksumMismatch proves the checksum-verification
// path (spec.VerifyChecksum) is reached and correctly returns
// ErrChecksumMismatch, and that the new `if rmErr := os.Remove(archivePath);
// rmErr != nil` error-checking added around the pre-existing cleanup call
// didn't change behavior for the normal case where cleanup succeeds.
//
// The adjacent branch — cleanup (os.Remove) *also* failing, which wraps
// both errors into an *ActionableError — is not exercised here: simulating
// a directory-permission failure that blocks removal but not the
// preceding download write requires either root (to chattr/fake a
// different file owner) or a mid-call hook this function has none of;
// this repo has no privilege escalation available (sandboxed, no sudo).
// See the task's Report for this accepted, documented gap.
func TestDownloadAndInstall_ChecksumMismatch(t *testing.T) {
	svc := testService(t)
	apiURL := testChecksumMismatchDownloadServer(t, "1.0.0")

	spec := testDownloadSpec(svc, apiURL)
	spec.VerifyChecksum = true

	_, err := svc.downloadAndInstall(context.Background(), spec)
	if !errors.Is(err, ErrChecksumMismatch) {
		t.Fatalf("expected ErrChecksumMismatch, got %T: %v", err, err)
	}

	// The archive must have been removed (cleanup succeeded) — proving the
	// new `if rmErr != nil` guard's else-path (return ErrChecksumMismatch
	// directly) still behaves like the pre-existing unconditional-remove
	// code for the normal, cleanup-succeeds case.
	archivePath := filepath.Join(svc.CacheDir(), "downloads", "Godot_v1.0.0-stable_linux.x86_64.zip")
	if _, statErr := os.Stat(archivePath); !os.IsNotExist(statErr) {
		t.Errorf("expected the mismatched archive to have been removed, stat error: %v", statErr)
	}
}
