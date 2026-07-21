package download

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func createTestTarGz(t *testing.T, path string, files map[string][]byte, modes map[string]int64) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)

	for name, content := range files {
		mode := int64(0644)
		if m, ok := modes[name]; ok {
			mode = m
		}
		hdr := &tar.Header{
			Name: name,
			Mode: mode,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(content); err != nil {
			t.Fatal(err)
		}
	}

	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestExtractTarGz(t *testing.T) {
	dir := t.TempDir()
	tarGzPath := filepath.Join(dir, "test.tar.gz")
	destDir := filepath.Join(dir, "out")

	createTestTarGz(t, tarGzPath, map[string][]byte{
		"gdt": []byte("fake binary"),
	}, map[string]int64{
		"gdt": 0755,
	})

	if err := ExtractTarGz(tarGzPath, destDir); err != nil {
		t.Fatal(err)
	}

	extracted := filepath.Join(destDir, "gdt")
	data, err := os.ReadFile(extracted)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, []byte("fake binary")) {
		t.Error("extracted content mismatch")
	}

	if runtime.GOOS != "windows" {
		info, err := os.Stat(extracted)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm()&0755 != 0755 {
			t.Errorf("expected mode 0755 bits set, got %v", info.Mode().Perm())
		}
	}
}

func TestExtractTarGz_ZipSlip(t *testing.T) {
	dir := t.TempDir()
	tarGzPath := filepath.Join(dir, "evil.tar.gz")
	destDir := filepath.Join(dir, "out")

	createTestTarGz(t, tarGzPath, map[string][]byte{
		"../evil": []byte("malicious"),
	}, nil)

	if err := ExtractTarGz(tarGzPath, destDir); err == nil {
		t.Error("expected error for zip slip path, got nil")
	}
}
