package selfupdate

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type releaseAsset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

type releaseFixture struct {
	TagName string         `json:"tag_name"`
	Assets  []releaseAsset `json:"assets"`
}

func buildTarGz(t *testing.T, name string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0755, Size: int64(len(content))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func buildZip(t *testing.T, name string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	fw, err := zw.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fw.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func newReleaseServer(t *testing.T, apiStatus int, fixture *releaseFixture, assetPath string, assetBytes []byte) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/release", func(w http.ResponseWriter, r *http.Request) {
		if apiStatus != http.StatusOK {
			w.WriteHeader(apiStatus)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(fixture)
	})
	if assetPath != "" {
		mux.HandleFunc(assetPath, func(w http.ResponseWriter, r *http.Request) {
			w.Write(assetBytes)
		})
	}
	return httptest.NewServer(mux)
}

func newReleaseServerMulti(t *testing.T, fixture *releaseFixture, assets map[string][]byte) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/release", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(fixture)
	})
	for path, data := range assets {
		data := data
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			w.Write(data)
		})
	}
	return httptest.NewServer(mux)
}

func TestUpdate_AlreadyUpToDate(t *testing.T) {
	fixture := &releaseFixture{TagName: "v1.0.0"}
	srv := newReleaseServer(t, http.StatusOK, fixture, "", nil)
	defer srv.Close()

	res, err := Update(context.Background(), "1.0.0", srv.URL+"/release")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Updated {
		t.Error("expected Updated=false")
	}
}

func TestUpdate_FetchError(t *testing.T) {
	fixture := &releaseFixture{TagName: "v1.0.0"}
	srv := newReleaseServer(t, http.StatusInternalServerError, fixture, "", nil)
	defer srv.Close()

	_, err := Update(context.Background(), "0.9.0", srv.URL+"/release")
	if err == nil {
		t.Fatal("expected error for failed fetch")
	}
}

func TestUpdate_NoAssetForPlatform(t *testing.T) {
	fixture := &releaseFixture{
		TagName: "v2.0.0",
		Assets: []releaseAsset{
			{Name: "gdt-2.0.0-plan9-mips.tar.gz", URL: "http://example.invalid/gdt-2.0.0-plan9-mips.tar.gz"},
		},
	}
	srv := newReleaseServer(t, http.StatusOK, fixture, "", nil)
	defer srv.Close()

	_, err := Update(context.Background(), "1.0.0", srv.URL+"/release")
	if err == nil {
		t.Fatal("expected error for missing platform asset")
	}
	if !strings.Contains(err.Error(), "no binary found") {
		t.Errorf("expected error to mention %q, got %q", "no binary found", err.Error())
	}
}

func TestUpdate_HappyPath(t *testing.T) {
	newBinContent := []byte("new gdt binary bytes")

	ext := ".tar.gz"
	binName := "gdt"
	if runtime.GOOS == "windows" {
		ext = ".zip"
		binName = "gdt.exe"
	}

	assetName := "gdt-2.0.0-" + runtime.GOOS + "-" + runtime.GOARCH + ext
	assetPath := "/assets/" + assetName

	var archiveBytes []byte
	if ext == ".zip" {
		archiveBytes = buildZip(t, binName, newBinContent)
	} else {
		archiveBytes = buildTarGz(t, binName, newBinContent)
	}

	fixture := &releaseFixture{TagName: "v2.0.0"}
	srv := newReleaseServer(t, http.StatusOK, fixture, assetPath, archiveBytes)
	defer srv.Close()

	fixture.Assets = []releaseAsset{
		{Name: assetName, URL: srv.URL + assetPath},
	}

	dir := t.TempDir()
	exePath := filepath.Join(dir, "gdt-current")
	if err := os.WriteFile(exePath, []byte("old placeholder binary"), 0755); err != nil {
		t.Fatal(err)
	}

	origExecutable := osExecutable
	osExecutable = func() (string, error) { return exePath, nil }
	defer func() { osExecutable = origExecutable }()

	res, err := Update(context.Background(), "1.0.0", srv.URL+"/release")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Updated {
		t.Fatal("expected Updated=true")
	}
	if res.NewVersion != "v2.0.0" {
		t.Errorf("got NewVersion %q, want %q", res.NewVersion, "v2.0.0")
	}

	data, err := os.ReadFile(exePath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, newBinContent) {
		t.Errorf("exe content mismatch: got %q, want %q", data, newBinContent)
	}
}

