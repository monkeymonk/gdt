package download

import (
	"archive/zip"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadFile(t *testing.T) {
	content := []byte("fake binary content")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.Write(content)
	}))
	defer srv.Close()

	dir := t.TempDir()
	dest := filepath.Join(dir, "test.zip")

	err := File(context.Background(), srv.URL+"/test.zip", dest, DownloadOpts{})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(content) {
		t.Error("downloaded content mismatch")
	}
}

func TestDownloadFileResume(t *testing.T) {
	content := []byte("0123456789abcdefghij") // 20 bytes
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rangeHeader := r.Header.Get("Range")
		if rangeHeader != "" {
			var start int64
			fmt.Sscanf(rangeHeader, "bytes=%d-", &start)
			if start >= int64(len(content)) {
				w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
				return
			}
			w.Header().Set("Content-Length", fmt.Sprintf("%d", int64(len(content))-start))
			w.WriteHeader(http.StatusPartialContent)
			w.Write(content[start:])
			return
		}
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.Write(content)
	}))
	defer srv.Close()

	dir := t.TempDir()
	dest := filepath.Join(dir, "resumed.bin")
	partialPath := dest + ".partial"

	// Write first 10 bytes as partial file.
	if err := os.WriteFile(partialPath, content[:10], 0644); err != nil {
		t.Fatal(err)
	}

	err := File(context.Background(), srv.URL+"/file.bin", dest, DownloadOpts{Resume: true})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(content) {
		t.Errorf("content mismatch: got %q, want %q", data, content)
	}

	// Partial file should be gone after rename.
	if _, err := os.Stat(partialPath); !os.IsNotExist(err) {
		t.Error("partial file should have been removed")
	}
}

func TestDownloadFileResumeNoPartial(t *testing.T) {
	content := []byte("full download content")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.Write(content)
	}))
	defer srv.Close()

	dir := t.TempDir()
	dest := filepath.Join(dir, "fresh.bin")

	// Resume=true but no partial file exists — should do full download.
	err := File(context.Background(), srv.URL+"/file.bin", dest, DownloadOpts{Resume: true})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(content) {
		t.Errorf("content mismatch: got %q, want %q", data, content)
	}
}

func TestVerifyChecksum(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.bin")
	content := []byte("test content")
	os.WriteFile(path, content, 0644)

	h := sha512.Sum512(content)
	checksum := hex.EncodeToString(h[:])

	err := VerifyChecksum(path, checksum)
	if err != nil {
		t.Fatal(err)
	}
}

func TestVerifyChecksumMismatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.bin")
	os.WriteFile(path, []byte("content"), 0644)

	err := VerifyChecksum(path, "0000000000000000")
	if err == nil {
		t.Error("should error on checksum mismatch")
	}
}

func TestExtractZip(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "test.zip")
	destDir := filepath.Join(dir, "out")

	createTestZip(t, zipPath, map[string][]byte{
		"godot": []byte("fake binary"),
	})

	err := ExtractZip(zipPath, destDir)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(destDir, "godot"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "fake binary" {
		t.Error("extracted content mismatch")
	}
}

func createTestZip(t *testing.T, path string, files map[string][]byte) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	for name, content := range files {
		fw, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		fw.Write(content)
	}
	w.Close()
}