func TestUpdate_ChecksumValid(t *testing.T) {
	newBinContent := []byte("new gdt binary bytes")

	ext := ".tar.gz"
	binName := "gdt"
	if runtime.GOOS == "windows" {
		ext = ".zip"
		binName = "gdt.exe"
	}

	assetName := "gdt-2.0.0-" + runtime.GOOS + "-" + runtime.GOARCH + ext
	assetPath := "/assets/" + assetName

	var archiveBytes []byte
	if ext == ".zip" {
		archiveBytes = buildZip(t, binName, newBinContent)
	} else {
		archiveBytes = buildTarGz(t, binName, newBinContent)
	}

	sum := sha256.Sum256(archiveBytes)
	checksumsContent := []byte(hex.EncodeToString(sum[:]) + "  " + assetName + "\n")

	fixture := &releaseFixture{TagName: "v2.0.0"}
	assets := map[string][]byte{
		assetPath:        archiveBytes,
		"/checksums.txt": checksumsContent,
	}
	srv := newReleaseServerMulti(t, fixture, assets)
	defer srv.Close()

	fixture.Assets = []releaseAsset{
		{Name: assetName, URL: srv.URL + assetPath},
		{Name: "checksums.txt", URL: srv.URL + "/checksums.txt"},
	}

	dir := t.TempDir()
	exePath := filepath.Join(dir, "gdt-current")
	if err := os.WriteFile(exePath, []byte("old placeholder binary"), 0755); err != nil {
		t.Fatal(err)
	}

	origExecutable := osExecutable
	osExecutable = func() (string, error) { return exePath, nil }
	defer func() { osExecutable = origExecutable }()

	res, err := Update(context.Background(), "1.0.0", srv.URL+"/release")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Updated {
		t.Fatal("expected Updated=true")
	}

	data, err := os.ReadFile(exePath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, newBinContent) {
		t.Errorf("exe content mismatch: got %q, want %q", data, newBinContent)
	}
}

func TestUpdate_ChecksumMismatch(t *testing.T) {
	newBinContent := []byte("new gdt binary bytes")
	placeholder := []byte("old placeholder binary")

	ext := ".tar.gz"
	binName := "gdt"
	if runtime.GOOS == "windows" {
		ext = ".zip"
		binName = "gdt.exe"
	}

	assetName := "gdt-2.0.0-" + runtime.GOOS + "-" + runtime.GOARCH + ext
	assetPath := "/assets/" + assetName

	var archiveBytes []byte
	if ext == ".zip" {
		archiveBytes = buildZip(t, binName, newBinContent)
	} else {
		archiveBytes = buildTarGz(t, binName, newBinContent)
	}

	wrongHash := strings.Repeat("0", 64)
	checksumsContent := []byte(wrongHash + "  " + assetName + "\n")

	fixture := &releaseFixture{TagName: "v2.0.0"}
	assets := map[string][]byte{
		assetPath:        archiveBytes,
		"/checksums.txt": checksumsContent,
	}
	srv := newReleaseServerMulti(t, fixture, assets)
	defer srv.Close()

	fixture.Assets = []releaseAsset{
		{Name: assetName, URL: srv.URL + assetPath},
		{Name: "checksums.txt", URL: srv.URL + "/checksums.txt"},
	}

	dir := t.TempDir()
	exePath := filepath.Join(dir, "gdt-current")
	if err := os.WriteFile(exePath, placeholder, 0755); err != nil {
		t.Fatal(err)
	}

	origExecutable := osExecutable
	osExecutable = func() (string, error) { return exePath, nil }
	defer func() { osExecutable = origExecutable }()

	_, err := Update(context.Background(), "1.0.0", srv.URL+"/release")
	if err == nil {
		t.Fatal("expected checksum mismatch error")
	}
	if !strings.Contains(err.Error(), "checksum mismatch") {
		t.Errorf("expected error to mention %q, got %q", "checksum mismatch", err.Error())
	}

	data, err := os.ReadFile(exePath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, placeholder) {
		t.Errorf("exe should not have been modified: got %q", data)
	}
}
